package server

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/debrajrout/gosecurehook/internal/security"
)

var webhookSecret = "supersecretkey" // Ideally load from config

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
		// Reset the body so next handler can read it
		r.Body = io.NopCloser(strings.NewReader(string(body)))

		if !security.VerifyHMAC(body, signatureValue, webhookSecret) {
			log.Println("Signature mismatch")
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}
