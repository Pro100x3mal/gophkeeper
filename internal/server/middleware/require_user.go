package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

// UserHandler is a handler function that receives a user ID as a parameter.
type UserHandler func(http.ResponseWriter, *http.Request, uuid.UUID)

// RequireUser wraps a UserHandler and extracts the user ID from the request context.
// Returns 401 Unauthorized if no user ID is found in the context.
func RequireUser(next UserHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next(w, r, userID)
	}
}
