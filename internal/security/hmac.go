package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateHMAC computes HMAC SHA256 digest from the body using the secret.
func GenerateHMAC(body []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyHMAC compares the given signature with the generated digest.
func VerifyHMAC(body []byte, signature, secret string) bool {
	expected := GenerateHMAC(body, secret)
	return hmac.Equal([]byte(expected), []byte(signature))
}
