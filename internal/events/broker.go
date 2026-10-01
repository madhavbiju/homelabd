package events

import (
	"sync"
)

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Broker struct {
	clients map[chan Event]bool
	mu      sync.RWMutex
}

func NewBroker() *Broker {
	return &Broker{
		clients: make(map[chan Event]bool),
	}
}

func (b *Broker) Subscribe() chan Event {
	ch := make(chan Event, 100) // Buffered to prevent blocking the broker
	b.mu.Lock()
	b.clients[ch] = true
	b.mu.Unlock()
	return ch
}

func (b *Broker) Unsubscribe(ch chan Event) {
	b.mu.Lock()
	if _, ok := b.clients[ch]; ok {
		delete(b.clients, ch)
		close(ch)
	}
	b.mu.Unlock()
}

func (b *Broker) Publish(eventType string, payload interface{}) {
	event := Event{
		Type:    eventType,
		Payload: payload,
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.clients {
		select {
		case ch <- event:
		default:
			// If a client's buffer is full, drop the event to protect the broker.
		}
	}
}
