package notifications

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ataiva-software/forge/pkg/events"
)

// TestNotificationChannel for testing
type TestNotificationChannel struct {
	name             string
	channelType      string
	sentNotifications []*Notification
	shouldFail       bool
	mu               sync.Mutex
}

func NewTestNotificationChannel(name, channelType string) *TestNotificationChannel {
	return &TestNotificationChannel{
		name:              name,
		channelType:       channelType,
		sentNotifications: make([]*Notification, 0),
	}
}

func (t *TestNotificationChannel) Send(ctx context.Context, notification *Notification) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if t.shouldFail {
		return fmt.Errorf("test channel failure")
	}
	
	// Make a copy of the notification
	notifCopy := *notification
	t.sentNotifications = append(t.sentNotifications, &notifCopy)
	return nil
}

func (t *TestNotificationChannel) Type() string {
	return t.channelType
}

func (t *TestNotificationChannel) Name() string {
	return t.name
}

func (t *TestNotificationChannel) GetSentNotifications() []*Notification {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	result := make([]*Notification, len(t.sentNotifications))
	copy(result, t.sentNotifications)
	return result
}

func (t *TestNotificationChannel) SetShouldFail(shouldFail bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.shouldFail = shouldFail
}

func TestNotificationManager_AddChannel(t *testing.T) {
	manager := NewNotificationManager(nil)
	
	channel := NewTestNotificationChannel("test-channel", "test")
	
	err := manager.AddChannel(channel)
	if err != nil {
		t.Fatalf("Failed to add channel: %v", err)
	}
	
	channels := manager.GetChannels()
	if len(channels) != 1 {
		t.Errorf("Expected 1 channel, got %d", len(channels))
	}
	
	if channels[0] != "test-channel" {
		t.Errorf("Expected channel name 'test-channel', got '%s'", channels[0])
	}
}

func TestNotificationManager_AddChannelErrors(t *testing.T) {
	manager := NewNotificationManager(nil)
	
	// Test nil channel
	err := manager.AddChannel(nil)
	if err == nil {
		t.Error("Expected error for nil channel, but got none")
	}
	
	// Test channel with empty name
	emptyNameChannel := &TestNotificationChannel{name: "", channelType: "test"}
	err = manager.AddChannel(emptyNameChannel)
	if err == nil {
		t.Error("Expected error for channel with empty name, but got none")
	}
}

func TestNotificationManager_RemoveChannel(t *testing.T) {
	manager := NewNotificationManager(nil)
	
	channel := NewTestNotificationChannel("test-channel", "test")
	manager.AddChannel(channel)
	
	err := manager.RemoveChannel("test-channel")
	if err != nil {
		t.Fatalf("Failed to remove channel: %v", err)
	}
	
	channels := manager.GetChannels()
	if len(channels) != 0 {
		t.Errorf("Expected 0 channels after removal, got %d", len(channels))
	}
	
	// Test removing non-existent channel
	err = manager.RemoveChannel("non-existent")
	if err == nil {
		t.Error("Expected error for removing non-existent channel, but got none")
	}
}

func TestNotificationManager_AddRule(t *testing.T) {
	manager := NewNotificationManager(nil)
	
	// Add a channel first
	channel := NewTestNotificationChannel("test-channel", "test")
	manager.AddChannel(channel)
	
	rule := NotificationRule{
		Name:     "test-rule",
		Enabled:  true,
		Channels: []string{"test-channel"},
		Levels:   []NotificationLevel{LevelError, LevelCritical},
	}
	
	err := manager.AddRule(rule)
	if err != nil {
		t.Fatalf("Failed to add rule: %v", err)
	}
	
	rules := manager.GetRules()
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}
	
	if rules[0].Name != "test-rule" {
		t.Errorf("Expected rule name 'test-rule', got '%s'", rules[0].Name)
	}
}

func TestNotificationManager_AddRuleErrors(t *testing.T) {
	manager := NewNotificationManager(nil)
	
	// Test rule with empty name
	rule := NotificationRule{
		Name:     "",
		Channels: []string{"test-channel"},
	}
	
	err := manager.AddRule(rule)
	if err == nil {
		t.Error("Expected error for rule with empty name, but got none")
	}
	
	// Test rule with no channels
	rule = NotificationRule{
		Name:     "test-rule",
		Channels: []string{},
	}
	
	err = manager.AddRule(rule)
	if err == nil {
		t.Error("Expected error for rule with no channels, but got none")
	}
	
	// Test rule with non-existent channel
	rule = NotificationRule{
		Name:     "test-rule",
		Channels: []string{"non-existent-channel"},
	}
	
	err = manager.AddRule(rule)
	if err == nil {
		t.Error("Expected error for rule with non-existent channel, but got none")
	}
}

func TestNotificationManager_Send(t *testing.T) {
	manager := NewNotificationManager(nil)
	
	// Add channels
	channel1 := NewTestNotificationChannel("channel1", "test")
	channel2 := NewTestNotificationChannel("channel2", "test")
	manager.AddChannel(channel1)
	manager.AddChannel(channel2)
	
	// Add rule that matches error notifications
	rule := NotificationRule{
		Name:     "error-rule",
		Enabled:  true,
		Channels: []string{"channel1", "channel2"},
		Levels:   []NotificationLevel{LevelError},
	}
	manager.AddRule(rule)
	
	// Send error notification
	notification := NewNotification("Test Error", "This is a test error", LevelError)
	
	ctx := context.Background()
	err := manager.Send(ctx, notification)
	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}
	
	// Verify both channels received the notification
	sent1 := channel1.GetSentNotifications()
	sent2 := channel2.GetSentNotifications()
	
	if len(sent1) != 1 {
		t.Errorf("Expected channel1 to receive 1 notification, got %d", len(sent1))
	}
	
	if len(sent2) != 1 {
		t.Errorf("Expected channel2 to receive 1 notification, got %d", len(sent2))
	}
	
	if len(sent1) > 0 && sent1[0].Title != "Test Error" {
		t.Errorf("Expected notification title 'Test Error', got '%s'", sent1[0].Title)
	}
}

func TestNotificationManager_SendNoMatchingRules(t *testing.T) {
	manager := NewNotificationManager(nil)
	
	channel := NewTestNotificationChannel("test-channel", "test")
	manager.AddChannel(channel)
	
	// Add rule that only matches critical notifications
	rule := NotificationRule{
		Name:     "critical-rule",
		Enabled:  true,
		Channels: []string{"test-channel"},
		Levels:   []NotificationLevel{LevelCritical},
	}
	manager.AddRule(rule)
	
	// Send info notification (should not match)
	notification := NewNotification("Test Info", "This is a test info", LevelInfo)
	
	ctx := context.Background()
	err := manager.Send(ctx, notification)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Verify no notifications were sent
	sent := channel.GetSentNotifications()
	if len(sent) != 0 {
		t.Errorf("Expected no notifications to be sent, got %d", len(sent))
	}
}

func TestNotificationManager_SendWithConditions(t *testing.T) {
	manager := NewNotificationManager(nil)
	
	channel := NewTestNotificationChannel("test-channel", "test")
	manager.AddChannel(channel)
	
	// Add rule with conditions
	rule := NotificationRule{
		Name:     "conditional-rule",
		Enabled:  true,
		Channels: []string{"test-channel"},
		Levels:   []NotificationLevel{LevelError},
		Conditions: map[string]interface{}{
			"module": "web-server",
		},
	}
	manager.AddRule(rule)
	
	// Send notification that matches conditions
	notification1 := NewNotification("Test Error", "This is a test error", LevelError)
	notification1.AddData("module", "web-server")
	
	ctx := context.Background()
	err := manager.Send(ctx, notification1)
	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}
	
	// Send notification that doesn't match conditions
	notification2 := NewNotification("Test Error", "This is a test error", LevelError)
	notification2.AddData("module", "database")
	
	err = manager.Send(ctx, notification2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Verify only the first notification was sent
	sent := channel.GetSentNotifications()
	if len(sent) != 1 {
		t.Errorf("Expected 1 notification to be sent, got %d", len(sent))
	}
}

func TestNotificationManager_SendToChannels(t *testing.T) {
	manager := NewNotificationManager(nil)
	
	channel1 := NewTestNotificationChannel("channel1", "test")
	channel2 := NewTestNotificationChannel("channel2", "test")
	manager.AddChannel(channel1)
	manager.AddChannel(channel2)
	
	notification := NewNotification("Test", "Message", LevelInfo)
	
	ctx := context.Background()
	err := manager.SendToChannels(ctx, notification, []string{"channel1"})
	if err != nil {
		t.Fatalf("Failed to send to specific channels: %v", err)
	}
	
	// Verify only channel1 received the notification
	sent1 := channel1.GetSentNotifications()
	sent2 := channel2.GetSentNotifications()
	
	if len(sent1) != 1 {
		t.Errorf("Expected channel1 to receive 1 notification, got %d", len(sent1))
	}
	
	if len(sent2) != 0 {
		t.Errorf("Expected channel2 to receive 0 notifications, got %d", len(sent2))
	}
}

func TestNotificationManager_DisabledRule(t *testing.T) {
	manager := NewNotificationManager(nil)
	
	channel := NewTestNotificationChannel("test-channel", "test")
	manager.AddChannel(channel)
	
	// Add disabled rule
	rule := NotificationRule{
		Name:     "disabled-rule",
		Enabled:  false,
		Channels: []string{"test-channel"},
		Levels:   []NotificationLevel{LevelError},
	}
	manager.AddRule(rule)
	
	notification := NewNotification("Test Error", "This is a test error", LevelError)
	
	ctx := context.Background()
	err := manager.Send(ctx, notification)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Verify no notifications were sent (rule is disabled)
	sent := channel.GetSentNotifications()
	if len(sent) != 0 {
		t.Errorf("Expected no notifications to be sent (rule disabled), got %d", len(sent))
	}
}

func TestNotificationManager_EnableDisable(t *testing.T) {
	manager := NewNotificationManager(nil)
	
	if !manager.IsEnabled() {
		t.Error("Expected manager to be enabled by default")
	}
	
	manager.Disable()
	if manager.IsEnabled() {
		t.Error("Expected manager to be disabled after Disable()")
	}
	
	manager.Enable()
	if !manager.IsEnabled() {
		t.Error("Expected manager to be enabled after Enable()")
	}
}

func TestNotificationManager_SendWhenDisabled(t *testing.T) {
	manager := NewNotificationManager(nil)
	manager.Disable()
	
	channel := NewTestNotificationChannel("test-channel", "test")
	manager.AddChannel(channel)
	
	rule := NotificationRule{
		Name:     "test-rule",
		Enabled:  true,
		Channels: []string{"test-channel"},
		Levels:   []NotificationLevel{LevelError},
	}
	manager.AddRule(rule)
	
	notification := NewNotification("Test Error", "This is a test error", LevelError)
	
	ctx := context.Background()
	err := manager.Send(ctx, notification)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Verify no notifications were sent (manager is disabled)
	sent := channel.GetSentNotifications()
	if len(sent) != 0 {
		t.Errorf("Expected no notifications to be sent (manager disabled), got %d", len(sent))
	}
}

func TestNotificationManager_ChannelFailure(t *testing.T) {
	manager := NewNotificationManager(nil)
	
	channel1 := NewTestNotificationChannel("channel1", "test")
	channel2 := NewTestNotificationChannel("channel2", "test")
	channel1.SetShouldFail(true) // Make channel1 fail
	
	manager.AddChannel(channel1)
	manager.AddChannel(channel2)
	
	rule := NotificationRule{
		Name:     "test-rule",
		Enabled:  true,
		Channels: []string{"channel1", "channel2"},
		Levels:   []NotificationLevel{LevelError},
	}
	manager.AddRule(rule)
	
	notification := NewNotification("Test Error", "This is a test error", LevelError)
	
	ctx := context.Background()
	err := manager.Send(ctx, notification)
	if err == nil {
		t.Error("Expected error due to channel failure, but got none")
	}
	
	// Verify channel2 still received the notification
	sent2 := channel2.GetSentNotifications()
	if len(sent2) != 1 {
		t.Errorf("Expected channel2 to receive 1 notification despite channel1 failure, got %d", len(sent2))
	}
}

func TestRateLimiter(t *testing.T) {
	// Create rate limiter: 2 tokens, refill every 100ms
	limiter := NewRateLimiter(2, 100*time.Millisecond, 2)
	
	// Should allow first 2 requests
	if !limiter.Allow() {
		t.Error("Expected first request to be allowed")
	}
	
	if !limiter.Allow() {
		t.Error("Expected second request to be allowed")
	}
	
	// Third request should be denied (no tokens left)
	if limiter.Allow() {
		t.Error("Expected third request to be denied")
	}
	
	// Wait for refill and try again
	time.Sleep(150 * time.Millisecond)
	
	if !limiter.Allow() {
		t.Error("Expected request to be allowed after refill")
	}
}

func TestNotificationManager_RateLimit(t *testing.T) {
	manager := NewNotificationManager(nil)
	
	// Set a very restrictive rate limiter
	manager.rateLimiter = NewRateLimiter(1, time.Hour, 1)
	
	channel := NewTestNotificationChannel("test-channel", "test")
	manager.AddChannel(channel)
	
	rule := NotificationRule{
		Name:     "test-rule",
		Enabled:  true,
		Channels: []string{"test-channel"},
		Levels:   []NotificationLevel{LevelError},
	}
	manager.AddRule(rule)
	
	notification := NewNotification("Test Error", "This is a test error", LevelError)
	
	ctx := context.Background()
	
	// First notification should succeed
	err := manager.Send(ctx, notification)
	if err != nil {
		t.Fatalf("First notification should succeed: %v", err)
	}
	
	// Second notification should be rate limited
	err = manager.Send(ctx, notification)
	if err == nil {
		t.Error("Expected second notification to be rate limited, but it succeeded")
	}
	
	if !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("Expected rate limit error, got: %v", err)
	}
}

func TestNotificationEventHandler(t *testing.T) {
	bus := events.NewEventBus(10, 2)
	defer bus.Close()
	
	manager := NewNotificationManager(bus)
	
	channel := NewTestNotificationChannel("test-channel", "test")
	manager.AddChannel(channel)
	
	// Add rule for all notification levels
	rule := NotificationRule{
		Name:     "all-events-rule",
		Enabled:  true,
		Channels: []string{"test-channel"},
		Levels:   []NotificationLevel{LevelInfo, LevelWarning, LevelError, LevelCritical},
	}
	manager.AddRule(rule)
	
	// Create event emitter
	emitter := events.NewEventEmitter(bus, "test")
	
	// Emit various events
	
	// Resource failed event
	err := emitter.EmitResourceFailed("test.resource", "create", fmt.Errorf("test error"), time.Second)
	if err != nil {
		t.Fatalf("Failed to emit resource failed event: %v", err)
	}
	
	// Apply completed event
	summary := map[string]int{"created": 1, "updated": 2}
	err = emitter.EmitApplyCompleted("test-module", summary, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to emit apply completed event: %v", err)
	}
	
	// Give time for async processing
	time.Sleep(200 * time.Millisecond)
	
	// Verify notifications were generated
	sent := channel.GetSentNotifications()
	if len(sent) < 2 {
		t.Errorf("Expected at least 2 notifications, got %d", len(sent))
	}
	
	// Check that we got the expected notification types
	var hasResourceFailed, hasApplyCompleted bool
	for _, notification := range sent {
		if strings.Contains(notification.Title, "Resource Failed") {
			hasResourceFailed = true
		}
		if strings.Contains(notification.Title, "Apply Completed") {
			hasApplyCompleted = true
		}
	}
	
	if !hasResourceFailed {
		t.Error("Expected to receive resource failed notification")
	}
	
	if !hasApplyCompleted {
		t.Error("Expected to receive apply completed notification")
	}
}
