// Package validators provides validation logic for authentication and item management requests.
// It ensures data integrity and correctness before processing business logic.
package validators

import "errors"

// ErrEmptyCredentials is returned when login or password is empty during authentication.
var ErrEmptyCredentials = errors.New("login and password cannot be empty")

// AuthValidator handles validation of authentication-related requests.
type AuthValidator struct{}

// NewAuthValidator creates a new instance of AuthValidator.
func NewAuthValidator() *AuthValidator {
	return &AuthValidator{}
}

// ValidateCredentials validates that both login and password are non-empty.
// Returns ErrEmptyCredentials if either field is empty.
func (v *AuthValidator) ValidateCredentials(login, password string) error {
	if login == "" || password == "" {
		return ErrEmptyCredentials
	}
	return nil
}
