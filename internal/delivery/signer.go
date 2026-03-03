package delivery

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Sign computes an HMAC-SHA256 signature over "timestamp.payload" using the given secret.
func Sign(payload []byte, secret, timestamp string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%s.%s", timestamp, payload)))
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify checks that the given signature matches the expected HMAC-SHA256.
func Verify(payload []byte, secret, timestamp, signature string) bool {
	expected := Sign(payload, secret, timestamp)
	return hmac.Equal([]byte(expected), []byte(signature))
}
