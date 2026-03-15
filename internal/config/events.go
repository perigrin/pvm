// ABOUTME: Configuration change event system and pub/sub mechanism
// ABOUTME: Provides event-driven notifications for configuration changes

package config

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EventType represents different types of configuration events
type EventType string

const (
	// EventTypeConfigLoaded indicates configuration was successfully loaded
	EventTypeConfigLoaded EventType = "config_loaded"

	// EventTypeConfigReloaded indicates configuration was successfully reloaded
	EventTypeConfigReloaded EventType = "config_reloaded"

	// EventTypeConfigChanged indicates specific configuration values changed
	EventTypeConfigChanged EventType = "config_changed"

	// EventTypeConfigValidated indicates configuration passed validation
	EventTypeConfigValidated EventType = "config_validated"

	// EventTypeConfigValidationFailed indicates configuration validation failed
	EventTypeConfigValidationFailed EventType = "config_validation_failed"

	// EventTypeComponentReconfigured indicates a component was reconfigured
	EventTypeComponentReconfigured EventType = "component_reconfigured"

	// EventTypeComponentReconfigFailed indicates component reconfiguration failed
	EventTypeComponentReconfigFailed EventType = "component_reconfig_failed"

	// EventTypeRollbackStarted indicates configuration rollback has started
	EventTypeRollbackStarted EventType = "rollback_started"

	// EventTypeRollbackCompleted indicates configuration rollback completed
	EventTypeRollbackCompleted EventType = "rollback_completed"
)

// ConfigEvent represents a configuration-related event
type ConfigEvent struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Component string                 `json:"component,omitempty"`
	Config    *Config                `json:"config,omitempty"`
	OldConfig *Config                `json:"old_config,omitempty"`
	Error     error                  `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// EventHandler is a function that handles configuration events
type EventHandler func(event ConfigEvent)

// EventSubscriber manages event subscriptions
type EventSubscriber struct {
	id       string
	handler  EventHandler
	filter   EventFilter
	active   bool
	lastSeen time.Time
}

// EventFilter defines criteria for filtering events
type EventFilter struct {
	Types      []EventType `json:"types,omitempty"`
	Sources    []string    `json:"sources,omitempty"`
	Components []string    `json:"components,omitempty"`
}

// Matches returns true if the event matches the filter criteria
func (f EventFilter) Matches(event ConfigEvent) bool {
	// If no filters are specified, match all events
	if len(f.Types) == 0 && len(f.Sources) == 0 && len(f.Components) == 0 {
		return true
	}

	// Check type filter
	if len(f.Types) > 0 {
		typeMatch := false
		for _, t := range f.Types {
			if t == event.Type {
				typeMatch = true
				break
			}
		}
		if !typeMatch {
			return false
		}
	}

	// Check source filter
	if len(f.Sources) > 0 {
		sourceMatch := false
		for _, s := range f.Sources {
			if s == event.Source {
				sourceMatch = true
				break
			}
		}
		if !sourceMatch {
			return false
		}
	}

	// Check component filter
	if len(f.Components) > 0 {
		if event.Component == "" {
			return false
		}
		componentMatch := false
		for _, c := range f.Components {
			if c == event.Component {
				componentMatch = true
				break
			}
		}
		if !componentMatch {
			return false
		}
	}

	return true
}

// EventBus manages configuration event publishing and subscription
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string]*EventSubscriber
	eventChan   chan ConfigEvent
	stopChan    chan struct{}
	running     bool

	// Configuration
	bufferSize int
	maxRetries int
	retryDelay time.Duration
	deadSubTTL time.Duration
}

// EventBusOptions configures the EventBus behavior
type EventBusOptions struct {
	// BufferSize specifies the size of the event channel buffer
	BufferSize int

	// MaxRetries specifies the maximum number of retry attempts for failed event delivery
	MaxRetries int

	// RetryDelay specifies the delay between retry attempts
	RetryDelay time.Duration

	// DeadSubscriberTTL specifies how long to keep inactive subscribers
	DeadSubscriberTTL time.Duration
}

// DefaultEventBusOptions returns default options for the event bus
func DefaultEventBusOptions() *EventBusOptions {
	return &EventBusOptions{
		BufferSize:        1000,
		MaxRetries:        3,
		RetryDelay:        100 * time.Millisecond,
		DeadSubscriberTTL: 5 * time.Minute,
	}
}

// NewEventBus creates a new configuration event bus
func NewEventBus(opts *EventBusOptions) *EventBus {
	if opts == nil {
		opts = DefaultEventBusOptions()
	}

	return &EventBus{
		subscribers: make(map[string]*EventSubscriber),
		eventChan:   make(chan ConfigEvent, opts.BufferSize),
		stopChan:    make(chan struct{}),
		bufferSize:  opts.BufferSize,
		maxRetries:  opts.MaxRetries,
		retryDelay:  opts.RetryDelay,
		deadSubTTL:  opts.DeadSubscriberTTL,
	}
}

// Start begins the event bus processing
func (eb *EventBus) Start(ctx context.Context) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.running {
		return fmt.Errorf("event bus is already running")
	}

	eb.running = true

	// Start the event processing loop
	go eb.eventLoop(ctx)

	// Start the cleanup routine for dead subscribers
	go eb.cleanupLoop(ctx)

	return nil
}

// Stop stops the event bus
func (eb *EventBus) Stop() error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if !eb.running {
		return nil
	}

	eb.running = false
	close(eb.stopChan)

	return nil
}

// Subscribe registers an event handler with an optional filter
func (eb *EventBus) Subscribe(id string, handler EventHandler, filter EventFilter) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if _, exists := eb.subscribers[id]; exists {
		return fmt.Errorf("subscriber with id %s already exists", id)
	}

	eb.subscribers[id] = &EventSubscriber{
		id:       id,
		handler:  handler,
		filter:   filter,
		active:   true,
		lastSeen: time.Now(),
	}

	return nil
}

// Unsubscribe removes an event handler
func (eb *EventBus) Unsubscribe(id string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	delete(eb.subscribers, id)
}

// Publish sends an event to all matching subscribers
func (eb *EventBus) Publish(event ConfigEvent) {
	// Set timestamp if not already set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	select {
	case eb.eventChan <- event:
	default:
		// Channel is full, drop the event to avoid blocking
		fmt.Printf("EventBus: dropping event due to full buffer\n")
	}
}

// eventLoop processes events and delivers them to subscribers
func (eb *EventBus) eventLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-eb.stopChan:
			return
		case event := <-eb.eventChan:
			eb.deliverEvent(event)
		}
	}
}

// deliverEvent delivers an event to all matching subscribers
func (eb *EventBus) deliverEvent(event ConfigEvent) {
	eb.mu.RLock()
	subscribers := make([]*EventSubscriber, 0)
	for _, sub := range eb.subscribers {
		if sub.active && sub.filter.Matches(event) {
			subscribers = append(subscribers, sub)
		}
	}
	eb.mu.RUnlock()

	// Deliver to all matching subscribers
	for _, sub := range subscribers {
		go eb.deliverToSubscriber(sub, event)
	}
}

// deliverToSubscriber delivers an event to a specific subscriber with retries
func (eb *EventBus) deliverToSubscriber(sub *EventSubscriber, event ConfigEvent) {
	for attempt := 0; attempt <= eb.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(eb.retryDelay)
		}

		// Use a goroutine with timeout to avoid blocking on slow handlers
		done := make(chan bool, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("EventBus: subscriber %s panicked: %v\n", sub.id, r)
				}
				done <- true
			}()

			sub.handler(event)
		}()

		// Wait for completion or timeout
		select {
		case <-done:
			// Success
			eb.mu.Lock()
			sub.lastSeen = time.Now()
			eb.mu.Unlock()
			return
		case <-time.After(5 * time.Second):
			// Timeout
			if attempt == eb.maxRetries {
				fmt.Printf("EventBus: subscriber %s timed out after %d attempts\n",
					sub.id, attempt+1)
				eb.deactivateSubscriber(sub.id)
			}
		}
	}
}

// deactivateSubscriber marks a subscriber as inactive
func (eb *EventBus) deactivateSubscriber(id string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if sub, exists := eb.subscribers[id]; exists {
		sub.active = false
		fmt.Printf("EventBus: deactivated subscriber %s\n", id)
	}
}

// cleanupLoop periodically removes inactive subscribers
func (eb *EventBus) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-eb.stopChan:
			return
		case <-ticker.C:
			eb.cleanupDeadSubscribers()
		}
	}
}

// cleanupDeadSubscribers removes inactive subscribers that have exceeded TTL
func (eb *EventBus) cleanupDeadSubscribers() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	now := time.Now()
	toDelete := make([]string, 0)

	for id, sub := range eb.subscribers {
		if !sub.active && now.Sub(sub.lastSeen) > eb.deadSubTTL {
			toDelete = append(toDelete, id)
		}
	}

	for _, id := range toDelete {
		delete(eb.subscribers, id)
		fmt.Printf("EventBus: removed dead subscriber %s\n", id)
	}
}

// GetSubscriberCount returns the number of active subscribers
func (eb *EventBus) GetSubscriberCount() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	count := 0
	for _, sub := range eb.subscribers {
		if sub.active {
			count++
		}
	}
	return count
}

// SubscriberInfo contains information about a subscriber
type SubscriberInfo struct {
	Active   bool        `json:"active"`
	LastSeen time.Time   `json:"last_seen"`
	Filter   EventFilter `json:"filter"`
}

// GetSubscriberInfo returns information about all subscribers
func (eb *EventBus) GetSubscriberInfo() map[string]SubscriberInfo {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	info := make(map[string]SubscriberInfo)
	for id, sub := range eb.subscribers {
		info[id] = SubscriberInfo{
			Active:   sub.active,
			LastSeen: sub.lastSeen,
			Filter:   sub.filter,
		}
	}
	return info
}

// IsRunning returns whether the event bus is currently running
func (eb *EventBus) IsRunning() bool {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return eb.running
}

// CreateEvent is a helper function to create a new ConfigEvent
func CreateEvent(eventType EventType, source string) ConfigEvent {
	return ConfigEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Source:    source,
		Metadata:  make(map[string]interface{}),
	}
}

// WithComponent adds component information to an event
func (e ConfigEvent) WithComponent(component string) ConfigEvent {
	e.Component = component
	return e
}

// WithConfig adds configuration information to an event
func (e ConfigEvent) WithConfig(config *Config) ConfigEvent {
	e.Config = config
	return e
}

// WithOldConfig adds old configuration information to an event
func (e ConfigEvent) WithOldConfig(oldConfig *Config) ConfigEvent {
	e.OldConfig = oldConfig
	return e
}

// WithError adds error information to an event
func (e ConfigEvent) WithError(err error) ConfigEvent {
	e.Error = err
	return e
}

// WithMetadata adds metadata to an event
func (e ConfigEvent) WithMetadata(key string, value interface{}) ConfigEvent {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}
