// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	// Note: sync.RWMutex protects dispatcher handlers and has no core equivalent.
	"sync"
	// Note: time.Time is retained for event timestamps in the public event struct.
	"time"
)

const (
	EventStart    = "start"
	EventProgress = "progress"
	EventItem     = "item"
	EventError    = "error"
	EventComplete = "complete"
)

// Event represents a collection event.
type Event struct {
	Type    string    `json:"type"`
	Source  string    `json:"source"`
	Message string    `json:"message"`
	Data    any       `json:"data,omitempty"`
	Time    time.Time `json:"time"`
}

// EventHandler handles collection events.
type EventHandler func(Event)

// Dispatcher manages event dispatch.
type Dispatcher struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}

// NewDispatcher creates a new event dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{handlers: make(map[string][]EventHandler)}
}

// On registers a handler for an event type.
func (d *Dispatcher) On(eventType string, handler EventHandler) {
	if d == nil || handler == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.handlers == nil {
		d.handlers = make(map[string][]EventHandler)
	}
	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

// Emit dispatches an event to all registered handlers for that event type.
func (d *Dispatcher) Emit(event Event) {
	if d == nil {
		return
	}
	if event.Time.IsZero() {
		event.Time = time.Now().UTC()
	}
	d.mu.RLock()
	handlers := append([]EventHandler(nil), d.handlers[event.Type]...)
	d.mu.RUnlock()
	for _, handler := range handlers {
		handler(event)
	}
}

func (d *Dispatcher) EmitStart(source, message string) {
	d.Emit(Event{Type: EventStart, Source: source, Message: message})
}
func (d *Dispatcher) EmitProgress(source, message string, data any) {
	d.Emit(Event{Type: EventProgress, Source: source, Message: message, Data: data})
}
func (d *Dispatcher) EmitItem(source, message string, data any) {
	d.Emit(Event{Type: EventItem, Source: source, Message: message, Data: data})
}
func (d *Dispatcher) EmitError(source, message string, data any) {
	d.Emit(Event{Type: EventError, Source: source, Message: message, Data: data})
}
func (d *Dispatcher) EmitComplete(source, message string, data any) {
	d.Emit(Event{Type: EventComplete, Source: source, Message: message, Data: data})
}
