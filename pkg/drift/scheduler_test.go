package drift

import (
	"context"
	"testing"
	"time"

	"github.com/ataiva-software/forge/pkg/core"
	"github.com/ataiva-software/forge/pkg/types"
)

func TestDriftScheduler_New(t *testing.T) {
	registry := types.NewProviderRegistry()
	planner := core.NewPlanner(registry)
	
	scheduler := NewDriftScheduler(planner, registry)
	
	if scheduler == nil {
		t.Fatal("Expected non-nil scheduler")
	}
	
	if !scheduler.IsEnabled() {
		t.Error("Expected scheduler to be enabled by default")
	}
	
	if scheduler.IsRunning() {
		t.Error("Expected scheduler to not be running initially")
	}
}

func TestDriftScheduler_AddModule(t *testing.T) {
	registry := types.NewProviderRegistry()
	planner := core.NewPlanner(registry)
	scheduler := NewDriftScheduler(planner, registry)
	
	module := &core.Module{
		APIVersion: "ataiva.com/chisel/v1",
		Kind:       "Module",
		Metadata: core.ModuleMetadata{
			Name:    "test-module",
			Version: "1.0.0",
		},
		Spec: core.ModuleSpec{
			Resources: []types.Resource{
				{
					Type:  "file",
					Name:  "test-file",
					State: types.StatePresent,
				},
			},
		},
	}
	
	config := &DriftScheduleConfig{
		Interval:    5 * time.Minute,
		Enabled:     true,
		MaxRetries:  3,
		RetryDelay:  30 * time.Second,
	}
	
	err := scheduler.AddModule(module, config)
	if err != nil {
		t.Fatalf("Failed to add module: %v", err)
	}
	
	modules := scheduler.GetScheduledModules()
	if len(modules) != 1 {
		t.Errorf("Expected 1 scheduled module, got %d", len(modules))
	}
	
	if modules[0].Module.Metadata.Name != "test-module" {
		t.Errorf("Expected module name 'test-module', got '%s'", modules[0].Module.Metadata.Name)
	}
}

func TestDriftScheduler_AddModuleErrors(t *testing.T) {
	registry := types.NewProviderRegistry()
	planner := core.NewPlanner(registry)
	scheduler := NewDriftScheduler(planner, registry)
	
	// Test nil module
	err := scheduler.AddModule(nil, &DriftScheduleConfig{})
	if err == nil {
		t.Error("Expected error for nil module")
	}
	
	// Test nil config
	module := &core.Module{
		Metadata: core.ModuleMetadata{Name: "test"},
	}
	err = scheduler.AddModule(module, nil)
	if err == nil {
		t.Error("Expected error for nil config")
	}
	
	// Test invalid config
	invalidConfig := &DriftScheduleConfig{
		Interval: -1 * time.Minute, // Invalid negative interval
	}
	err = scheduler.AddModule(module, invalidConfig)
	if err == nil {
		t.Error("Expected error for invalid config")
	}
}

func TestDriftScheduler_RemoveModule(t *testing.T) {
	registry := types.NewProviderRegistry()
	planner := core.NewPlanner(registry)
	scheduler := NewDriftScheduler(planner, registry)
	
	module := &core.Module{
		Metadata: core.ModuleMetadata{Name: "test-module"},
		Spec:     core.ModuleSpec{Resources: []types.Resource{}},
	}
	
	config := &DriftScheduleConfig{
		Interval: 5 * time.Minute,
		Enabled:  true,
	}
	
	// Add module
	scheduler.AddModule(module, config)
	
	// Remove module
	err := scheduler.RemoveModule("test-module")
	if err != nil {
		t.Fatalf("Failed to remove module: %v", err)
	}
	
	modules := scheduler.GetScheduledModules()
	if len(modules) != 0 {
		t.Errorf("Expected 0 scheduled modules after removal, got %d", len(modules))
	}
	
	// Try to remove non-existent module
	err = scheduler.RemoveModule("non-existent")
	if err == nil {
		t.Error("Expected error when removing non-existent module")
	}
}

func TestDriftScheduler_StartStop(t *testing.T) {
	registry := types.NewProviderRegistry()
	planner := core.NewPlanner(registry)
	scheduler := NewDriftScheduler(planner, registry)
	
	// Test start
	ctx := context.Background()
	err := scheduler.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}
	
	if !scheduler.IsRunning() {
		t.Error("Expected scheduler to be running after start")
	}
	
	// Test double start
	err = scheduler.Start(ctx)
	if err == nil {
		t.Error("Expected error when starting already running scheduler")
	}
	
	// Test stop
	err = scheduler.Stop()
	if err != nil {
		t.Fatalf("Failed to stop scheduler: %v", err)
	}
	
	if scheduler.IsRunning() {
		t.Error("Expected scheduler to not be running after stop")
	}
	
	// Test double stop
	err = scheduler.Stop()
	if err != nil {
		t.Fatalf("Unexpected error when stopping already stopped scheduler: %v", err)
	}
}

func TestDriftScheduler_EnableDisable(t *testing.T) {
	registry := types.NewProviderRegistry()
	planner := core.NewPlanner(registry)
	scheduler := NewDriftScheduler(planner, registry)
	
	// Test initial state
	if !scheduler.IsEnabled() {
		t.Error("Expected scheduler to be enabled by default")
	}
	
	// Test disable
	scheduler.Disable()
	if scheduler.IsEnabled() {
		t.Error("Expected scheduler to be disabled after Disable()")
	}
	
	// Test enable
	scheduler.Enable()
	if !scheduler.IsEnabled() {
		t.Error("Expected scheduler to be enabled after Enable()")
	}
}

func TestDriftScheduler_ScheduleExecution(t *testing.T) {
	registry := types.NewProviderRegistry()
	planner := core.NewPlanner(registry)
	scheduler := NewDriftScheduler(planner, registry)
	
	// Add a mock provider for testing
	mockProvider := &MockDriftProvider{
		resourceType: "test",
		readResult:   map[string]interface{}{"state": "present"},
		diffResult: &types.ResourceDiff{
			ResourceID: "test.resource",
			Action:     types.ActionNoop,
			Changes:    map[string]interface{}{},
		},
	}
	registry.Register(mockProvider)
	
	module := &core.Module{
		APIVersion: "ataiva.com/chisel/v1",
		Kind:       "Module",
		Metadata: core.ModuleMetadata{
			Name:    "test-module",
			Version: "1.0.0",
		},
		Spec: core.ModuleSpec{
			Resources: []types.Resource{
				{
					Type:  "test",
					Name:  "resource",
					State: types.StatePresent,
				},
			},
		},
	}
	
	config := &DriftScheduleConfig{
		Interval:   50 * time.Millisecond, // Very short interval for testing
		Enabled:    true,
		MaxRetries: 1,
		RetryDelay: 10 * time.Millisecond,
	}
	
	scheduler.AddModule(module, config)
	
	// Set fast check interval for testing
	scheduler.SetCheckInterval(25 * time.Millisecond)
	
	// Start scheduler
	ctx := context.Background()
	err := scheduler.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}
	defer scheduler.Stop()
	
	// Wait for scheduler to run and process
	time.Sleep(200 * time.Millisecond)
	
	// Check that we have scheduled modules
	modules := scheduler.GetScheduledModules()
	if len(modules) == 0 {
		t.Fatal("Expected scheduled modules")
	}
	
	// Check that the module has been run at least once
	if modules[0].RunCount == 0 {
		t.Error("Expected module to have been run at least once")
	}
	
	// The reports might be empty if drift detection completes but doesn't find drift
	// This is actually correct behavior, so we'll just check that the module ran
}

func TestDriftScheduleConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *DriftScheduleConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &DriftScheduleConfig{
				Interval:   5 * time.Minute,
				Enabled:    true,
				MaxRetries: 3,
				RetryDelay: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "zero interval (will be set by defaults)",
			config: &DriftScheduleConfig{
				Interval: 0,
			},
			wantErr: false,
		},
		{
			name: "negative interval",
			config: &DriftScheduleConfig{
				Interval: -1 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "negative max retries",
			config: &DriftScheduleConfig{
				Interval:   5 * time.Minute,
				MaxRetries: -1,
			},
			wantErr: true,
		},
		{
			name: "negative retry delay",
			config: &DriftScheduleConfig{
				Interval:   5 * time.Minute,
				MaxRetries: 3,
				RetryDelay: -1 * time.Second,
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("Expected validation error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

// MockDriftProvider for testing drift detection
type MockDriftProvider struct {
	resourceType string
	readResult   map[string]interface{}
	readError    error
	diffResult   *types.ResourceDiff
	diffError    error
}

func (m *MockDriftProvider) Type() string {
	return m.resourceType
}

func (m *MockDriftProvider) Validate(resource *types.Resource) error {
	return nil
}

func (m *MockDriftProvider) Read(ctx context.Context, resource *types.Resource) (map[string]interface{}, error) {
	if m.readError != nil {
		return nil, m.readError
	}
	return m.readResult, nil
}

func (m *MockDriftProvider) Diff(ctx context.Context, resource *types.Resource, current map[string]interface{}) (*types.ResourceDiff, error) {
	if m.diffError != nil {
		return nil, m.diffError
	}
	return m.diffResult, nil
}

func (m *MockDriftProvider) Apply(ctx context.Context, resource *types.Resource, diff *types.ResourceDiff) error {
	return nil
}
