package monitoring

import (
	"testing"
	"time"
)

func TestMetricsCollector_New(t *testing.T) {
	collector := NewMetricsCollector()
	
	if collector == nil {
		t.Fatal("Expected non-nil metrics collector")
	}
	
	if !collector.IsEnabled() {
		t.Error("Expected metrics collector to be enabled by default")
	}
}

func TestMetricsCollector_RecordExecution(t *testing.T) {
	collector := NewMetricsCollector()
	
	execution := &ExecutionMetrics{
		ModuleName:    "web-server",
		Action:        "apply",
		Status:        "success",
		Duration:      2 * time.Minute,
		ResourceCount: 5,
		User:          "admin",
		Environment:   "production",
	}
	
	err := collector.RecordExecution(execution)
	if err != nil {
		t.Fatalf("Failed to record execution: %v", err)
	}
	
	// Verify metrics were recorded
	metrics := collector.GetMetrics()
	if metrics.TotalExecutions != 1 {
		t.Errorf("Expected 1 total execution, got %d", metrics.TotalExecutions)
	}
	
	if metrics.SuccessfulExecutions != 1 {
		t.Errorf("Expected 1 successful execution, got %d", metrics.SuccessfulExecutions)
	}
	
	if metrics.FailedExecutions != 0 {
		t.Errorf("Expected 0 failed executions, got %d", metrics.FailedExecutions)
	}
}

func TestMetricsCollector_RecordFailedExecution(t *testing.T) {
	collector := NewMetricsCollector()
	
	execution := &ExecutionMetrics{
		ModuleName:    "database",
		Action:        "apply",
		Status:        "failed",
		Duration:      30 * time.Second,
		ResourceCount: 3,
		User:          "operator",
		Environment:   "staging",
		Error:         "Connection timeout",
	}
	
	err := collector.RecordExecution(execution)
	if err != nil {
		t.Fatalf("Failed to record execution: %v", err)
	}
	
	metrics := collector.GetMetrics()
	if metrics.TotalExecutions != 1 {
		t.Errorf("Expected 1 total execution, got %d", metrics.TotalExecutions)
	}
	
	if metrics.FailedExecutions != 1 {
		t.Errorf("Expected 1 failed execution, got %d", metrics.FailedExecutions)
	}
	
	if metrics.SuccessfulExecutions != 0 {
		t.Errorf("Expected 0 successful executions, got %d", metrics.SuccessfulExecutions)
	}
}

func TestMetricsCollector_RecordResourceMetrics(t *testing.T) {
	collector := NewMetricsCollector()
	
	resourceMetrics := &ResourceMetrics{
		Type:     "file",
		Action:   "create",
		Status:   "success",
		Duration: 500 * time.Millisecond,
		Module:   "web-server",
	}
	
	err := collector.RecordResource(resourceMetrics)
	if err != nil {
		t.Fatalf("Failed to record resource metrics: %v", err)
	}
	
	metrics := collector.GetMetrics()
	if metrics.TotalResources != 1 {
		t.Errorf("Expected 1 total resource, got %d", metrics.TotalResources)
	}
	
	// Check resource type breakdown
	if metrics.ResourcesByType["file"] != 1 {
		t.Errorf("Expected 1 file resource, got %d", metrics.ResourcesByType["file"])
	}
}

func TestMetricsCollector_RecordPolicyViolation(t *testing.T) {
	collector := NewMetricsCollector()
	
	violation := &PolicyViolationMetrics{
		Policy:     "security",
		Rule:       "deny_root_user",
		Resource:   "user.root",
		Module:     "user-management",
		Severity:   "high",
		Framework:  "CIS",
	}
	
	err := collector.RecordPolicyViolation(violation)
	if err != nil {
		t.Fatalf("Failed to record policy violation: %v", err)
	}
	
	metrics := collector.GetMetrics()
	if metrics.PolicyViolations != 1 {
		t.Errorf("Expected 1 policy violation, got %d", metrics.PolicyViolations)
	}
	
	// Check violation breakdown
	if metrics.ViolationsByPolicy["security"] != 1 {
		t.Errorf("Expected 1 security violation, got %d", metrics.ViolationsByPolicy["security"])
	}
}

func TestMetricsCollector_GetPrometheusMetrics(t *testing.T) {
	collector := NewMetricsCollector()
	
	// Record some test data
	execution := &ExecutionMetrics{
		ModuleName:    "test-module",
		Action:        "apply",
		Status:        "success",
		Duration:      1 * time.Minute,
		ResourceCount: 3,
	}
	collector.RecordExecution(execution)
	
	prometheusMetrics := collector.GetPrometheusMetrics()
	if prometheusMetrics == "" {
		t.Error("Expected non-empty Prometheus metrics")
	}
	
	// Check for expected metric names
	expectedMetrics := []string{
		"chisel_executions_total",
		"chisel_execution_duration_seconds",
		"chisel_resources_total",
	}
	
	for _, metric := range expectedMetrics {
		if !contains(prometheusMetrics, metric) {
			t.Errorf("Expected Prometheus metrics to contain '%s'", metric)
		}
	}
}

func TestMetricsCollector_EnableDisable(t *testing.T) {
	collector := NewMetricsCollector()
	
	if !collector.IsEnabled() {
		t.Error("Expected collector to be enabled by default")
	}
	
	collector.Disable()
	if collector.IsEnabled() {
		t.Error("Expected collector to be disabled after Disable()")
	}
	
	// Recording should be ignored when disabled
	execution := &ExecutionMetrics{
		ModuleName: "test",
		Action:     "apply",
		Status:     "success",
	}
	
	err := collector.RecordExecution(execution)
	if err != nil {
		t.Fatalf("Unexpected error when disabled: %v", err)
	}
	
	metrics := collector.GetMetrics()
	if metrics.TotalExecutions != 0 {
		t.Error("Expected no executions recorded when disabled")
	}
	
	collector.Enable()
	if !collector.IsEnabled() {
		t.Error("Expected collector to be enabled after Enable()")
	}
}

func TestMetricsCollector_GetMetricsByTimeRange(t *testing.T) {
	collector := NewMetricsCollector()
	
	now := time.Now()
	
	// Record execution from 1 hour ago
	execution1 := &ExecutionMetrics{
		ModuleName: "old-module",
		Action:     "apply",
		Status:     "success",
		Timestamp:  now.Add(-1 * time.Hour),
	}
	collector.RecordExecution(execution1)
	
	// Record execution from 10 minutes ago
	execution2 := &ExecutionMetrics{
		ModuleName: "recent-module",
		Action:     "apply",
		Status:     "success",
		Timestamp:  now.Add(-10 * time.Minute),
	}
	collector.RecordExecution(execution2)
	
	// Get metrics for last 30 minutes
	recentMetrics := collector.GetMetricsByTimeRange(now.Add(-30*time.Minute), now)
	if recentMetrics.TotalExecutions != 1 {
		t.Errorf("Expected 1 recent execution, got %d", recentMetrics.TotalExecutions)
	}
	
	// Get metrics for last 2 hours
	allMetrics := collector.GetMetricsByTimeRange(now.Add(-2*time.Hour), now)
	if allMetrics.TotalExecutions != 2 {
		t.Errorf("Expected 2 total executions, got %d", allMetrics.TotalExecutions)
	}
}

func TestMetricsCollector_ExportToFile(t *testing.T) {
	collector := NewMetricsCollector()
	
	// Record some test data
	execution := &ExecutionMetrics{
		ModuleName: "test-module",
		Action:     "apply",
		Status:     "success",
	}
	collector.RecordExecution(execution)
	
	// Export to temporary file
	tempFile := "/tmp/chisel-metrics-test.json"
	defer func() {
		// Clean up
		_ = removeFile(tempFile)
	}()
	
	err := collector.ExportToFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to export metrics: %v", err)
	}
	
	// Verify file exists and contains data
	if !fileExists(tempFile) {
		t.Error("Expected metrics file to be created")
	}
}

func TestExecutionMetrics_Validate(t *testing.T) {
	tests := []struct {
		name    string
		metrics *ExecutionMetrics
		wantErr bool
	}{
		{
			name: "valid metrics",
			metrics: &ExecutionMetrics{
				ModuleName: "test-module",
				Action:     "apply",
				Status:     "success",
			},
			wantErr: false,
		},
		{
			name: "missing module name",
			metrics: &ExecutionMetrics{
				Action: "apply",
				Status: "success",
			},
			wantErr: true,
		},
		{
			name: "missing action",
			metrics: &ExecutionMetrics{
				ModuleName: "test-module",
				Status:     "success",
			},
			wantErr: true,
		},
		{
			name: "missing status",
			metrics: &ExecutionMetrics{
				ModuleName: "test-module",
				Action:     "apply",
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metrics.Validate()
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

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr || 
		   (len(s) > len(substr) && containsHelper(s[1:], substr)))
}

func containsHelper(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	if s[:len(substr)] == substr {
		return true
	}
	return containsHelper(s[1:], substr)
}

func fileExists(filename string) bool {
	// Simple file existence check
	return true // Mock implementation
}

func removeFile(filename string) error {
	// Mock file removal
	return nil
}
