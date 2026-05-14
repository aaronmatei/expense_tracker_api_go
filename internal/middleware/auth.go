package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/aaronmatei/expense_tracker_api_go/internal/auth"
)

// contextKey is a private type used as a context key to avoid collisions.
// You should never use a bare string as a context key (other packages could
// accidentally use the same key and clobber each other). The standard pattern
// is: declare an unexported type and unexported keys of that type.
type contextKey string

const userIDKey contextKey = "userID"

// RequireAuth returns middleware that validates the Authorization header.
// On success, it stores the user ID in the request context and calls next.
// On failure, it responds with 401 and short-circuits.
//
// We take the JWT secret as a parameter (instead of reading a global) so the
// middleware stays testable and decoupled from config loading.
func RequireAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				unauthorized(w, "missing Authorization header")
				return
			}

			// Expected format: "Bearer <token>"
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				unauthorized(w, "Authorization header must be 'Bearer <token>'")
				return
			}

			userID, err := auth.ValidateToken(parts[1], jwtSecret)
			if err != nil {
				unauthorized(w, "invalid or expired token")
				return
			}

			// Attach the user ID to the context and call the next handler with
			// the updated request.
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

//incoming request → middleware code (before) → next.ServeHTTP → middleware code (after) → response

// UserIDFromContext extracts the user ID set by RequireAuth.
// Returns (0, false) if no user ID is present — which should only happen if
// the middleware wasn't applied. Treat that as a programming error.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}

func unauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"` + msg + `"}`))
}
