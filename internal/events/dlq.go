package events

import (
	"sync"
)

type DLQ struct {
	mu     sync.RWMutex
	events map[string]Event
}

func NewDLQ() *DLQ {
	return &DLQ{
		events: make(map[string]Event),
	}
}

func (d *DLQ) Save(event Event) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.events[event.ID] = event
}

func (d *DLQ) Get(id string) (Event, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	e, ok := d.events[id]
	return e, ok
}

func (d *DLQ) GetAll() []Event {
	d.mu.RLock()
	defer d.mu.RUnlock()
	list := make([]Event, 0, len(d.events))
	for _, e := range d.events {
		list = append(list, e)
	}
	return list
}
