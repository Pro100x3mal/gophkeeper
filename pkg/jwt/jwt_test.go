package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenerator(t *testing.T) {
	secret := "test_secret"
	expiration := time.Hour

	gen := NewGenerator(secret, expiration)

	require.NotNil(t, gen)
	assert.Equal(t, []byte(secret), gen.secret)
	assert.Equal(t, expiration, gen.expiration)
}

func TestGenerateToken_Success(t *testing.T) {
	gen := NewGenerator("test_secret", time.Hour)
	userID := uuid.New()

	token, err := gen.GenerateToken(userID)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateToken_ValidFormat(t *testing.T) {
	gen := NewGenerator("test_secret", time.Hour)
	userID := uuid.New()

	tokenString, err := gen.GenerateToken(userID)
	require.NoError(t, err)

	// Parse token to verify format
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test_secret"), nil
	})

	require.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	require.True(t, ok)
	assert.Equal(t, userID.String(), claims.Subject)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.NotBefore)
}

func TestValidateToken_Success(t *testing.T) {
	gen := NewGenerator("test_secret", time.Hour)
	userID := uuid.New()

	tokenString, err := gen.GenerateToken(userID)
	require.NoError(t, err)

	validatedUserID, err := gen.ValidateToken(tokenString)

	require.NoError(t, err)
	assert.Equal(t, userID, validatedUserID)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	gen := NewGenerator("test_secret", time.Hour)

	tests := []struct {
		name  string
		token string
	}{
		{"Empty token", ""},
		{"Invalid format", "invalid.token.format"},
		{"Random string", "random_string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := gen.ValidateToken(tt.token)
			assert.Error(t, err)
			assert.Equal(t, uuid.Nil, userID)
		})
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	gen1 := NewGenerator("secret1", time.Hour)
	gen2 := NewGenerator("secret2", time.Hour)
	userID := uuid.New()

	tokenString, err := gen1.GenerateToken(userID)
	require.NoError(t, err)

	validatedUserID, err := gen2.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, validatedUserID)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	gen := NewGenerator("test_secret", 1*time.Millisecond)
	userID := uuid.New()

	tokenString, err := gen.GenerateToken(userID)
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	validatedUserID, err := gen.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, validatedUserID)
}

func TestValidateToken_InvalidUserID(t *testing.T) {
	gen := NewGenerator("test_secret", time.Hour)

	// Create token with invalid UUID in subject
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   "invalid-uuid",
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test_secret"))
	require.NoError(t, err)

	validatedUserID, err := gen.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user ID")
	assert.Equal(t, uuid.Nil, validatedUserID)
}

func TestValidateToken_WrongSigningMethod(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}

	// Use different signing method (RS256 instead of HS256)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	gen := NewGenerator("test_secret", time.Hour)
	validatedUserID, err := gen.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, validatedUserID)
}

func TestErrInvalidToken(t *testing.T) {
	assert.Equal(t, "invalid token", ErrInvalidToken.Error())
}

func TestGenerateAndValidate_MultipleTokens(t *testing.T) {
	gen := NewGenerator("test_secret", time.Hour)

	userIDs := []uuid.UUID{
		uuid.New(),
		uuid.New(),
		uuid.New(),
	}

	for _, userID := range userIDs {
		tokenString, err := gen.GenerateToken(userID)
		require.NoError(t, err)

		validatedUserID, err := gen.ValidateToken(tokenString)
		require.NoError(t, err)
		assert.Equal(t, userID, validatedUserID)
	}
}
