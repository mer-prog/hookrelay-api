package delivery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSign(t *testing.T) {
	tests := []struct {
		name      string
		payload   []byte
		secret    string
		timestamp string
	}{
		{
			name:      "basic signing",
			payload:   []byte(`{"event":"test"}`),
			secret:    "whsec_test123",
			timestamp: "1700000000",
		},
		{
			name:      "empty payload",
			payload:   []byte{},
			secret:    "secret",
			timestamp: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := Sign(tt.payload, tt.secret, tt.timestamp)
			assert.NotEmpty(t, sig)
			assert.Len(t, sig, 64, "HMAC-SHA256 hex should be 64 chars")
		})
	}
}

func TestVerify(t *testing.T) {
	tests := []struct {
		name      string
		payload   []byte
		secret    string
		timestamp string
		tamper    bool
		want      bool
	}{
		{
			name:      "valid signature",
			payload:   []byte(`{"event":"order.created"}`),
			secret:    "whsec_abc",
			timestamp: "1700000000",
			want:      true,
		},
		{
			name:      "tampered payload",
			payload:   []byte(`{"event":"order.created"}`),
			secret:    "whsec_abc",
			timestamp: "1700000000",
			tamper:    true,
			want:      false,
		},
		{
			name:      "wrong secret",
			payload:   []byte(`{"data":1}`),
			secret:    "wrong",
			timestamp: "123",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := Sign(tt.payload, tt.secret, tt.timestamp)

			verifyPayload := tt.payload
			verifySecret := tt.secret
			if tt.tamper {
				verifyPayload = []byte("tampered")
			}
			if tt.name == "wrong secret" {
				sig = Sign(tt.payload, "different-secret", tt.timestamp)
				// Verify with original secret should fail
				assert.False(t, Verify(verifyPayload, verifySecret, tt.timestamp, sig))
				return
			}
			assert.Equal(t, tt.want, Verify(verifyPayload, verifySecret, tt.timestamp, sig))
		})
	}
}

func TestSignDeterministic(t *testing.T) {
	payload := []byte(`test`)
	sig1 := Sign(payload, "secret", "ts")
	sig2 := Sign(payload, "secret", "ts")
	assert.Equal(t, sig1, sig2, "same input should produce same signature")
}
