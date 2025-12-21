package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

type UserHandler func(http.ResponseWriter, *http.Request, uuid.UUID)

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
