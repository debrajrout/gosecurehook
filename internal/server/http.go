package server

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/debrajrout/gosecurehook/internal/events"
	"github.com/debrajrout/gosecurehook/internal/storage"
	"github.com/google/uuid"
)

type HTTPServer struct {
	server *http.Server
	store  *storage.Store
}

type HTTPServer2 struct {
	postman *http.Server
	android *http.Server
}

func NewHTTPServer(addr string, store *storage.Store) *HTTPServer {
	s := &HTTPServer{store: store}

	mux := http.NewServeMux()

	// Register all endpoints
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/webhook", withHMACVerification(s.webhookHandler))
	mux.HandleFunc("/admin/events", s.listEventsHandler)
	mux.HandleFunc("/admin/dlq", s.listDLQHandler)
	mux.HandleFunc("/admin/replay/", s.replayHandler)
	mux.HandleFunc("/metrics", MetricsHandler)

	// Apply middleware chain
	chain := LoggerMiddleware(
		RecoveryMiddleware(
			RateLimitMiddleware(mux),
		),
	)

	s.server = &http.Server{
		Addr:    addr,
		Handler: chain,
	}

	return s
}

func (h *HTTPServer) Start() error {
	log.Printf("Server running at %s\n", h.server.Addr)
	return h.server.ListenAndServe()
}

func (h *HTTPServer) Shutdown(ctx context.Context) error {
	log.Println("Graceful shutdown initiated...")
	return h.server.Shutdown(ctx)
}

// ────────────── Handlers ──────────────

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *HTTPServer) webhookHandler(w http.ResponseWriter, r *http.Request) {
	bodyBytes, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	event := events.Event{
		ID:         uuid.NewString(),
		Body:       string(bodyBytes),
		Headers:    extractHeaders(r),
		ReceivedAt: time.Now(),
	}

	// Simulate failure if body contains `"fail": true`
	if strings.Contains(event.Body, `"fail": true`) {
		if err := h.store.SaveToDLQ(event); err != nil {
			http.Error(w, "Failed to store in DLQ", http.StatusInternalServerError)
			return
		}
		log.Printf("Event %s failed and moved to DLQ\n", event.ID)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Simulated failure. Moved to DLQ with ID: " + event.ID))
		return
	}

	if err := h.store.SaveEvent(event); err != nil {
		http.Error(w, "Failed to store event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook stored with ID: " + event.ID))
}

func (h *HTTPServer) listEventsHandler(w http.ResponseWriter, r *http.Request) {
	allEvents, err := h.store.ListEvents()
	if err != nil {
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allEvents)
}

func (h *HTTPServer) listDLQHandler(w http.ResponseWriter, r *http.Request) {
	all, err := h.store.ListDLQ()
	if err != nil {
		http.Error(w, "Failed to fetch DLQ", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(all)
}

func (h *HTTPServer) replayHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/admin/replay/")
	event, err := h.store.GetDLQEvent(id)
	if err != nil {
		http.Error(w, "Event not found in DLQ", http.StatusNotFound)
		return
	}

	if err := h.store.SaveEvent(*event); err != nil {
		http.Error(w, "Failed to replay event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Replayed event with ID: " + id))
}

// ────────────── Utility ──────────────

func extractHeaders(r *http.Request) map[string]string {
	headers := make(map[string]string)
	for k, v := range r.Header {
		headers[k] = v[0]
	}
	return headers
}
