package server

import (
	"context"
	"log"
	"net/http"
)

type HTTPServer struct {
	server *http.Server
}

func NewHTTPServer(addr string) *HTTPServer {
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/webhook", withHMACVerification(webhookHandler))

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &HTTPServer{server: srv}
}

func (h *HTTPServer) Start() error {
	log.Printf("Server running at %s\n", h.server.Addr)
	return h.server.ListenAndServe()
}

func (h *HTTPServer) Shutdown(ctx context.Context) error {
	log.Println("Graceful shutdown initiated...")
	return h.server.Shutdown(ctx)
}

// /healthz
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// /webhook
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received"))
}
