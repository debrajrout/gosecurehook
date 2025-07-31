package events

import "time"

type Event struct {
	ID         string            `json:"id"`
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers"`
	ReceivedAt time.Time         `json:"receivedAt"`
}
