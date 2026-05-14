// Package server wires the HTTP layer: it composes middleware, builds the
// users feature stack (repository → service → handler), and mounts the routes
// returned to main as an http.Handler.
package server

import (
	"encoding/json"
	"net/http"
	"time"

	chimw "github.com/aaronmatei/expense_tracker_api_go/internal/middleware"
	"github.com/aaronmatei/expense_tracker_api_go/internal/users"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds runtime settings sourced from env/flags by the caller.
// JWTSecret signs and verifies access tokens; JWTExpiry sets their lifetime;
// CORSOrigins is the allow-list of browser origins permitted to call the API.
type Config struct {
	JWTSecret   string
	JWTExpiry   time.Duration
	CORSOrigins []string
}

// New constructs the root router with all middleware and routes registered.
// The returned http.Handler is ready to pass to http.Server.
func New(pool *pgxpool.Pool, cfg Config) http.Handler {
	r := chi.NewRouter()

	// Cross-cutting middleware applied to every request.
	// Order matters: Recoverer must be outermost so it catches panics from
	// anything below it; Logger sits after RequestID/RealIP so log lines
	// carry the correlation ID and true client IP.
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300, // cache preflight for 5 minutes
	}))

	// Build the users feature: repo talks to Postgres, service holds the
	// business logic (register/login/JWT issuance), handler exposes HTTP.
	userRepo := users.NewRepository(pool)
	userService := users.NewService(userRepo, cfg.JWTSecret, cfg.JWTExpiry)
	userHandler := users.NewHandler(userService)

	// Public routes — no auth required.
	r.Get("/health", healthHandler)
	r.Mount("/auth", userHandler.Routes())

	// Protected routes — require a valid bearer token. RequireAuth verifies
	// the JWT and injects the user ID into the request context.
	r.Group(func(r chi.Router) {
		r.Use(chimw.RequireAuth(cfg.JWTSecret))
		r.Get("/me", meHandler(userService))
	})

	return r
}

// healthHandler is a liveness probe for load balancers and uptime checks.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// meHandler returns the authenticated user's profile. The user ID is read
// from the context (populated by RequireAuth); a missing ID here means the
// middleware chain is misconfigured, hence the 500 rather than 401.
func meHandler(svc *users.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := chimw.UserIDFromContext(r.Context())
		if !ok {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "user ID missing from context"})
			return
		}

		user, err := svc.GetByID(r.Context(), userID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"created_at": user.CreatedAt,
		})
	}
}

// writeJSON serializes v as JSON with the given status code. The encode
// error is intentionally ignored: headers are already flushed by WriteHeader,
// so there is no way to surface a different status to the client.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
