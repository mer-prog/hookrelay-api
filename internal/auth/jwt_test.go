package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-jwt-secret"

func TestGenerateAndValidateAccessToken(t *testing.T) {
	token, err := GenerateAccessToken(testSecret, "user-123", "test@example.com", "developer")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateToken(testSecret, token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "developer", claims.Role)
	assert.Equal(t, "hookrelay", claims.Issuer)
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken(testSecret, "user-456")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateToken(testSecret, token)
	require.NoError(t, err)
	assert.Equal(t, "user-456", claims.UserID)
	assert.Empty(t, claims.Email)
	assert.Empty(t, claims.Role)
}

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name    string
		token   func() string
		wantErr string
	}{
		{
			name: "expired token",
			token: func() string {
				claims := Claims{
					UserID: "user-1",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(-1 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now().UTC().Add(-2 * time.Hour)),
					},
				}
				t, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
				return t
			},
			wantErr: "token is expired",
		},
		{
			name: "wrong secret",
			token: func() string {
				t, _ := GenerateAccessToken("wrong-secret", "user-1", "a@b.com", "dev")
				return t
			},
			wantErr: "signature is invalid",
		},
		{
			name: "malformed token",
			token: func() string {
				return "not.a.valid.jwt"
			},
			wantErr: "ValidateToken",
		},
		{
			name: "empty token",
			token: func() string {
				return ""
			},
			wantErr: "ValidateToken",
		},
		{
			name: "wrong signing method",
			token: func() string {
				claims := Claims{UserID: "user-1"}
				// Sign with none method
				t := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
				s, _ := t.SignedString(jwt.UnsafeAllowNoneSignatureType)
				return s
			},
			wantErr: "unexpected signing method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(testSecret, tt.token())
			require.Error(t, err)
			assert.Nil(t, claims)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}
