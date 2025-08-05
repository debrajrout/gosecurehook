package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/debrajrout/gosecurehook/internal/security"
)

var (
	webhookSecret = "supersecretkey"
	requestCount  = 0

	rateLimitWindow = 1 * time.Minute
	rateLimitMax    = 10
	rateLimitStore  = make(map[string][]time.Time)
	rateLimitMutex  sync.Mutex
)

// ───── HMAC Signature Middleware ─────
func withHMACVerification(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		signature := r.Header.Get("X-Signature")
		if signature == "" || !strings.HasPrefix(signature, "sha256=") {
			http.Error(w, "Missing or invalid signature", http.StatusUnauthorized)
			return
		}

		signatureValue := strings.TrimPrefix(signature, "sha256=")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(strings.NewReader(string(body)))

		if !security.VerifyHMAC(body, signatureValue, webhookSecret) {
			log.Println("Signature mismatch")
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// ───── Logger Middleware ─────
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		log.Printf("[HTTP] %s %s (%v)", r.Method, r.URL.Path, duration)
		requestCount++
	})
}

// ───── Panic Recovery Middleware ─────
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC RECOVERED] %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ───── Rate Limiting Middleware ─────
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)

		rateLimitMutex.Lock()
		defer rateLimitMutex.Unlock()

		now := time.Now()
		windowStart := now.Add(-rateLimitWindow)

		requests := rateLimitStore[ip]
		filtered := make([]time.Time, 0, len(requests))
		for _, t := range requests {
			if t.After(windowStart) {
				filtered = append(filtered, t)
			}
		}

		if len(filtered) >= rateLimitMax {
			log.Printf("[RATE LIMIT] IP %s exceeded limit", ip)
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		rateLimitStore[ip] = append(filtered, now)
		next.ServeHTTP(w, r)
	})
}

func getIP(r *http.Request) string {
	// Check for X-Forwarded-For (if behind proxy)
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[0])
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// ───── Prometheus Metrics ─────
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("# HELP gosecurehook_http_requests_total The total number of HTTP requests\n"))
	w.Write([]byte("# TYPE gosecurehook_http_requests_total counter\n"))
	w.Write([]byte("gosecurehook_http_requests_total " + fmt.Sprintf("%d", requestCount) + "\n"))
}
