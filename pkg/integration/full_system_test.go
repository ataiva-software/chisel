package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ataiva-software/chisel/pkg/core"
	"github.com/ataiva-software/chisel/pkg/drift"
	"github.com/ataiva-software/chisel/pkg/events"
	"github.com/ataiva-software/chisel/pkg/executor"
	"github.com/ataiva-software/chisel/pkg/inventory"
	"github.com/ataiva-software/chisel/pkg/notifications"
	"github.com/ataiva-software/chisel/pkg/providers"
	"github.com/ataiva-software/chisel/pkg/ssh"
	"github.com/ataiva-software/chisel/pkg/types"
)

// TestFullSystemIntegration tests the complete system working together
func TestFullSystemIntegration(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "chisel-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 1. Set up Event System
	eventBus := events.NewEventBus(100, 5)
	defer eventBus.Close()

	// Add event handlers
	logHandler := events.NewLogEventHandler("integration-log", []events.EventType{
		events.EventTypeResourceStarted,
		events.EventTypeResourceCompleted,
		events.EventTypeResourceFailed,
		events.EventTypePlanStarted,
		events.EventTypePlanCompleted,
		events.EventTypeApplyStarted,
		events.EventTypeApplyCompleted,
		events.EventTypeDriftDetected,
	}, false)
	eventBus.Subscribe(logHandler)

	metricsHandler := events.NewMetricsEventHandler("integration-metrics", []events.EventType{
		events.EventTypeResourceStarted,
		events.EventTypeResourceCompleted,
		events.EventTypeResourceFailed,
	})
	eventBus.Subscribe(metricsHandler)

	// 2. Set up Notification System
	notificationManager := notifications.NewNotificationManager(eventBus)

	// Add console notification channel
	consoleChannel := notifications.NewConsoleChannel("console", false, true)
	notificationManager.AddChannel(consoleChannel)

	// Add file notification channel
	notifLogPath := filepath.Join(tempDir, "notifications.log")
	fileChannel := notifications.NewFileChannel("file", notifLogPath, "json")
	notificationManager.AddChannel(fileChannel)

	// Add notification rules
	errorRule := notifications.NotificationRule{
		Name:     "error-notifications",
		Enabled:  true,
		Channels: []string{"console", "file"},
		Levels:   []notifications.NotificationLevel{notifications.LevelError, notifications.LevelCritical},
	}
	notificationManager.AddRule(errorRule)

	infoRule := notifications.NotificationRule{
		Name:     "info-notifications",
		Enabled:  true,
		Channels: []string{"file"},
		Levels:   []notifications.NotificationLevel{notifications.LevelInfo},
	}
	notificationManager.AddRule(infoRule)

	// 3. Set up SSH Connection (using mock for integration test)
	sshConfig := &ssh.ConnectionConfig{
		Host: "localhost",
		Port: 22,
		User: "test",
	}
	sshConfig.SetDefaults()

	mockConnection := ssh.NewMockExecutor()
	err = mockConnection.Connect(context.Background())
	if err != nil {
		t.Fatalf("Failed to connect mock SSH: %v", err)
	}

	// 4. Set up Provider Registry
	providerRegistry := types.NewProviderRegistry()

	// Register all providers
	providers := []types.Provider{
		providers.NewFileProvider(mockConnection),
		providers.NewPkgProvider(mockConnection),
		providers.NewServiceProvider(mockConnection),
		providers.NewUserProvider(mockConnection),
		providers.NewShellProvider(mockConnection),
	}

	for _, provider := range providers {
		err := providerRegistry.Register(provider)
		if err != nil {
			t.Fatalf("Failed to register provider %s: %v", provider.Type(), err)
		}
	}

	// 5. Set up Inventory System
	inventoryRegistry := inventory.NewInventoryRegistry()

	// Create static targets
	targets := []types.Target{
		{
			Host: "web1.example.com",
			Port: 22,
			User: "ubuntu",
			Labels: map[string]string{
				"role": "web",
				"env":  "prod",
			},
		},
		{
			Host: "web2.example.com",
			Port: 22,
			User: "ubuntu",
			Labels: map[string]string{
				"role": "web",
				"env":  "staging",
			},
		},
	}

	// 6. Create Test Module
	module := &core.Module{
		APIVersion: "ataiva.com/chisel/v1",
		Kind:       "Module",
		Metadata: core.ModuleMetadata{
			Name:    "integration-test",
			Version: "1.0.0",
		},
		Spec: core.ModuleSpec{
			Resources: []types.Resource{
				{
					Type:  "pkg",
					Name:  "git",
					State: types.StatePresent,
				},
				{
					Type:  "pkg",
					Name:  "vim",
					State: types.StatePresent,
				},
				{
					Type:  "user",
					Name:  "testuser",
					State: types.StatePresent,
				},
				{
					Type: "file",
					Name: "test-file",
					Properties: map[string]interface{}{
						"path":    filepath.Join(tempDir, "test.txt"),
						"content": "Hello from Chisel integration test!",
						"mode":    "0644",
					},
				},
				{
					Type: "shell",
					Name: "test-command",
					Properties: map[string]interface{}{
						"command": "echo 'Integration test successful'",
					},
				},
			},
		},
	}

	// Validate module
	err = module.Validate()
	if err != nil {
		t.Fatalf("Module validation failed: %v", err)
	}

	// 7. Set up Planner
	planner := core.NewPlanner(providerRegistry)

	// 8. Set up Parallel Executor with Events
	parallelExecutor := executor.NewEnhancedParallelExecutor(3, 30*time.Second, true)
	emitter := events.NewEventEmitter(eventBus, "integration-executor")

	// 9. Create Plan
	t.Log("Creating execution plan...")
	emitter.EmitPlanStarted(module.Metadata.Name, len(module.Spec.Resources))

	plan, err := planner.CreatePlan(module)
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}

	summary := plan.Summary()
	summaryMap := map[string]int{
		"create": summary.ToCreate,
		"update": summary.ToUpdate,
		"delete": summary.ToDelete,
	}
	emitter.EmitPlanCompleted(module.Metadata.Name, summaryMap, time.Second)

	t.Logf("Plan created: %d to create, %d to update, %d to delete",
		summary.ToCreate, summary.ToUpdate, summary.ToDelete)

	// 10. Create Execution Plan
	var executionPlan *executor.ExecutionPlan
	if len(plan.Changes) > 0 {
		var diffs []*types.ResourceDiff
		for _, change := range plan.Changes {
			if change.Error == nil {
				diffs = append(diffs, change.Diff)
			}
		}

		executionPlan, err = parallelExecutor.CreateExecutionPlan(diffs)
		if err != nil {
			t.Fatalf("Failed to create execution plan: %v", err)
		}

		t.Logf("Execution plan created with %d batches", len(executionPlan.Batches))
	}

	// 11. Execute Plan with Events and Rollback
	if executionPlan != nil && len(executionPlan.Batches) > 0 {
		t.Log("Executing plan...")
		emitter.EmitApplyStarted(module.Metadata.Name, len(module.Spec.Resources))

		// Create executor function that emits events
		executorFunc := func(ctx context.Context, diff *types.ResourceDiff) error {
			emitter.EmitResourceStarted(diff.ResourceID, diff.Action)
			start := time.Now()

			// Get provider and execute
			resourceType := getResourceType(diff.ResourceID)
			provider, err := providerRegistry.Get(resourceType)
			if err != nil {
				duration := time.Since(start)
				emitter.EmitResourceFailed(diff.ResourceID, diff.Action, err, duration)
				return err
			}

			// Find the resource
			var resource *types.Resource
			for _, res := range module.Spec.Resources {
				if res.ResourceID() == diff.ResourceID {
					resource = &res
					break
				}
			}

			if resource == nil {
				err := fmt.Errorf("resource not found: %s", diff.ResourceID)
				duration := time.Since(start)
				emitter.EmitResourceFailed(diff.ResourceID, diff.Action, err, duration)
				return err
			}

			// Apply the change
			err = provider.Apply(ctx, resource, diff)
			duration := time.Since(start)

			if err != nil {
				emitter.EmitResourceFailed(diff.ResourceID, diff.Action, err, duration)
				return err
			}

			emitter.EmitResourceCompleted(diff.ResourceID, diff.Action, duration)
			return nil
		}

		// Create reader function for rollback
		readerFunc := func(ctx context.Context, diff *types.ResourceDiff) (map[string]interface{}, error) {
			resourceType := getResourceType(diff.ResourceID)
			provider, err := providerRegistry.Get(resourceType)
			if err != nil {
				return nil, err
			}

			// Find the resource
			var resource *types.Resource
			for _, res := range module.Spec.Resources {
				if res.ResourceID() == diff.ResourceID {
					resource = &res
					break
				}
			}

			if resource == nil {
				return nil, fmt.Errorf("resource not found: %s", diff.ResourceID)
			}

			return provider.Read(ctx, resource)
		}

		ctx := context.Background()
		results, err := parallelExecutor.ExecuteWithRollback(ctx, executionPlan, executorFunc, readerFunc)

		if err != nil {
			// Create a manual event for apply failed since EmitApplyFailed doesn't exist
			event := events.NewEvent(events.EventTypeApplyFailed, "integration-executor", map[string]interface{}{
				"module_name": module.Metadata.Name,
				"error":       err.Error(),
			})
			eventBus.Publish(event)
			t.Logf("Execution failed (with rollback): %v", err)
		} else {
			emitter.EmitApplyCompleted(module.Metadata.Name, summaryMap, time.Since(time.Now()))
			t.Log("Execution completed successfully")
		}

		t.Logf("Execution results: %d operations completed", len(results))
	}

	// 12. Set up Drift Detection
	t.Log("Setting up drift detection...")
	driftDetector := drift.NewDriftDetector(planner, providerRegistry, 30*time.Second)

	// Perform drift check
	driftReport, err := driftDetector.CheckDrift(context.Background(), module)
	if err != nil {
		t.Fatalf("Drift detection failed: %v", err)
	}

	t.Logf("Drift detection completed: %d resources checked, %d with drift, %d errors",
		driftReport.TotalChecked, driftReport.DriftDetected, driftReport.Errors)

	// Emit drift events for resources with drift
	for _, result := range driftReport.Results {
		if result.HasDrift {
			emitter.EmitDriftDetected(module.Metadata.Name, result.ResourceID, result.Changes)
		}
	}

	// 13. Give time for async event processing
	time.Sleep(500 * time.Millisecond)

	// 14. Verify Event Metrics
	metrics := metricsHandler.GetMetrics()
	t.Logf("Event metrics: %+v", metrics)

	// Verify we got some events
	if metrics[string(events.EventTypeResourceStarted)] == 0 {
		t.Error("Expected resource started events")
	}

	if metrics[string(events.EventTypeResourceCompleted)] == 0 {
		t.Error("Expected resource completed events")
	}

	// 15. Verify Notifications were sent
	// Check that notification file was created
	if _, err := os.Stat(notifLogPath); os.IsNotExist(err) {
		t.Log("No notification file created (no notifications triggered)")
	} else {
		content, err := os.ReadFile(notifLogPath)
		if err != nil {
			t.Errorf("Failed to read notification file: %v", err)
		} else {
			t.Logf("Notifications written to file: %d bytes", len(content))
		}
	}

	// 16. Test Inventory Discovery
	t.Log("Testing inventory discovery...")
	
	// Create a mock dynamic inventory
	mockInventory := &MockDynamicInventory{
		inventoryType: "test",
		targets:       targets,
	}
	
	err = inventoryRegistry.Register(mockInventory)
	if err != nil {
		t.Fatalf("Failed to register mock inventory: %v", err)
	}
	
	// Test discovery
	discoveredTargets, err := inventoryRegistry.Discover(context.Background(), "test", "role=web")
	if err != nil {
		t.Fatalf("Failed to discover targets: %v", err)
	}
	
	t.Logf("Discovered %d targets with role=web", len(discoveredTargets))
	
	if len(discoveredTargets) != 2 {
		t.Errorf("Expected 2 discovered targets, got %d", len(discoveredTargets))
	}

	t.Log("Full system integration test completed successfully!")
}

// getResourceType extracts resource type from ResourceID
func getResourceType(resourceID string) string {
	for i, char := range resourceID {
		if char == '.' {
			return resourceID[:i]
		}
	}
	return resourceID
}

// MockDynamicInventory for testing
type MockDynamicInventory struct {
	inventoryType string
	targets       []types.Target
}

func (m *MockDynamicInventory) Type() string {
	return m.inventoryType
}

func (m *MockDynamicInventory) Validate() error {
	return nil
}

func (m *MockDynamicInventory) Discover(ctx context.Context, selector string) ([]types.Target, error) {
	selectorMap, err := inventory.ParseSelector(selector)
	if err != nil {
		return nil, err
	}
	
	var matchingTargets []types.Target
	for _, target := range m.targets {
		if inventory.MatchesSelector(target, selectorMap) {
			matchingTargets = append(matchingTargets, target)
		}
	}
	
	if matchingTargets == nil {
		return []types.Target{}, nil
	}
	
	return matchingTargets, nil
}
