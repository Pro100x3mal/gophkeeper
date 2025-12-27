// Package jwt provides JSON Web Token generation and validation functionality.
//
// This package implements JWT-based authentication using HMAC-SHA256 signing.
// Tokens contain user ID in the subject claim and include standard expiration checks.
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// ErrInvalidToken is returned when a JWT token is invalid or cannot be verified.
var ErrInvalidToken = errors.New("invalid token")

// Generator handles JWT token generation and validation with a configured secret and expiration.
type Generator struct {
	secret     []byte
	expiration time.Duration
}

// NewGenerator creates a new JWT generator with the specified secret and expiration duration.
//
// Parameters:
//   - secret: the secret key used for signing tokens
//   - expiration: how long tokens remain valid after creation
//
// Returns a configured Generator instance.
func NewGenerator(secret string, expiration time.Duration) *Generator {
	return &Generator{
		secret:     []byte(secret),
		expiration: expiration,
	}
}

// GenerateToken creates a new JWT token for the specified user ID.
// The token includes standard claims (subject, issued at, expires at, not before).
//
// Parameters:
//   - userID: the UUID of the user to create a token for
//
// Returns the signed JWT token string or an error if signing fails.
func (g *Generator) GenerateToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(now.Add(g.expiration)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(g.secret)
}

// ValidateToken validates a JWT token string and extracts the user ID.
// Checks the signature, expiration, and extracts the subject claim.
//
// Parameters:
//   - tokenString: the JWT token to validate
//
// Returns the user ID from the token or an error if validation fails.
func (g *Generator) ValidateToken(tokenString string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return g.secret, nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w", err)
	}

	return userID, nil
}
