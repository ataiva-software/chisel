package events

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ataiva-software/forge/pkg/types"
)

func TestEventBus_SubscribeAndPublish(t *testing.T) {
	bus := NewEventBus(10, 2)
	defer bus.Close()
	
	// Create a test handler
	handler := NewLogEventHandler("test-handler", []EventType{EventTypeResourceStarted}, false)
	
	// Subscribe handler
	bus.Subscribe(handler)
	
	// Create and publish event
	event := NewEvent(EventTypeResourceStarted, "test", map[string]interface{}{
		"resource_id": "test.resource",
	})
	
	err := bus.Publish(event)
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}
	
	// Give time for async processing
	time.Sleep(100 * time.Millisecond)
	
	// Verify event was processed (we can't easily test the log output,
	// but we can verify no errors occurred)
}

func TestEventBus_PublishSync(t *testing.T) {
	bus := NewEventBus(10, 2)
	defer bus.Close()
	
	// Create a metrics handler to verify event processing
	handler := NewMetricsEventHandler("metrics", []EventType{EventTypeResourceStarted})
	bus.Subscribe(handler)
	
	// Create and publish event synchronously
	event := NewEvent(EventTypeResourceStarted, "test", map[string]interface{}{
		"resource_id": "test.resource",
	})
	
	ctx := context.Background()
	err := bus.PublishSync(ctx, event)
	if err != nil {
		t.Fatalf("Failed to publish event synchronously: %v", err)
	}
	
	// Verify metrics were updated
	metrics := handler.GetMetrics()
	if metrics[string(EventTypeResourceStarted)] != 1 {
		t.Errorf("Expected 1 resource.started event, got %d", metrics[string(EventTypeResourceStarted)])
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewEventBus(10, 2)
	defer bus.Close()
	
	handler := NewMetricsEventHandler("metrics", []EventType{EventTypeResourceStarted})
	
	// Subscribe and verify
	bus.Subscribe(handler)
	
	event := NewEvent(EventTypeResourceStarted, "test", map[string]interface{}{})
	ctx := context.Background()
	
	err := bus.PublishSync(ctx, event)
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}
	
	metrics := handler.GetMetrics()
	if metrics[string(EventTypeResourceStarted)] != 1 {
		t.Errorf("Expected 1 event before unsubscribe, got %d", metrics[string(EventTypeResourceStarted)])
	}
	
	// Unsubscribe and verify
	bus.Unsubscribe(handler)
	
	err = bus.PublishSync(ctx, event)
	if err != nil {
		t.Fatalf("Failed to publish event after unsubscribe: %v", err)
	}
	
	metrics = handler.GetMetrics()
	if metrics[string(EventTypeResourceStarted)] != 1 {
		t.Errorf("Expected 1 event after unsubscribe (no new events), got %d", metrics[string(EventTypeResourceStarted)])
	}
}

func TestEventBus_BufferFull(t *testing.T) {
	// Create bus with small buffer
	bus := NewEventBus(1, 1)
	defer bus.Close()
	
	// Fill buffer
	event1 := NewEvent(EventTypeResourceStarted, "test", map[string]interface{}{})
	err := bus.Publish(event1)
	if err != nil {
		t.Fatalf("Failed to publish first event: %v", err)
	}
	
	// Try to publish when buffer is full
	event2 := NewEvent(EventTypeResourceStarted, "test", map[string]interface{}{})
	err = bus.Publish(event2)
	if err == nil {
		t.Error("Expected error when buffer is full, but got none")
	}
}

func TestLogEventHandler(t *testing.T) {
	handler := NewLogEventHandler("test-log", []EventType{EventTypeResourceStarted, EventTypeResourceCompleted}, true)
	
	// Test handler properties
	if handler.Name() != "test-log" {
		t.Errorf("Expected handler name 'test-log', got '%s'", handler.Name())
	}
	
	types := handler.Types()
	if len(types) != 2 {
		t.Errorf("Expected 2 event types, got %d", len(types))
	}
	
	// Test event handling
	event := NewEvent(EventTypeResourceStarted, "test", map[string]interface{}{
		"resource_id": "test.resource",
	})
	
	ctx := context.Background()
	err := handler.Handle(ctx, event)
	if err != nil {
		t.Errorf("Unexpected error handling event: %v", err)
	}
}

func TestFileEventHandler(t *testing.T) {
	handler := NewFileEventHandler("test-file", "/tmp/test-events.log", []EventType{EventTypeResourceStarted})
	
	// Test handler properties
	if handler.Name() != "test-file" {
		t.Errorf("Expected handler name 'test-file', got '%s'", handler.Name())
	}
	
	// Test event handling
	event := NewEvent(EventTypeResourceStarted, "test", map[string]interface{}{
		"resource_id": "test.resource",
	})
	
	ctx := context.Background()
	err := handler.Handle(ctx, event)
	if err != nil {
		t.Errorf("Unexpected error handling event: %v", err)
	}
}

func TestMetricsEventHandler(t *testing.T) {
	handler := NewMetricsEventHandler("test-metrics", []EventType{
		EventTypeResourceStarted,
		EventTypeResourceCompleted,
		EventTypeResourceFailed,
	})
	
	ctx := context.Background()
	
	// Handle different types of events
	events := []*Event{
		NewEvent(EventTypeResourceStarted, "executor", map[string]interface{}{
			"resource_id": "test.resource1",
		}),
		NewEvent(EventTypeResourceCompleted, "executor", map[string]interface{}{
			"resource_id": "test.resource1",
			"duration":    100 * time.Millisecond,
		}),
		NewEvent(EventTypeResourceStarted, "executor", map[string]interface{}{
			"resource_id": "test.resource2",
		}),
		NewEvent(EventTypeResourceFailed, "executor", map[string]interface{}{
			"resource_id": "test.resource2",
			"duration":    50 * time.Millisecond,
		}),
	}
	
	for _, event := range events {
		err := handler.Handle(ctx, event)
		if err != nil {
			t.Errorf("Unexpected error handling event: %v", err)
		}
	}
	
	// Verify metrics
	metrics := handler.GetMetrics()
	
	if metrics[string(EventTypeResourceStarted)] != 2 {
		t.Errorf("Expected 2 resource.started events, got %d", metrics[string(EventTypeResourceStarted)])
	}
	
	if metrics[string(EventTypeResourceCompleted)] != 1 {
		t.Errorf("Expected 1 resource.completed event, got %d", metrics[string(EventTypeResourceCompleted)])
	}
	
	if metrics[string(EventTypeResourceFailed)] != 1 {
		t.Errorf("Expected 1 resource.failed event, got %d", metrics[string(EventTypeResourceFailed)])
	}
	
	if metrics["source.executor"] != 4 {
		t.Errorf("Expected 4 events from executor source, got %d", metrics["source.executor"])
	}
	
	// Check duration metrics
	completedDuration := metrics["duration.resource.completed"]
	if completedDuration != 100 {
		t.Errorf("Expected 100ms duration for completed events, got %d", completedDuration)
	}
	
	failedDuration := metrics["duration.resource.failed"]
	if failedDuration != 50 {
		t.Errorf("Expected 50ms duration for failed events, got %d", failedDuration)
	}
}

func TestEventEmitter(t *testing.T) {
	bus := NewEventBus(10, 2)
	defer bus.Close()
	
	// Create metrics handler to verify events
	handler := NewMetricsEventHandler("test-metrics", []EventType{
		EventTypeResourceStarted,
		EventTypeResourceCompleted,
		EventTypeResourceFailed,
		EventTypePlanStarted,
		EventTypePlanCompleted,
		EventTypeApplyStarted,
		EventTypeApplyCompleted,
		EventTypeDriftDetected,
	})
	bus.Subscribe(handler)
	
	emitter := NewEventEmitter(bus, "test-emitter")
	
	// Test resource events
	err := emitter.EmitResourceStarted("test.resource", types.ActionCreate)
	if err != nil {
		t.Errorf("Failed to emit resource started: %v", err)
	}
	
	err = emitter.EmitResourceCompleted("test.resource", types.ActionCreate, 100*time.Millisecond)
	if err != nil {
		t.Errorf("Failed to emit resource completed: %v", err)
	}
	
	err = emitter.EmitResourceFailed("test.resource2", types.ActionUpdate, 
		fmt.Errorf("test error"), 50*time.Millisecond)
	if err != nil {
		t.Errorf("Failed to emit resource failed: %v", err)
	}
	
	// Test plan events
	err = emitter.EmitPlanStarted("test-module", 5)
	if err != nil {
		t.Errorf("Failed to emit plan started: %v", err)
	}
	
	summary := map[string]int{"create": 2, "update": 1}
	err = emitter.EmitPlanCompleted("test-module", summary, 200*time.Millisecond)
	if err != nil {
		t.Errorf("Failed to emit plan completed: %v", err)
	}
	
	// Test apply events
	err = emitter.EmitApplyStarted("test-module", 3)
	if err != nil {
		t.Errorf("Failed to emit apply started: %v", err)
	}
	
	err = emitter.EmitApplyCompleted("test-module", summary, 500*time.Millisecond)
	if err != nil {
		t.Errorf("Failed to emit apply completed: %v", err)
	}
	
	// Test drift event
	changes := map[string]interface{}{
		"content": map[string]interface{}{
			"from": "old",
			"to":   "new",
		},
	}
	err = emitter.EmitDriftDetected("test-module", "test.resource", changes)
	if err != nil {
		t.Errorf("Failed to emit drift detected: %v", err)
	}
	
	// Give time for async processing
	time.Sleep(100 * time.Millisecond)
	
	// Verify all events were processed
	metrics := handler.GetMetrics()
	
	expectedEvents := map[string]int64{
		string(EventTypeResourceStarted):   1,
		string(EventTypeResourceCompleted): 1,
		string(EventTypeResourceFailed):    1,
		string(EventTypePlanStarted):       1,
		string(EventTypePlanCompleted):     1,
		string(EventTypeApplyStarted):      1,
		string(EventTypeApplyCompleted):    1,
		string(EventTypeDriftDetected):     1,
	}
	
	for eventType, expectedCount := range expectedEvents {
		if metrics[eventType] != expectedCount {
			t.Errorf("Expected %d %s events, got %d", expectedCount, eventType, metrics[eventType])
		}
	}
}

func TestEventBus_ConcurrentPublish(t *testing.T) {
	bus := NewEventBus(100, 5)
	defer bus.Close()
	
	handler := NewMetricsEventHandler("concurrent-test", []EventType{EventTypeResourceStarted})
	bus.Subscribe(handler)
	
	// Publish events concurrently
	var wg sync.WaitGroup
	numGoroutines := 10
	eventsPerGoroutine := 10
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < eventsPerGoroutine; j++ {
				event := NewEvent(EventTypeResourceStarted, fmt.Sprintf("source-%d", id), map[string]interface{}{
					"resource_id": fmt.Sprintf("resource-%d-%d", id, j),
				})
				
				err := bus.Publish(event)
				if err != nil {
					t.Errorf("Failed to publish event: %v", err)
				}
			}
		}(i)
	}
	
	wg.Wait()
	
	// Give time for all events to be processed
	time.Sleep(500 * time.Millisecond)
	
	// Verify all events were processed
	metrics := handler.GetMetrics()
	expectedTotal := int64(numGoroutines * eventsPerGoroutine)
	
	if metrics[string(EventTypeResourceStarted)] != expectedTotal {
		t.Errorf("Expected %d events, got %d", expectedTotal, metrics[string(EventTypeResourceStarted)])
	}
}

func TestNewEvent(t *testing.T) {
	data := map[string]interface{}{
		"resource_id": "test.resource",
		"action":      "create",
	}
	
	event := NewEvent(EventTypeResourceStarted, "test-source", data)
	
	// Verify event properties
	if event.Type != EventTypeResourceStarted {
		t.Errorf("Expected event type %s, got %s", EventTypeResourceStarted, event.Type)
	}
	
	if event.Source != "test-source" {
		t.Errorf("Expected source 'test-source', got '%s'", event.Source)
	}
	
	if event.ID == "" {
		t.Error("Expected non-empty event ID")
	}
	
	if event.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
	
	if event.Data["resource_id"] != "test.resource" {
		t.Errorf("Expected resource_id 'test.resource', got '%v'", event.Data["resource_id"])
	}
	
	if event.Tags == nil {
		t.Error("Expected non-nil tags map")
	}
}

func TestEventBus_Close(t *testing.T) {
	bus := NewEventBus(10, 2)
	
	// Publish some events
	for i := 0; i < 5; i++ {
		event := NewEvent(EventTypeResourceStarted, "test", map[string]interface{}{
			"resource_id": fmt.Sprintf("resource-%d", i),
		})
		bus.Publish(event)
	}
	
	// Close the bus
	err := bus.Close()
	if err != nil {
		t.Errorf("Unexpected error closing bus: %v", err)
	}
	
	// Try to publish after close (should fail)
	event := NewEvent(EventTypeResourceStarted, "test", map[string]interface{}{})
	err = bus.Publish(event)
	if err == nil {
		t.Error("Expected error when publishing to closed bus, but got none")
	}
}
