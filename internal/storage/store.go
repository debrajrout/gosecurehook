package storage

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/debrajrout/gosecurehook/internal/events"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) SaveEvent(e events.Event) error {
	headers, _ := json.Marshal(e.Headers)
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO events (id, body, headers, received_at)
		VALUES (?, ?, ?, ?)
	`, e.ID, e.Body, string(headers), e.ReceivedAt.Format(time.RFC3339))
	return err
}

func (s *Store) SaveToDLQ(e events.Event) error {
	headers, _ := json.Marshal(e.Headers)
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO dlq (id, body, headers, received_at)
		VALUES (?, ?, ?, ?)
	`, e.ID, e.Body, string(headers), e.ReceivedAt.Format(time.RFC3339))
	return err
}

func (s *Store) ListEvents() ([]events.Event, error) {
	rows, err := s.db.Query(`SELECT id, body, headers, received_at FROM events`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []events.Event
	for rows.Next() {
		var id, body, headersJSON, receivedAtStr string
		if err := rows.Scan(&id, &body, &headersJSON, &receivedAtStr); err != nil {
			continue
		}

		var headers map[string]string
		_ = json.Unmarshal([]byte(headersJSON), &headers)
		parsedTime, _ := time.Parse(time.RFC3339, receivedAtStr)

		results = append(results, events.Event{
			ID:         id,
			Body:       body,
			Headers:    headers,
			ReceivedAt: parsedTime,
		})
	}
	return results, nil
}

func (s *Store) ListDLQ() ([]events.Event, error) {
	rows, err := s.db.Query(`SELECT id, body, headers, received_at FROM dlq`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []events.Event
	for rows.Next() {
		var id, body, headersJSON, receivedAtStr string
		if err := rows.Scan(&id, &body, &headersJSON, &receivedAtStr); err != nil {
			continue
		}

		var headers map[string]string
		_ = json.Unmarshal([]byte(headersJSON), &headers)
		parsedTime, _ := time.Parse(time.RFC3339, receivedAtStr)

		results = append(results, events.Event{
			ID:         id,
			Body:       body,
			Headers:    headers,
			ReceivedAt: parsedTime,
		})
	}
	return results, nil
}

func (s *Store) GetDLQEvent(id string) (*events.Event, error) {
	row := s.db.QueryRow(`SELECT body, headers, received_at FROM dlq WHERE id = ?`, id)

	var body, headersJSON, receivedAtStr string
	if err := row.Scan(&body, &headersJSON, &receivedAtStr); err != nil {
		return nil, err
	}

	var headers map[string]string
	_ = json.Unmarshal([]byte(headersJSON), &headers)
	parsedTime, _ := time.Parse(time.RFC3339, receivedAtStr)

	return &events.Event{
		ID:         id,
		Body:       body,
		Headers:    headers,
		ReceivedAt: parsedTime,
	}, nil
}
