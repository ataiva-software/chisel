package monitoring

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// ExecutionMetrics represents metrics for a module execution
type ExecutionMetrics struct {
	ModuleName    string        `json:"module_name"`
	Action        string        `json:"action"`
	Status        string        `json:"status"`
	Duration      time.Duration `json:"duration"`
	ResourceCount int           `json:"resource_count"`
	User          string        `json:"user"`
	Environment   string        `json:"environment"`
	Error         string        `json:"error,omitempty"`
	Timestamp     time.Time     `json:"timestamp"`
}

// Validate validates the execution metrics
func (e *ExecutionMetrics) Validate() error {
	if e.ModuleName == "" {
		return fmt.Errorf("module name is required")
	}
	
	if e.Action == "" {
		return fmt.Errorf("action is required")
	}
	
	if e.Status == "" {
		return fmt.Errorf("status is required")
	}
	
	return nil
}

// ResourceMetrics represents metrics for a resource operation
type ResourceMetrics struct {
	Type     string        `json:"type"`
	Action   string        `json:"action"`
	Status   string        `json:"status"`
	Duration time.Duration `json:"duration"`
	Module   string        `json:"module"`
	Error    string        `json:"error,omitempty"`
}

// PolicyViolationMetrics represents metrics for policy violations
type PolicyViolationMetrics struct {
	Policy    string `json:"policy"`
	Rule      string `json:"rule"`
	Resource  string `json:"resource"`
	Module    string `json:"module"`
	Severity  string `json:"severity"`
	Framework string `json:"framework"`
}

// AggregatedMetrics represents aggregated system metrics
type AggregatedMetrics struct {
	TotalExecutions      int                    `json:"total_executions"`
	SuccessfulExecutions int                    `json:"successful_executions"`
	FailedExecutions     int                    `json:"failed_executions"`
	TotalResources       int                    `json:"total_resources"`
	PolicyViolations     int                    `json:"policy_violations"`
	ResourcesByType      map[string]int         `json:"resources_by_type"`
	ViolationsByPolicy   map[string]int         `json:"violations_by_policy"`
	ExecutionsByStatus   map[string]int         `json:"executions_by_status"`
	AverageDuration      time.Duration          `json:"average_duration"`
	LastUpdated          time.Time              `json:"last_updated"`
}

// MetricsCollector collects and aggregates system metrics
type MetricsCollector struct {
	executions        []*ExecutionMetrics
	resources         []*ResourceMetrics
	policyViolations  []*PolicyViolationMetrics
	enabled           bool
	mu                sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		executions:       make([]*ExecutionMetrics, 0),
		resources:        make([]*ResourceMetrics, 0),
		policyViolations: make([]*PolicyViolationMetrics, 0),
		enabled:          true,
	}
}

// IsEnabled returns whether metrics collection is enabled
func (m *MetricsCollector) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// Enable enables metrics collection
func (m *MetricsCollector) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
}

// Disable disables metrics collection
func (m *MetricsCollector) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
}

// RecordExecution records execution metrics
func (m *MetricsCollector) RecordExecution(execution *ExecutionMetrics) error {
	if !m.IsEnabled() {
		return nil
	}
	
	if err := execution.Validate(); err != nil {
		return fmt.Errorf("invalid execution metrics: %w", err)
	}
	
	// Set timestamp if not provided
	if execution.Timestamp.IsZero() {
		execution.Timestamp = time.Now()
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.executions = append(m.executions, execution)
	
	// Keep only the last 1000 executions
	if len(m.executions) > 1000 {
		m.executions = m.executions[1:]
	}
	
	return nil
}

// RecordResource records resource metrics
func (m *MetricsCollector) RecordResource(resource *ResourceMetrics) error {
	if !m.IsEnabled() {
		return nil
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.resources = append(m.resources, resource)
	
	// Keep only the last 5000 resources
	if len(m.resources) > 5000 {
		m.resources = m.resources[1:]
	}
	
	return nil
}

// RecordPolicyViolation records policy violation metrics
func (m *MetricsCollector) RecordPolicyViolation(violation *PolicyViolationMetrics) error {
	if !m.IsEnabled() {
		return nil
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.policyViolations = append(m.policyViolations, violation)
	
	// Keep only the last 1000 violations
	if len(m.policyViolations) > 1000 {
		m.policyViolations = m.policyViolations[1:]
	}
	
	return nil
}

// GetMetrics returns aggregated metrics
func (m *MetricsCollector) GetMetrics() *AggregatedMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	metrics := &AggregatedMetrics{
		ResourcesByType:    make(map[string]int),
		ViolationsByPolicy: make(map[string]int),
		ExecutionsByStatus: make(map[string]int),
		LastUpdated:        time.Now(),
	}
	
	// Aggregate execution metrics
	var totalDuration time.Duration
	for _, exec := range m.executions {
		metrics.TotalExecutions++
		metrics.ExecutionsByStatus[exec.Status]++
		totalDuration += exec.Duration
		
		if exec.Status == "success" {
			metrics.SuccessfulExecutions++
		} else if exec.Status == "failed" {
			metrics.FailedExecutions++
		}
	}
	
	// Calculate average duration
	if metrics.TotalExecutions > 0 {
		metrics.AverageDuration = totalDuration / time.Duration(metrics.TotalExecutions)
	}
	
	// Aggregate resource metrics
	for _, resource := range m.resources {
		metrics.TotalResources++
		metrics.ResourcesByType[resource.Type]++
	}
	
	// Aggregate policy violation metrics
	for _, violation := range m.policyViolations {
		metrics.PolicyViolations++
		metrics.ViolationsByPolicy[violation.Policy]++
	}
	
	return metrics
}

// GetMetricsByTimeRange returns metrics for a specific time range
func (m *MetricsCollector) GetMetricsByTimeRange(start, end time.Time) *AggregatedMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	metrics := &AggregatedMetrics{
		ResourcesByType:    make(map[string]int),
		ViolationsByPolicy: make(map[string]int),
		ExecutionsByStatus: make(map[string]int),
		LastUpdated:        time.Now(),
	}
	
	// Filter executions by time range
	var totalDuration time.Duration
	for _, exec := range m.executions {
		if exec.Timestamp.After(start) && exec.Timestamp.Before(end) {
			metrics.TotalExecutions++
			metrics.ExecutionsByStatus[exec.Status]++
			totalDuration += exec.Duration
			
			if exec.Status == "success" {
				metrics.SuccessfulExecutions++
			} else if exec.Status == "failed" {
				metrics.FailedExecutions++
			}
		}
	}
	
	// Calculate average duration
	if metrics.TotalExecutions > 0 {
		metrics.AverageDuration = totalDuration / time.Duration(metrics.TotalExecutions)
	}
	
	// Note: Resources and violations don't have timestamps in this implementation
	// In a real implementation, they would be filtered by time range as well
	
	return metrics
}

// GetPrometheusMetrics returns metrics in Prometheus format
func (m *MetricsCollector) GetPrometheusMetrics() string {
	metrics := m.GetMetrics()
	
	prometheus := fmt.Sprintf(`# HELP chisel_executions_total Total number of executions
# TYPE chisel_executions_total counter
chisel_executions_total %d

# HELP chisel_executions_successful_total Total number of successful executions
# TYPE chisel_executions_successful_total counter
chisel_executions_successful_total %d

# HELP chisel_executions_failed_total Total number of failed executions
# TYPE chisel_executions_failed_total counter
chisel_executions_failed_total %d

# HELP chisel_execution_duration_seconds Average execution duration in seconds
# TYPE chisel_execution_duration_seconds gauge
chisel_execution_duration_seconds %.2f

# HELP chisel_resources_total Total number of resources processed
# TYPE chisel_resources_total counter
chisel_resources_total %d

# HELP chisel_policy_violations_total Total number of policy violations
# TYPE chisel_policy_violations_total counter
chisel_policy_violations_total %d
`,
		metrics.TotalExecutions,
		metrics.SuccessfulExecutions,
		metrics.FailedExecutions,
		metrics.AverageDuration.Seconds(),
		metrics.TotalResources,
		metrics.PolicyViolations,
	)
	
	// Add resource type breakdown
	for resourceType, count := range metrics.ResourcesByType {
		prometheus += fmt.Sprintf(`
# HELP chisel_resources_by_type_total Resources processed by type
# TYPE chisel_resources_by_type_total counter
chisel_resources_by_type_total{type="%s"} %d`, resourceType, count)
	}
	
	return prometheus
}

// ExportToFile exports metrics to a JSON file
func (m *MetricsCollector) ExportToFile(filename string) error {
	metrics := m.GetMetrics()
	
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}
	
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write metrics file: %w", err)
	}
	
	return nil
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Enabled           bool   `yaml:"enabled" json:"enabled"`
	PrometheusEnabled bool   `yaml:"prometheus_enabled" json:"prometheus_enabled"`
	PrometheusPort    int    `yaml:"prometheus_port" json:"prometheus_port"`
	MetricsFile       string `yaml:"metrics_file" json:"metrics_file"`
	ExportInterval    string `yaml:"export_interval" json:"export_interval"`
}

// DefaultMonitoringConfig returns default monitoring configuration
func DefaultMonitoringConfig() *MonitoringConfig {
	return &MonitoringConfig{
		Enabled:           true,
		PrometheusEnabled: false,
		PrometheusPort:    9090,
		MetricsFile:       "/var/log/chisel/metrics.json",
		ExportInterval:    "5m",
	}
}
