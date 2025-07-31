package events

import (
	"sync"
)

type EventStore struct {
	mu     sync.RWMutex
	events map[string]Event
}

func NewEventStore() *EventStore {
	return &EventStore{
		events: make(map[string]Event),
	}
}

func (s *EventStore) Save(e Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events[e.ID] = e
}

func (s *EventStore) GetAll() []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	all := make([]Event, 0, len(s.events))
	for _, e := range s.events {
		all = append(all, e)
	}
	return all
}
