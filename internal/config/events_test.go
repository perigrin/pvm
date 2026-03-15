// ABOUTME: Tests for configuration event system and pub/sub functionality
// ABOUTME: Comprehensive test coverage for event filtering and delivery

package config

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestEventFilter_Matches(t *testing.T) {
	tests := []struct {
		name     string
		filter   EventFilter
		event    ConfigEvent
		expected bool
	}{
		{
			name:     "empty filter matches all",
			filter:   EventFilter{},
			event:    CreateEvent(EventTypeConfigLoaded, "test"),
			expected: true,
		},
		{
			name: "type filter matches",
			filter: EventFilter{
				Types: []EventType{EventTypeConfigLoaded, EventTypeConfigReloaded},
			},
			event:    CreateEvent(EventTypeConfigLoaded, "test"),
			expected: true,
		},
		{
			name: "type filter doesn't match",
			filter: EventFilter{
				Types: []EventType{EventTypeConfigReloaded},
			},
			event:    CreateEvent(EventTypeConfigLoaded, "test"),
			expected: false,
		},
		{
			name: "source filter matches",
			filter: EventFilter{
				Sources: []string{"test", "other"},
			},
			event:    CreateEvent(EventTypeConfigLoaded, "test"),
			expected: true,
		},
		{
			name: "source filter doesn't match",
			filter: EventFilter{
				Sources: []string{"other"},
			},
			event:    CreateEvent(EventTypeConfigLoaded, "test"),
			expected: false,
		},
		{
			name: "component filter matches",
			filter: EventFilter{
				Components: []string{"comp1", "comp2"},
			},
			event:    CreateEvent(EventTypeComponentReconfigured, "test").WithComponent("comp1"),
			expected: true,
		},
		{
			name: "component filter doesn't match",
			filter: EventFilter{
				Components: []string{"comp2"},
			},
			event:    CreateEvent(EventTypeComponentReconfigured, "test").WithComponent("comp1"),
			expected: false,
		},
		{
			name: "component filter with no component in event",
			filter: EventFilter{
				Components: []string{"comp1"},
			},
			event:    CreateEvent(EventTypeConfigLoaded, "test"),
			expected: false,
		},
		{
			name: "multiple filters all match",
			filter: EventFilter{
				Types:      []EventType{EventTypeConfigLoaded},
				Sources:    []string{"test"},
				Components: []string{"comp1"},
			},
			event:    CreateEvent(EventTypeConfigLoaded, "test").WithComponent("comp1"),
			expected: true,
		},
		{
			name: "multiple filters one doesn't match",
			filter: EventFilter{
				Types:      []EventType{EventTypeConfigLoaded},
				Sources:    []string{"test"},
				Components: []string{"comp2"},
			},
			event:    CreateEvent(EventTypeConfigLoaded, "test").WithComponent("comp1"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.filter.Matches(tt.event)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEventBus_Creation(t *testing.T) {
	opts := DefaultEventBusOptions()
	bus := NewEventBus(opts)

	if bus == nil {
		t.Fatal("Expected non-nil event bus")
	}

	if bus.bufferSize != opts.BufferSize {
		t.Errorf("Expected buffer size %d, got %d", opts.BufferSize, bus.bufferSize)
	}

	if bus.maxRetries != opts.MaxRetries {
		t.Errorf("Expected max retries %d, got %d", opts.MaxRetries, bus.maxRetries)
	}
}

func TestEventBus_StartStop(t *testing.T) {
	bus := NewEventBus(nil)

	if bus.IsRunning() {
		t.Error("Expected bus not to be running initially")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test start
	if err := bus.Start(ctx); err != nil {
		t.Fatalf("Failed to start event bus: %v", err)
	}

	if !bus.IsRunning() {
		t.Error("Expected bus to be running after start")
	}

	// Test double start
	if err := bus.Start(ctx); err == nil {
		t.Error("Expected error for double start")
	}

	// Test stop
	if err := bus.Stop(); err != nil {
		t.Fatalf("Failed to stop event bus: %v", err)
	}

	if bus.IsRunning() {
		t.Error("Expected bus not to be running after stop")
	}

	// Test double stop
	if err := bus.Stop(); err != nil {
		t.Fatalf("Failed to stop already stopped bus: %v", err)
	}
}

func TestEventBus_SubscriptionManagement(t *testing.T) {
	bus := NewEventBus(nil)

	receivedEvents := make([]ConfigEvent, 0)
	var mu sync.Mutex

	handler := func(event ConfigEvent) {
		mu.Lock()
		defer mu.Unlock()
		receivedEvents = append(receivedEvents, event)
	}

	filter := EventFilter{
		Types: []EventType{EventTypeConfigLoaded},
	}

	// Test subscription
	if err := bus.Subscribe("test-sub", handler, filter); err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	if bus.GetSubscriberCount() != 1 {
		t.Errorf("Expected 1 subscriber, got %d", bus.GetSubscriberCount())
	}

	// Test duplicate subscription
	if err := bus.Subscribe("test-sub", handler, filter); err == nil {
		t.Error("Expected error for duplicate subscription")
	}

	// Test unsubscription
	bus.Unsubscribe("test-sub")

	if bus.GetSubscriberCount() != 0 {
		t.Errorf("Expected 0 subscribers after unsubscription, got %d", bus.GetSubscriberCount())
	}

	// Test unsubscribing non-existent subscriber (should not panic)
	bus.Unsubscribe("non-existent")
}

func TestEventBus_EventDelivery(t *testing.T) {
	bus := NewEventBus(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := bus.Start(ctx); err != nil {
		t.Fatalf("Failed to start event bus: %v", err)
	}
	defer bus.Stop()

	receivedEvents := make([]ConfigEvent, 0)
	var mu sync.Mutex

	handler := func(event ConfigEvent) {
		mu.Lock()
		defer mu.Unlock()
		receivedEvents = append(receivedEvents, event)
	}

	filter := EventFilter{
		Types: []EventType{EventTypeConfigLoaded, EventTypeConfigReloaded},
	}

	if err := bus.Subscribe("test-sub", handler, filter); err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish matching events
	event1 := CreateEvent(EventTypeConfigLoaded, "test-source")
	event2 := CreateEvent(EventTypeConfigReloaded, "test-source")
	event3 := CreateEvent(EventTypeConfigChanged, "test-source") // Should not match filter

	bus.Publish(event1)
	bus.Publish(event2)
	bus.Publish(event3)

	// Wait for event delivery
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(receivedEvents) != 2 {
		t.Errorf("Expected 2 events delivered, got %d", len(receivedEvents))
		return
	}

	// Events may arrive in any order due to goroutines
	eventTypes := make(map[EventType]bool)
	for _, event := range receivedEvents {
		eventTypes[event.Type] = true
	}

	if !eventTypes[EventTypeConfigLoaded] {
		t.Errorf("Expected to receive EventTypeConfigLoaded")
	}

	if !eventTypes[EventTypeConfigReloaded] {
		t.Errorf("Expected to receive EventTypeConfigReloaded")
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	bus := NewEventBus(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := bus.Start(ctx); err != nil {
		t.Fatalf("Failed to start event bus: %v", err)
	}
	defer bus.Stop()

	var receivedCount1, receivedCount2 int
	var mu1, mu2 sync.Mutex

	handler1 := func(event ConfigEvent) {
		mu1.Lock()
		defer mu1.Unlock()
		receivedCount1++
	}

	handler2 := func(event ConfigEvent) {
		mu2.Lock()
		defer mu2.Unlock()
		receivedCount2++
	}

	filter1 := EventFilter{Types: []EventType{EventTypeConfigLoaded}}
	filter2 := EventFilter{Types: []EventType{EventTypeConfigLoaded, EventTypeConfigReloaded}}

	if err := bus.Subscribe("sub1", handler1, filter1); err != nil {
		t.Fatalf("Failed to subscribe sub1: %v", err)
	}

	if err := bus.Subscribe("sub2", handler2, filter2); err != nil {
		t.Fatalf("Failed to subscribe sub2: %v", err)
	}

	if bus.GetSubscriberCount() != 2 {
		t.Errorf("Expected 2 subscribers, got %d", bus.GetSubscriberCount())
	}

	// Publish events
	bus.Publish(CreateEvent(EventTypeConfigLoaded, "test"))   // Both should receive
	bus.Publish(CreateEvent(EventTypeConfigReloaded, "test")) // Only sub2 should receive

	// Wait for event delivery
	time.Sleep(100 * time.Millisecond)

	mu1.Lock()
	count1 := receivedCount1
	mu1.Unlock()

	mu2.Lock()
	count2 := receivedCount2
	mu2.Unlock()

	if count1 != 1 {
		t.Errorf("Expected sub1 to receive 1 event, got %d", count1)
	}

	if count2 != 2 {
		t.Errorf("Expected sub2 to receive 2 events, got %d", count2)
	}
}

func TestEventBus_SubscriberInfo(t *testing.T) {
	bus := NewEventBus(nil)

	handler := func(event ConfigEvent) {}
	filter := EventFilter{Types: []EventType{EventTypeConfigLoaded}}

	if err := bus.Subscribe("test-sub", handler, filter); err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	info := bus.GetSubscriberInfo()
	if len(info) != 1 {
		t.Errorf("Expected 1 subscriber info, got %d", len(info))
	}

	subInfo, exists := info["test-sub"]
	if !exists {
		t.Error("Expected subscriber info for 'test-sub'")
		return
	}

	if !subInfo.Active {
		t.Error("Expected subscriber to be active")
	}

	if subInfo.LastSeen.IsZero() {
		t.Error("Expected non-zero last_seen in subscriber info")
	}

	if len(subInfo.Filter.Types) != 1 || subInfo.Filter.Types[0] != EventTypeConfigLoaded {
		t.Error("Expected filter in subscriber info to match subscription")
	}
}

func TestEventBus_HandlerPanic(t *testing.T) {
	bus := NewEventBus(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := bus.Start(ctx); err != nil {
		t.Fatalf("Failed to start event bus: %v", err)
	}
	defer bus.Stop()

	panicHandler := func(event ConfigEvent) {
		panic("test panic")
	}

	if err := bus.Subscribe("panic-sub", panicHandler, EventFilter{}); err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish event that will cause panic
	bus.Publish(CreateEvent(EventTypeConfigLoaded, "test"))

	// Wait for event processing
	time.Sleep(100 * time.Millisecond)

	// Bus should still be running and subscriber should be marked active (for now)
	if !bus.IsRunning() {
		t.Error("Expected bus to still be running after handler panic")
	}

	if bus.GetSubscriberCount() != 1 {
		t.Errorf("Expected 1 subscriber after panic, got %d", bus.GetSubscriberCount())
	}
}

func TestCreateEvent_Helpers(t *testing.T) {
	event := CreateEvent(EventTypeConfigLoaded, "test-source")

	if event.Type != EventTypeConfigLoaded {
		t.Errorf("Expected event type %v, got %v", EventTypeConfigLoaded, event.Type)
	}

	if event.Source != "test-source" {
		t.Errorf("Expected source 'test-source', got %s", event.Source)
	}

	if event.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}

	if event.Metadata == nil {
		t.Error("Expected non-nil metadata")
	}

	// Test fluent methods
	config := NewDefaultConfig()
	oldConfig := NewDefaultConfig()
	testError := &ValidationError{Message: "test error"}

	event = event.
		WithComponent("test-component").
		WithConfig(config).
		WithOldConfig(oldConfig).
		WithError(testError).
		WithMetadata("key", "value")

	if event.Component != "test-component" {
		t.Errorf("Expected component 'test-component', got %s", event.Component)
	}

	if event.Config != config {
		t.Error("Expected config to be set")
	}

	if event.OldConfig != oldConfig {
		t.Error("Expected old config to be set")
	}

	if event.Error != testError {
		t.Error("Expected error to be set")
	}

	if event.Metadata["key"] != "value" {
		t.Errorf("Expected metadata key 'value', got %v", event.Metadata["key"])
	}
}

// ValidationError for testing
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
