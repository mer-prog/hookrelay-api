package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 7 * 24 * time.Hour
)

// Claims represents the JWT payload for access tokens.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GetUserUUID parses the UserID field as a UUID.
func (c *Claims) GetUserUUID() uuid.UUID {
	id, _ := uuid.Parse(c.UserID)
	return id
}

// GenerateAccessToken creates a signed JWT valid for 15 minutes.
func GenerateAccessToken(jwtSecret, userID, email, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Issuer:    "hookrelay",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("GenerateAccessToken: %w", err)
	}
	return signed, nil
}

// GenerateRefreshToken creates a signed JWT valid for 7 days.
func GenerateRefreshToken(jwtSecret, userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(RefreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Issuer:    "hookrelay",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("GenerateRefreshToken: %w", err)
	}
	return signed, nil
}

// ValidateToken parses and validates a JWT string, returning the claims.
func ValidateToken(jwtSecret, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("ValidateToken: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("ValidateToken: invalid token")
	}
	return claims, nil
}
