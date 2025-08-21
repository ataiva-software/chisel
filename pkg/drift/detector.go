package drift

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ataiva-software/chisel/pkg/core"
	"github.com/ataiva-software/chisel/pkg/types"
)

// DriftResult represents the result of a drift detection check
type DriftResult struct {
	ResourceID    string                 `json:"resource_id"`
	HasDrift      bool                   `json:"has_drift"`
	Changes       map[string]interface{} `json:"changes,omitempty"`
	LastChecked   time.Time              `json:"last_checked"`
	Error         error                  `json:"error,omitempty"`
	CheckDuration time.Duration          `json:"check_duration"`
}

// DriftReport represents a complete drift detection report
type DriftReport struct {
	ModuleName    string        `json:"module_name"`
	Timestamp     time.Time     `json:"timestamp"`
	TotalChecked  int           `json:"total_checked"`
	DriftDetected int           `json:"drift_detected"`
	Errors        int           `json:"errors"`
	Results       []DriftResult `json:"results"`
	Duration      time.Duration `json:"duration"`
}

// DriftDetector performs drift detection on resources
type DriftDetector struct {
	planner  *core.Planner
	registry *types.ProviderRegistry
	interval time.Duration
	enabled  bool
	mu       sync.RWMutex
	
	// Channels for control
	stopChan   chan struct{}
	reportChan chan *DriftReport
	
	// Configuration
	maxConcurrency int
	timeout        time.Duration
	
	// State
	lastReport *DriftReport
	running    bool
}

// NewDriftDetector creates a new drift detector
func NewDriftDetector(planner *core.Planner, registry *types.ProviderRegistry, interval time.Duration) *DriftDetector {
	return &DriftDetector{
		planner:        planner,
		registry:       registry,
		interval:       interval,
		enabled:        false,
		stopChan:       make(chan struct{}),
		reportChan:     make(chan *DriftReport, 10),
		maxConcurrency: 5,
		timeout:        30 * time.Second,
	}
}

// Start starts the drift detection scheduler
func (d *DriftDetector) Start(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if d.running {
		return fmt.Errorf("drift detector is already running")
	}
	
	d.enabled = true
	d.running = true
	
	go d.schedulerLoop(ctx)
	
	return nil
}

// Stop stops the drift detection scheduler
func (d *DriftDetector) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if !d.running {
		return nil
	}
	
	d.enabled = false
	close(d.stopChan)
	d.running = false
	
	return nil
}

// CheckDrift performs a one-time drift check on a module
func (d *DriftDetector) CheckDrift(ctx context.Context, module *core.Module) (*DriftReport, error) {
	start := time.Now()
	
	report := &DriftReport{
		ModuleName:   module.Metadata.Name,
		Timestamp:    start,
		TotalChecked: len(module.Spec.Resources),
		Results:      make([]DriftResult, 0, len(module.Spec.Resources)),
	}
	
	// Create a semaphore for concurrency control
	semaphore := make(chan struct{}, d.maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	// Check each resource for drift
	for _, resource := range module.Spec.Resources {
		wg.Add(1)
		go func(res types.Resource) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			result := d.checkResourceDrift(ctx, &res)
			
			mu.Lock()
			report.Results = append(report.Results, result)
			if result.HasDrift {
				report.DriftDetected++
			}
			if result.Error != nil {
				report.Errors++
			}
			mu.Unlock()
		}(resource)
	}
	
	// Wait for all checks to complete
	wg.Wait()
	
	report.Duration = time.Since(start)
	
	// Update last report
	d.mu.Lock()
	d.lastReport = report
	d.mu.Unlock()
	
	return report, nil
}

// checkResourceDrift checks a single resource for drift
func (d *DriftDetector) checkResourceDrift(ctx context.Context, resource *types.Resource) DriftResult {
	start := time.Now()
	
	result := DriftResult{
		ResourceID:    resource.ResourceID(),
		LastChecked:   start,
		CheckDuration: 0,
		HasDrift:      false,
	}
	
	// Create context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	
	// Get provider for this resource type
	provider, err := d.registry.Get(resource.Type)
	if err != nil {
		result.Error = fmt.Errorf("no provider for resource type %s: %w", resource.Type, err)
		result.CheckDuration = time.Since(start)
		return result
	}
	
	// Read current state
	currentState, err := provider.Read(checkCtx, resource)
	if err != nil {
		result.Error = fmt.Errorf("failed to read current state: %w", err)
		result.CheckDuration = time.Since(start)
		return result
	}
	
	// Calculate diff
	diff, err := provider.Diff(checkCtx, resource, currentState)
	if err != nil {
		result.Error = fmt.Errorf("failed to calculate diff: %w", err)
		result.CheckDuration = time.Since(start)
		return result
	}
	
	// Check if there's drift
	if diff.Action != types.ActionNoop {
		result.HasDrift = true
		result.Changes = diff.Changes
	}
	
	result.CheckDuration = time.Since(start)
	return result
}

// schedulerLoop runs the drift detection scheduler
func (d *DriftDetector) schedulerLoop(ctx context.Context) {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopChan:
			return
		case <-ticker.C:
			if d.enabled {
				// This would be enhanced to check configured modules
				// For now, it's a placeholder for the scheduler logic
				fmt.Printf("Drift detection scheduled check at %v\n", time.Now())
			}
		}
	}
}

// GetLastReport returns the last drift detection report
func (d *DriftDetector) GetLastReport() *DriftReport {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.lastReport
}

// GetReportChannel returns the channel for receiving drift reports
func (d *DriftDetector) GetReportChannel() <-chan *DriftReport {
	return d.reportChan
}

// IsRunning returns whether the drift detector is currently running
func (d *DriftDetector) IsRunning() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.running
}

// SetInterval updates the drift detection interval
func (d *DriftDetector) SetInterval(interval time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.interval = interval
}

// SetConcurrency updates the maximum concurrency for drift checks
func (d *DriftDetector) SetConcurrency(maxConcurrency int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if maxConcurrency > 0 {
		d.maxConcurrency = maxConcurrency
	}
}

// SetTimeout updates the timeout for individual drift checks
func (d *DriftDetector) SetTimeout(timeout time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if timeout > 0 {
		d.timeout = timeout
	}
}

// DriftConfig represents configuration for drift detection
type DriftConfig struct {
	Enabled        bool          `yaml:"enabled" json:"enabled"`
	Interval       time.Duration `yaml:"interval" json:"interval"`
	MaxConcurrency int           `yaml:"max_concurrency" json:"max_concurrency"`
	Timeout        time.Duration `yaml:"timeout" json:"timeout"`
	Modules        []string      `yaml:"modules" json:"modules"`
	Notifications  struct {
		Enabled   bool     `yaml:"enabled" json:"enabled"`
		Channels  []string `yaml:"channels" json:"channels"`
		Threshold int      `yaml:"threshold" json:"threshold"` // Minimum drift count to trigger notification
	} `yaml:"notifications" json:"notifications"`
}

// DefaultDriftConfig returns a default drift detection configuration
func DefaultDriftConfig() *DriftConfig {
	return &DriftConfig{
		Enabled:        false,
		Interval:       15 * time.Minute,
		MaxConcurrency: 5,
		Timeout:        30 * time.Second,
		Modules:        []string{},
		Notifications: struct {
			Enabled   bool     `yaml:"enabled" json:"enabled"`
			Channels  []string `yaml:"channels" json:"channels"`
			Threshold int      `yaml:"threshold" json:"threshold"`
		}{
			Enabled:   false,
			Channels:  []string{},
			Threshold: 1,
		},
	}
}

// Validate validates the drift configuration
func (c *DriftConfig) Validate() error {
	if c.Interval <= 0 {
		return fmt.Errorf("drift detection interval must be positive")
	}
	
	if c.MaxConcurrency <= 0 {
		return fmt.Errorf("max concurrency must be positive")
	}
	
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	
	return nil
}

// DriftNotifier handles notifications for drift detection
type DriftNotifier struct {
	channels  []NotificationChannel
	threshold int
	enabled   bool
}

// NotificationChannel represents a notification channel
type NotificationChannel interface {
	Send(ctx context.Context, report *DriftReport) error
	Type() string
}

// NewDriftNotifier creates a new drift notifier
func NewDriftNotifier(threshold int) *DriftNotifier {
	return &DriftNotifier{
		channels:  make([]NotificationChannel, 0),
		threshold: threshold,
		enabled:   false,
	}
}

// AddChannel adds a notification channel
func (n *DriftNotifier) AddChannel(channel NotificationChannel) {
	n.channels = append(n.channels, channel)
}

// Enable enables notifications
func (n *DriftNotifier) Enable() {
	n.enabled = true
}

// Disable disables notifications
func (n *DriftNotifier) Disable() {
	n.enabled = false
}

// Notify sends notifications if drift is detected above threshold
func (n *DriftNotifier) Notify(ctx context.Context, report *DriftReport) error {
	if !n.enabled || report.DriftDetected < n.threshold {
		return nil
	}
	
	var errors []error
	for _, channel := range n.channels {
		if err := channel.Send(ctx, report); err != nil {
			errors = append(errors, fmt.Errorf("failed to send notification via %s: %w", channel.Type(), err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("notification errors: %v", errors)
	}
	
	return nil
}

// LogNotificationChannel logs drift reports to stdout
type LogNotificationChannel struct{}

// Send sends a notification to the log
func (l *LogNotificationChannel) Send(ctx context.Context, report *DriftReport) error {
	fmt.Printf("DRIFT DETECTED: Module %s has %d resources with drift (checked %d resources)\n",
		report.ModuleName, report.DriftDetected, report.TotalChecked)
	
	for _, result := range report.Results {
		if result.HasDrift {
			fmt.Printf("  - %s: drift detected\n", result.ResourceID)
		}
	}
	
	return nil
}

// Type returns the notification channel type
func (l *LogNotificationChannel) Type() string {
	return "log"
}
