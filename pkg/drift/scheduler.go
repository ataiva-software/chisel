package drift

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ataiva-software/forge/pkg/core"
	"github.com/ataiva-software/forge/pkg/types"
)

// DriftScheduler manages scheduled drift detection for multiple modules
type DriftScheduler struct {
	planner  *core.Planner
	registry *types.ProviderRegistry
	
	// Scheduled modules
	modules map[string]*ScheduledModule
	mu      sync.RWMutex
	
	// Scheduler state
	enabled bool
	running bool
	
	// Control channels
	stopChan chan struct{}
	
	// Recent reports
	recentReports []*DriftReport
	maxReports    int
	reportsMu     sync.RWMutex
	
	// Configuration
	checkInterval time.Duration
}

// ScheduledModule represents a module scheduled for drift detection
type ScheduledModule struct {
	Module     *core.Module
	Config     *DriftScheduleConfig
	LastRun    time.Time
	NextRun    time.Time
	RunCount   int64
	ErrorCount int64
}

// DriftScheduleConfig defines the schedule configuration for drift detection
type DriftScheduleConfig struct {
	Interval    time.Duration `yaml:"interval" json:"interval"`
	Enabled     bool          `yaml:"enabled" json:"enabled"`
	MaxRetries  int           `yaml:"max_retries" json:"max_retries"`
	RetryDelay  time.Duration `yaml:"retry_delay" json:"retry_delay"`
	Timeout     time.Duration `yaml:"timeout" json:"timeout"`
}

// NewDriftScheduler creates a new drift scheduler
func NewDriftScheduler(planner *core.Planner, registry *types.ProviderRegistry) *DriftScheduler {
	return &DriftScheduler{
		planner:       planner,
		registry:      registry,
		modules:       make(map[string]*ScheduledModule),
		enabled:       true,
		running:       false,
		stopChan:      make(chan struct{}),
		recentReports: make([]*DriftReport, 0),
		maxReports:    100, // Keep last 100 reports
		checkInterval: 10 * time.Second, // Default check interval
	}
}

// SetCheckInterval sets the interval for checking scheduled modules (useful for testing)
func (s *DriftScheduler) SetCheckInterval(interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkInterval = interval
}

// Validate validates the drift schedule configuration
func (c *DriftScheduleConfig) Validate() error {
	// Allow zero values that will be set by defaults, but reject negative values
	if c.Interval < 0 {
		return fmt.Errorf("interval cannot be negative")
	}
	
	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}
	
	if c.RetryDelay < 0 {
		return fmt.Errorf("retry_delay cannot be negative")
	}
	
	return nil
}

// SetDefaults sets default values for the configuration
func (c *DriftScheduleConfig) SetDefaults() {
	if c.Interval == 0 {
		c.Interval = 15 * time.Minute
	}
	
	if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}
	
	if c.RetryDelay == 0 {
		c.RetryDelay = 30 * time.Second
	}
	
	if c.Timeout == 0 {
		c.Timeout = 5 * time.Minute
	}
}

// AddModule adds a module to the drift detection schedule
func (s *DriftScheduler) AddModule(module *core.Module, config *DriftScheduleConfig) error {
	if module == nil {
		return fmt.Errorf("module cannot be nil")
	}
	
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	// Validate first, then set defaults
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	
	config.SetDefaults()
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	scheduledModule := &ScheduledModule{
		Module:  module,
		Config:  config,
		NextRun: time.Now(), // Run immediately on first check
	}
	
	s.modules[module.Metadata.Name] = scheduledModule
	
	return nil
}

// RemoveModule removes a module from the drift detection schedule
func (s *DriftScheduler) RemoveModule(moduleName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.modules[moduleName]; !exists {
		return fmt.Errorf("module '%s' not found in schedule", moduleName)
	}
	
	delete(s.modules, moduleName)
	return nil
}

// GetScheduledModules returns a copy of all scheduled modules
func (s *DriftScheduler) GetScheduledModules() []*ScheduledModule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	modules := make([]*ScheduledModule, 0, len(s.modules))
	for _, module := range s.modules {
		// Create a copy
		moduleCopy := *module
		modules = append(modules, &moduleCopy)
	}
	
	return modules
}

// Start starts the drift detection scheduler
func (s *DriftScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.running {
		return fmt.Errorf("scheduler is already running")
	}
	
	s.running = true
	s.stopChan = make(chan struct{})
	
	// Start the scheduler loop
	go s.schedulerLoop(ctx)
	
	return nil
}

// Stop stops the drift detection scheduler
func (s *DriftScheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.running {
		return nil
	}
	
	s.running = false
	close(s.stopChan)
	
	return nil
}

// IsRunning returns whether the scheduler is currently running
func (s *DriftScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// IsEnabled returns whether the scheduler is enabled
func (s *DriftScheduler) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// Enable enables the scheduler
func (s *DriftScheduler) Enable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = true
}

// Disable disables the scheduler
func (s *DriftScheduler) Disable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = false
}

// GetRecentReports returns the most recent drift reports
func (s *DriftScheduler) GetRecentReports(limit int) []*DriftReport {
	s.reportsMu.RLock()
	defer s.reportsMu.RUnlock()
	
	if limit <= 0 || limit > len(s.recentReports) {
		limit = len(s.recentReports)
	}
	
	// Return the most recent reports
	start := len(s.recentReports) - limit
	reports := make([]*DriftReport, limit)
	copy(reports, s.recentReports[start:])
	
	return reports
}

// schedulerLoop runs the main scheduler loop
func (s *DriftScheduler) schedulerLoop(ctx context.Context) {
	s.mu.RLock()
	interval := s.checkInterval
	s.mu.RUnlock()
	
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			if s.enabled {
				s.checkAndRunDriftDetection(ctx)
			}
		}
	}
}

// checkAndRunDriftDetection checks if any modules need drift detection and runs them
func (s *DriftScheduler) checkAndRunDriftDetection(ctx context.Context) {
	s.mu.RLock()
	modulesToRun := make([]*ScheduledModule, 0)
	now := time.Now()
	
	for _, module := range s.modules {
		if module.Config.Enabled && now.After(module.NextRun) {
			modulesToRun = append(modulesToRun, module)
		}
	}
	s.mu.RUnlock()
	
	// Run drift detection for each module
	for _, module := range modulesToRun {
		go s.runDriftDetectionForModule(ctx, module)
	}
}

// runDriftDetectionForModule runs drift detection for a specific module
func (s *DriftScheduler) runDriftDetectionForModule(ctx context.Context, scheduledModule *ScheduledModule) {
	// Create detector
	detector := NewDriftDetector(s.planner, s.registry, scheduledModule.Config.Interval)
	
	// Run drift detection with retries
	var report *DriftReport
	var err error
	
	for attempt := 0; attempt <= scheduledModule.Config.MaxRetries; attempt++ {
		// Create context with timeout
		detectCtx, cancel := context.WithTimeout(ctx, scheduledModule.Config.Timeout)
		
		report, err = detector.CheckDrift(detectCtx, scheduledModule.Module)
		cancel()
		
		if err == nil {
			break
		}
		
		// If not the last attempt, wait before retrying
		if attempt < scheduledModule.Config.MaxRetries {
			select {
			case <-ctx.Done():
				return
			case <-time.After(scheduledModule.Config.RetryDelay):
				// Continue to next attempt
			}
		}
	}
	
	// Update module statistics
	s.mu.Lock()
	scheduledModule.LastRun = time.Now()
	scheduledModule.NextRun = scheduledModule.LastRun.Add(scheduledModule.Config.Interval)
	scheduledModule.RunCount++
	
	if err != nil {
		scheduledModule.ErrorCount++
	}
	s.mu.Unlock()
	
	// Store report if successful
	if report != nil {
		s.addReport(report)
	}
}

// addReport adds a report to the recent reports list
func (s *DriftScheduler) addReport(report *DriftReport) {
	s.reportsMu.Lock()
	defer s.reportsMu.Unlock()
	
	s.recentReports = append(s.recentReports, report)
	
	// Keep only the most recent reports
	if len(s.recentReports) > s.maxReports {
		s.recentReports = s.recentReports[1:]
	}
}
