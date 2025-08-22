package notifications

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ataiva-software/forge/pkg/events"
)

// NotificationManager manages notification channels and routing
type NotificationManager struct {
	channels map[string]NotificationChannel
	rules    []NotificationRule
	mu       sync.RWMutex
	
	// Event integration
	eventBus *events.EventBus
	emitter  *events.EventEmitter
	
	// Configuration
	enabled     bool
	rateLimiter *RateLimiter
}

// NotificationRule defines when and how to send notifications
type NotificationRule struct {
	Name        string                        `json:"name"`
	Enabled     bool                          `json:"enabled"`
	Channels    []string                      `json:"channels"`
	EventTypes  []events.EventType            `json:"event_types"`
	Levels      []NotificationLevel           `json:"levels"`
	Conditions  map[string]interface{}        `json:"conditions,omitempty"`
	Template    NotificationTemplate          `json:"template"`
	RateLimit   *RateLimitConfig              `json:"rate_limit,omitempty"`
}

// NotificationTemplate defines how to format notifications
type NotificationTemplate struct {
	TitleTemplate   string `json:"title_template"`
	MessageTemplate string `json:"message_template"`
	DefaultLevel    NotificationLevel `json:"default_level"`
}

// RateLimitConfig defines rate limiting for notifications
type RateLimitConfig struct {
	MaxNotifications int           `json:"max_notifications"`
	TimeWindow       time.Duration `json:"time_window"`
	BurstSize        int           `json:"burst_size"`
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(eventBus *events.EventBus) *NotificationManager {
	manager := &NotificationManager{
		channels:    make(map[string]NotificationChannel),
		rules:       make([]NotificationRule, 0),
		enabled:     true,
		rateLimiter: NewRateLimiter(100, time.Minute, 10), // Default rate limit
	}
	
	if eventBus != nil {
		manager.eventBus = eventBus
		manager.emitter = events.NewEventEmitter(eventBus, "notification-manager")
		
		// Subscribe to events for automatic notification generation
		eventHandler := &NotificationEventHandler{manager: manager}
		eventBus.Subscribe(eventHandler)
	}
	
	return manager
}

// AddChannel adds a notification channel
func (m *NotificationManager) AddChannel(channel NotificationChannel) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if channel == nil {
		return fmt.Errorf("channel cannot be nil")
	}
	
	name := channel.Name()
	if name == "" {
		return fmt.Errorf("channel name cannot be empty")
	}
	
	m.channels[name] = channel
	return nil
}

// RemoveChannel removes a notification channel
func (m *NotificationManager) RemoveChannel(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.channels[name]; !exists {
		return fmt.Errorf("channel '%s' not found", name)
	}
	
	delete(m.channels, name)
	return nil
}

// AddRule adds a notification rule
func (m *NotificationManager) AddRule(rule NotificationRule) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Validate rule
	if rule.Name == "" {
		return fmt.Errorf("rule name cannot be empty")
	}
	
	if len(rule.Channels) == 0 {
		return fmt.Errorf("rule must specify at least one channel")
	}
	
	// Check that all specified channels exist
	for _, channelName := range rule.Channels {
		if _, exists := m.channels[channelName]; !exists {
			return fmt.Errorf("channel '%s' not found", channelName)
		}
	}
	
	m.rules = append(m.rules, rule)
	return nil
}

// Send sends a notification through appropriate channels
func (m *NotificationManager) Send(ctx context.Context, notification *Notification) error {
	if !m.enabled {
		return nil
	}
	
	// Check rate limiting
	if !m.rateLimiter.Allow() {
		return fmt.Errorf("notification rate limit exceeded")
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Find matching rules
	matchingRules := m.findMatchingRules(notification)
	if len(matchingRules) == 0 {
		return nil // No rules match, don't send
	}
	
	// Collect unique channels from all matching rules
	channelNames := make(map[string]bool)
	for _, rule := range matchingRules {
		if !rule.Enabled {
			continue
		}
		
		for _, channelName := range rule.Channels {
			channelNames[channelName] = true
		}
	}
	
	// Send to all channels
	var errors []error
	for channelName := range channelNames {
		channel, exists := m.channels[channelName]
		if !exists {
			errors = append(errors, fmt.Errorf("channel '%s' not found", channelName))
			continue
		}
		
		if err := channel.Send(ctx, notification); err != nil {
			errors = append(errors, fmt.Errorf("failed to send to channel '%s': %w", channelName, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("notification errors: %v", errors)
	}
	
	return nil
}

// SendToChannels sends a notification to specific channels
func (m *NotificationManager) SendToChannels(ctx context.Context, notification *Notification, channelNames []string) error {
	if !m.enabled {
		return nil
	}
	
	if !m.rateLimiter.Allow() {
		return fmt.Errorf("notification rate limit exceeded")
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var errors []error
	for _, channelName := range channelNames {
		channel, exists := m.channels[channelName]
		if !exists {
			errors = append(errors, fmt.Errorf("channel '%s' not found", channelName))
			continue
		}
		
		if err := channel.Send(ctx, notification); err != nil {
			errors = append(errors, fmt.Errorf("failed to send to channel '%s': %w", channelName, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("notification errors: %v", errors)
	}
	
	return nil
}

// findMatchingRules finds rules that match the notification
func (m *NotificationManager) findMatchingRules(notification *Notification) []NotificationRule {
	var matching []NotificationRule
	
	for _, rule := range m.rules {
		if m.ruleMatches(rule, notification) {
			matching = append(matching, rule)
		}
	}
	
	return matching
}

// ruleMatches checks if a rule matches a notification
func (m *NotificationManager) ruleMatches(rule NotificationRule, notification *Notification) bool {
	// Check level filter
	if len(rule.Levels) > 0 {
		levelMatches := false
		for _, level := range rule.Levels {
			if level == notification.Level {
				levelMatches = true
				break
			}
		}
		if !levelMatches {
			return false
		}
	}
	
	// Check conditions
	for key, expectedValue := range rule.Conditions {
		actualValue, exists := notification.Data[key]
		if !exists || actualValue != expectedValue {
			return false
		}
	}
	
	return true
}

// Enable enables the notification manager
func (m *NotificationManager) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
}

// Disable disables the notification manager
func (m *NotificationManager) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
}

// IsEnabled returns whether the notification manager is enabled
func (m *NotificationManager) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// GetChannels returns a list of registered channel names
func (m *NotificationManager) GetChannels() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	names := make([]string, 0, len(m.channels))
	for name := range m.channels {
		names = append(names, name)
	}
	return names
}

// GetRules returns a copy of all notification rules
func (m *NotificationManager) GetRules() []NotificationRule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	rules := make([]NotificationRule, len(m.rules))
	copy(rules, m.rules)
	return rules
}

// NotificationEventHandler handles events and generates notifications
type NotificationEventHandler struct {
	manager *NotificationManager
}

// Handle handles an event by generating appropriate notifications
func (h *NotificationEventHandler) Handle(ctx context.Context, event *events.Event) error {
	notification := h.eventToNotification(event)
	if notification == nil {
		return nil // No notification needed for this event
	}
	
	return h.manager.Send(ctx, notification)
}

// Types returns the event types this handler subscribes to
func (h *NotificationEventHandler) Types() []events.EventType {
	return []events.EventType{
		events.EventTypeResourceFailed,
		events.EventTypeApplyFailed,
		events.EventTypePlanFailed,
		events.EventTypeDriftDetected,
		events.EventTypeRollbackStarted,
		events.EventTypeApplyCompleted,
	}
}

// Name returns the handler name
func (h *NotificationEventHandler) Name() string {
	return "notification-event-handler"
}

// eventToNotification converts an event to a notification
func (h *NotificationEventHandler) eventToNotification(event *events.Event) *Notification {
	var title, message string
	var level NotificationLevel
	
	switch event.Type {
	case events.EventTypeResourceFailed:
		title = "Resource Failed"
		message = fmt.Sprintf("Resource %s failed: %s", 
			event.Data["resource_id"], event.Data["error"])
		level = LevelError
		
	case events.EventTypeApplyFailed:
		title = "Apply Failed"
		message = fmt.Sprintf("Apply operation failed for module %s", 
			event.Data["module_name"])
		level = LevelCritical
		
	case events.EventTypePlanFailed:
		title = "Plan Failed"
		message = fmt.Sprintf("Plan operation failed for module %s", 
			event.Data["module_name"])
		level = LevelError
		
	case events.EventTypeDriftDetected:
		title = "Configuration Drift Detected"
		message = fmt.Sprintf("Drift detected in resource %s of module %s", 
			event.Data["resource_id"], event.Data["module_name"])
		level = LevelWarning
		
	case events.EventTypeRollbackStarted:
		title = "Rollback Started"
		message = "Automatic rollback initiated due to execution failure"
		level = LevelWarning
		
	case events.EventTypeApplyCompleted:
		title = "Apply Completed"
		message = fmt.Sprintf("Apply operation completed successfully for module %s", 
			event.Data["module_name"])
		level = LevelInfo
		
	default:
		return nil // No notification for this event type
	}
	
	notification := NewNotification(title, message, level)
	
	// Copy event data to notification
	for key, value := range event.Data {
		notification.AddData(key, value)
	}
	
	// Add event metadata
	notification.AddData("event_id", event.ID)
	notification.AddData("event_source", event.Source)
	notification.AddData("event_timestamp", event.Timestamp)
	
	return notification
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	maxTokens   int
	tokens      int
	refillRate  time.Duration
	lastRefill  time.Time
	burstSize   int
	mu          sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int, refillInterval time.Duration, burstSize int) *RateLimiter {
	return &RateLimiter{
		maxTokens:   maxTokens,
		tokens:      maxTokens,
		refillRate:  refillInterval,
		lastRefill:  time.Now(),
		burstSize:   burstSize,
	}
}

// Allow checks if an operation is allowed under the rate limit
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	now := time.Now()
	
	// Refill tokens based on time elapsed
	if now.Sub(r.lastRefill) >= r.refillRate {
		tokensToAdd := int(now.Sub(r.lastRefill) / r.refillRate)
		r.tokens = min(r.maxTokens, r.tokens+tokensToAdd)
		r.lastRefill = now
	}
	
	// Check if we have tokens available
	if r.tokens > 0 {
		r.tokens--
		return true
	}
	
	return false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
