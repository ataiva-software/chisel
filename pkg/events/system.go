package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ataiva-software/forge/pkg/types"
)

// EventType represents the type of event
type EventType string

const (
	EventTypeResourceStarted   EventType = "resource.started"
	EventTypeResourceCompleted EventType = "resource.completed"
	EventTypeResourceFailed    EventType = "resource.failed"
	EventTypeResourceSkipped   EventType = "resource.skipped"
	EventTypePlanStarted       EventType = "plan.started"
	EventTypePlanCompleted     EventType = "plan.completed"
	EventTypePlanFailed        EventType = "plan.failed"
	EventTypeApplyStarted      EventType = "apply.started"
	EventTypeApplyCompleted    EventType = "apply.completed"
	EventTypeApplyFailed       EventType = "apply.failed"
	EventTypeDriftDetected     EventType = "drift.detected"
	EventTypeRollbackStarted   EventType = "rollback.started"
	EventTypeRollbackCompleted EventType = "rollback.completed"
)

// Event represents a system event
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Tags      map[string]string      `json:"tags,omitempty"`
}

// EventHandler handles events
type EventHandler interface {
	Handle(ctx context.Context, event *Event) error
	Types() []EventType
	Name() string
}

// EventBus manages event publishing and subscription
type EventBus struct {
	handlers map[EventType][]EventHandler
	mu       sync.RWMutex
	buffer   chan *Event
	workers  int
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	closed   bool
}

// NewEventBus creates a new event bus
func NewEventBus(bufferSize, workers int) *EventBus {
	ctx, cancel := context.WithCancel(context.Background())
	
	bus := &EventBus{
		handlers: make(map[EventType][]EventHandler),
		buffer:   make(chan *Event, bufferSize),
		workers:  workers,
		ctx:      ctx,
		cancel:   cancel,
	}
	
	// Start worker goroutines
	for i := 0; i < workers; i++ {
		bus.wg.Add(1)
		go bus.worker()
	}
	
	return bus
}

// Subscribe subscribes a handler to specific event types
func (b *EventBus) Subscribe(handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	for _, eventType := range handler.Types() {
		b.handlers[eventType] = append(b.handlers[eventType], handler)
	}
}

// Unsubscribe removes a handler from all event types
func (b *EventBus) Unsubscribe(handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	for eventType, handlers := range b.handlers {
		var newHandlers []EventHandler
		for _, h := range handlers {
			if h.Name() != handler.Name() {
				newHandlers = append(newHandlers, h)
			}
		}
		b.handlers[eventType] = newHandlers
	}
}

// Publish publishes an event to the bus
func (b *EventBus) Publish(event *Event) error {
	b.mu.RLock()
	closed := b.closed
	b.mu.RUnlock()
	
	if closed {
		return fmt.Errorf("event bus is closed")
	}
	
	select {
	case b.buffer <- event:
		return nil
	case <-b.ctx.Done():
		return fmt.Errorf("event bus is shutting down")
	default:
		return fmt.Errorf("event buffer is full")
	}
}

// PublishSync publishes an event synchronously
func (b *EventBus) PublishSync(ctx context.Context, event *Event) error {
	b.mu.RLock()
	handlers := b.handlers[event.Type]
	b.mu.RUnlock()
	
	var errors []error
	for _, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			errors = append(errors, fmt.Errorf("handler %s failed: %w", handler.Name(), err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("event handling errors: %v", errors)
	}
	
	return nil
}

// Close shuts down the event bus
func (b *EventBus) Close() error {
	b.mu.Lock()
	b.closed = true
	b.mu.Unlock()
	
	b.cancel()
	close(b.buffer)
	b.wg.Wait()
	return nil
}

// worker processes events from the buffer
func (b *EventBus) worker() {
	defer b.wg.Done()
	
	for {
		select {
		case event, ok := <-b.buffer:
			if !ok {
				return
			}
			
			b.mu.RLock()
			handlers := b.handlers[event.Type]
			b.mu.RUnlock()
			
			for _, handler := range handlers {
				if err := handler.Handle(b.ctx, event); err != nil {
					// Log error but continue processing
					fmt.Printf("Event handler %s failed for event %s: %v\n", handler.Name(), event.ID, err)
				}
			}
			
		case <-b.ctx.Done():
			return
		}
	}
}

// NewEvent creates a new event with a unique ID
func NewEvent(eventType EventType, source string, data map[string]interface{}) *Event {
	return &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Timestamp: time.Now(),
		Source:    source,
		Data:      data,
		Tags:      make(map[string]string),
	}
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("evt_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond()%1000)
}

// LogEventHandler logs events to stdout
type LogEventHandler struct {
	name       string
	eventTypes []EventType
	verbose    bool
}

// NewLogEventHandler creates a new log event handler
func NewLogEventHandler(name string, eventTypes []EventType, verbose bool) *LogEventHandler {
	return &LogEventHandler{
		name:       name,
		eventTypes: eventTypes,
		verbose:    verbose,
	}
}

// Handle handles an event by logging it
func (h *LogEventHandler) Handle(ctx context.Context, event *Event) error {
	if h.verbose {
		data, _ := json.MarshalIndent(event.Data, "  ", "  ")
		fmt.Printf("[%s] %s from %s at %s\n  Data: %s\n",
			event.Type, event.ID, event.Source, event.Timestamp.Format(time.RFC3339), string(data))
	} else {
		fmt.Printf("[%s] %s from %s\n", event.Type, event.ID, event.Source)
	}
	return nil
}

// Types returns the event types this handler subscribes to
func (h *LogEventHandler) Types() []EventType {
	return h.eventTypes
}

// Name returns the handler name
func (h *LogEventHandler) Name() string {
	return h.name
}

// FileEventHandler writes events to a file
type FileEventHandler struct {
	name       string
	eventTypes []EventType
	filePath   string
	mu         sync.Mutex
}

// NewFileEventHandler creates a new file event handler
func NewFileEventHandler(name, filePath string, eventTypes []EventType) *FileEventHandler {
	return &FileEventHandler{
		name:       name,
		eventTypes: eventTypes,
		filePath:   filePath,
	}
}

// Handle handles an event by writing it to a file
func (h *FileEventHandler) Handle(ctx context.Context, event *Event) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// In a real implementation, this would write to a file
	// For now, we'll simulate file writing
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	
	// Simulate file write
	fmt.Printf("Writing to %s: %s\n", h.filePath, string(eventJSON))
	return nil
}

// Types returns the event types this handler subscribes to
func (h *FileEventHandler) Types() []EventType {
	return h.eventTypes
}

// Name returns the handler name
func (h *FileEventHandler) Name() string {
	return h.name
}

// MetricsEventHandler collects metrics from events
type MetricsEventHandler struct {
	name       string
	eventTypes []EventType
	metrics    map[string]int64
	mu         sync.RWMutex
}

// NewMetricsEventHandler creates a new metrics event handler
func NewMetricsEventHandler(name string, eventTypes []EventType) *MetricsEventHandler {
	return &MetricsEventHandler{
		name:       name,
		eventTypes: eventTypes,
		metrics:    make(map[string]int64),
	}
}

// Handle handles an event by updating metrics
func (h *MetricsEventHandler) Handle(ctx context.Context, event *Event) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// Count events by type
	h.metrics[string(event.Type)]++
	
	// Count events by source
	sourceKey := fmt.Sprintf("source.%s", event.Source)
	h.metrics[sourceKey]++
	
	// Extract duration if available
	if duration, ok := event.Data["duration"]; ok {
		if d, ok := duration.(time.Duration); ok {
			durationKey := fmt.Sprintf("duration.%s", event.Type)
			h.metrics[durationKey] += d.Milliseconds()
		}
	}
	
	return nil
}

// Types returns the event types this handler subscribes to
func (h *MetricsEventHandler) Types() []EventType {
	return h.eventTypes
}

// Name returns the handler name
func (h *MetricsEventHandler) Name() string {
	return h.name
}

// GetMetrics returns current metrics
func (h *MetricsEventHandler) GetMetrics() map[string]int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	result := make(map[string]int64)
	for k, v := range h.metrics {
		result[k] = v
	}
	return result
}

// EventEmitter helps emit events during operations
type EventEmitter struct {
	bus    *EventBus
	source string
}

// NewEventEmitter creates a new event emitter
func NewEventEmitter(bus *EventBus, source string) *EventEmitter {
	return &EventEmitter{
		bus:    bus,
		source: source,
	}
}

// EmitResourceStarted emits a resource started event
func (e *EventEmitter) EmitResourceStarted(resourceID string, action types.DiffAction) error {
	event := NewEvent(EventTypeResourceStarted, e.source, map[string]interface{}{
		"resource_id": resourceID,
		"action":      string(action),
	})
	return e.bus.Publish(event)
}

// EmitResourceCompleted emits a resource completed event
func (e *EventEmitter) EmitResourceCompleted(resourceID string, action types.DiffAction, duration time.Duration) error {
	event := NewEvent(EventTypeResourceCompleted, e.source, map[string]interface{}{
		"resource_id": resourceID,
		"action":      string(action),
		"duration":    duration,
		"success":     true,
	})
	return e.bus.Publish(event)
}

// EmitResourceFailed emits a resource failed event
func (e *EventEmitter) EmitResourceFailed(resourceID string, action types.DiffAction, err error, duration time.Duration) error {
	event := NewEvent(EventTypeResourceFailed, e.source, map[string]interface{}{
		"resource_id": resourceID,
		"action":      string(action),
		"error":       err.Error(),
		"duration":    duration,
		"success":     false,
	})
	return e.bus.Publish(event)
}

// EmitPlanStarted emits a plan started event
func (e *EventEmitter) EmitPlanStarted(moduleName string, resourceCount int) error {
	event := NewEvent(EventTypePlanStarted, e.source, map[string]interface{}{
		"module_name":    moduleName,
		"resource_count": resourceCount,
	})
	return e.bus.Publish(event)
}

// EmitPlanCompleted emits a plan completed event
func (e *EventEmitter) EmitPlanCompleted(moduleName string, summary map[string]int, duration time.Duration) error {
	event := NewEvent(EventTypePlanCompleted, e.source, map[string]interface{}{
		"module_name": moduleName,
		"summary":     summary,
		"duration":    duration,
	})
	return e.bus.Publish(event)
}

// EmitApplyStarted emits an apply started event
func (e *EventEmitter) EmitApplyStarted(moduleName string, resourceCount int) error {
	event := NewEvent(EventTypeApplyStarted, e.source, map[string]interface{}{
		"module_name":    moduleName,
		"resource_count": resourceCount,
	})
	return e.bus.Publish(event)
}

// EmitApplyCompleted emits an apply completed event
func (e *EventEmitter) EmitApplyCompleted(moduleName string, summary map[string]int, duration time.Duration) error {
	event := NewEvent(EventTypeApplyCompleted, e.source, map[string]interface{}{
		"module_name": moduleName,
		"summary":     summary,
		"duration":    duration,
	})
	return e.bus.Publish(event)
}

// EmitDriftDetected emits a drift detected event
func (e *EventEmitter) EmitDriftDetected(moduleName string, resourceID string, changes map[string]interface{}) error {
	event := NewEvent(EventTypeDriftDetected, e.source, map[string]interface{}{
		"module_name": moduleName,
		"resource_id": resourceID,
		"changes":     changes,
	})
	return e.bus.Publish(event)
}
