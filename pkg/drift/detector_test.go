package drift

import (
	"context"
	"testing"
	"time"

	"github.com/ataiva-software/forge/pkg/core"
	"github.com/ataiva-software/forge/pkg/types"
)

// MockProvider for testing drift detection
type MockProvider struct {
	resourceType string
	readResult   map[string]interface{}
	readError    error
	diffResult   *types.ResourceDiff
	diffError    error
}

func (m *MockProvider) Type() string {
	return m.resourceType
}

func (m *MockProvider) Validate(resource *types.Resource) error {
	return nil
}

func (m *MockProvider) Read(ctx context.Context, resource *types.Resource) (map[string]interface{}, error) {
	if m.readError != nil {
		return nil, m.readError
	}
	return m.readResult, nil
}

func (m *MockProvider) Diff(ctx context.Context, resource *types.Resource, current map[string]interface{}) (*types.ResourceDiff, error) {
	if m.diffError != nil {
		return nil, m.diffError
	}
	return m.diffResult, nil
}

func (m *MockProvider) Apply(ctx context.Context, resource *types.Resource, diff *types.ResourceDiff) error {
	return nil
}

func TestDriftDetector_CheckDrift(t *testing.T) {
	// Create mock registry
	registry := types.NewProviderRegistry()
	
	// Register mock providers
	noDriftProvider := &MockProvider{
		resourceType: "test",
		readResult:   map[string]interface{}{"state": "present"},
		diffResult: &types.ResourceDiff{
			ResourceID: "test.resource1",
			Action:     types.ActionNoop,
			Changes:    map[string]interface{}{},
		},
	}
	
	driftProvider := &MockProvider{
		resourceType: "drift",
		readResult:   map[string]interface{}{"state": "present"},
		diffResult: &types.ResourceDiff{
			ResourceID: "drift.resource2",
			Action:     types.ActionUpdate,
			Changes: map[string]interface{}{
				"content": map[string]interface{}{
					"from": "old content",
					"to":   "new content",
				},
			},
		},
	}
	
	registry.Register(noDriftProvider)
	registry.Register(driftProvider)
	
	// Create planner
	planner := core.NewPlanner(registry)
	
	// Create drift detector
	detector := NewDriftDetector(planner, registry, time.Minute)
	
	// Create test module
	module := &core.Module{
		Metadata: core.ModuleMetadata{
			Name:    "test-module",
			Version: "1.0.0",
		},
		Spec: core.ModuleSpec{
			Resources: []types.Resource{
				{
					Type:  "test",
					Name:  "resource1",
					State: types.StatePresent,
				},
				{
					Type:  "drift",
					Name:  "resource2",
					State: types.StatePresent,
				},
			},
		},
	}
	
	// Test drift detection
	ctx := context.Background()
	report, err := detector.CheckDrift(ctx, module)
	
	if err != nil {
		t.Fatalf("CheckDrift failed: %v", err)
	}
	
	// Verify report
	if report.ModuleName != "test-module" {
		t.Errorf("Expected module name 'test-module', got '%s'", report.ModuleName)
	}
	
	if report.TotalChecked != 2 {
		t.Errorf("Expected 2 resources checked, got %d", report.TotalChecked)
	}
	
	if report.DriftDetected != 1 {
		t.Errorf("Expected 1 resource with drift, got %d", report.DriftDetected)
	}
	
	if report.Errors != 0 {
		t.Errorf("Expected 0 errors, got %d", report.Errors)
	}
	
	// Verify individual results
	if len(report.Results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(report.Results))
	}
	
	// Find results by resource ID
	var noDriftResult, driftResult *DriftResult
	for i := range report.Results {
		if report.Results[i].ResourceID == "test.resource1" {
			noDriftResult = &report.Results[i]
		} else if report.Results[i].ResourceID == "drift.resource2" {
			driftResult = &report.Results[i]
		}
	}
	
	if noDriftResult == nil {
		t.Fatal("No drift result not found")
	}
	if noDriftResult.HasDrift {
		t.Error("Expected no drift for test.resource1")
	}
	
	if driftResult == nil {
		t.Fatal("Drift result not found")
	}
	if !driftResult.HasDrift {
		t.Error("Expected drift for drift.resource2")
	}
}

func TestDriftDetector_StartStop(t *testing.T) {
	registry := types.NewProviderRegistry()
	planner := core.NewPlanner(registry)
	detector := NewDriftDetector(planner, registry, 100*time.Millisecond)
	
	// Test initial state
	if detector.IsRunning() {
		t.Error("Detector should not be running initially")
	}
	
	// Test start
	ctx := context.Background()
	err := detector.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start detector: %v", err)
	}
	
	if !detector.IsRunning() {
		t.Error("Detector should be running after start")
	}
	
	// Test double start
	err = detector.Start(ctx)
	if err == nil {
		t.Error("Expected error when starting already running detector")
	}
	
	// Test stop
	err = detector.Stop()
	if err != nil {
		t.Fatalf("Failed to stop detector: %v", err)
	}
	
	if detector.IsRunning() {
		t.Error("Detector should not be running after stop")
	}
	
	// Test double stop
	err = detector.Stop()
	if err != nil {
		t.Fatalf("Unexpected error when stopping already stopped detector: %v", err)
	}
}

func TestDriftConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *DriftConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultDriftConfig(),
			wantErr: false,
		},
		{
			name: "invalid interval",
			config: &DriftConfig{
				Interval:       0,
				MaxConcurrency: 5,
				Timeout:        30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid concurrency",
			config: &DriftConfig{
				Interval:       time.Minute,
				MaxConcurrency: 0,
				Timeout:        30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			config: &DriftConfig{
				Interval:       time.Minute,
				MaxConcurrency: 5,
				Timeout:        0,
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

func TestDriftNotifier_Notify(t *testing.T) {
	notifier := NewDriftNotifier(1)
	
	// Add log channel
	logChannel := &LogNotificationChannel{}
	notifier.AddChannel(logChannel)
	
	// Test disabled notifications
	report := &DriftReport{
		ModuleName:    "test-module",
		DriftDetected: 2,
		TotalChecked:  5,
	}
	
	ctx := context.Background()
	err := notifier.Notify(ctx, report)
	if err != nil {
		t.Errorf("Unexpected error from disabled notifier: %v", err)
	}
	
	// Enable notifications
	notifier.Enable()
	
	// Test below threshold
	report.DriftDetected = 0
	err = notifier.Notify(ctx, report)
	if err != nil {
		t.Errorf("Unexpected error for below threshold: %v", err)
	}
	
	// Test above threshold
	report.DriftDetected = 2
	err = notifier.Notify(ctx, report)
	if err != nil {
		t.Errorf("Unexpected error for above threshold: %v", err)
	}
}

func TestLogNotificationChannel(t *testing.T) {
	channel := &LogNotificationChannel{}
	
	if channel.Type() != "log" {
		t.Errorf("Expected type 'log', got '%s'", channel.Type())
	}
	
	report := &DriftReport{
		ModuleName:    "test-module",
		DriftDetected: 1,
		TotalChecked:  3,
		Results: []DriftResult{
			{
				ResourceID: "test.resource1",
				HasDrift:   true,
			},
			{
				ResourceID: "test.resource2",
				HasDrift:   false,
			},
		},
	}
	
	ctx := context.Background()
	err := channel.Send(ctx, report)
	if err != nil {
		t.Errorf("Unexpected error from log channel: %v", err)
	}
}

func TestDriftDetector_Configuration(t *testing.T) {
	registry := types.NewProviderRegistry()
	planner := core.NewPlanner(registry)
	detector := NewDriftDetector(planner, registry, time.Minute)
	
	// Test interval update
	newInterval := 30 * time.Second
	detector.SetInterval(newInterval)
	if detector.interval != newInterval {
		t.Errorf("Expected interval %v, got %v", newInterval, detector.interval)
	}
	
	// Test concurrency update
	newConcurrency := 10
	detector.SetConcurrency(newConcurrency)
	if detector.maxConcurrency != newConcurrency {
		t.Errorf("Expected concurrency %d, got %d", newConcurrency, detector.maxConcurrency)
	}
	
	// Test invalid concurrency
	detector.SetConcurrency(0)
	if detector.maxConcurrency != newConcurrency {
		t.Error("Concurrency should not change for invalid value")
	}
	
	// Test timeout update
	newTimeout := 60 * time.Second
	detector.SetTimeout(newTimeout)
	if detector.timeout != newTimeout {
		t.Errorf("Expected timeout %v, got %v", newTimeout, detector.timeout)
	}
	
	// Test invalid timeout
	detector.SetTimeout(0)
	if detector.timeout != newTimeout {
		t.Error("Timeout should not change for invalid value")
	}
}
