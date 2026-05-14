package users

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aaronmatei/expense_tracker_api_go/internal/auth"
)

// ErrInvalidCredentials is returned for any login failure. We intentionally
// don't distinguish "user not found" from "wrong password" — that would let
// attackers enumerate which emails are registered.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Service holds the dependencies the user-facing operations need.
// We inject them so tests can swap real implementations for fakes.
type Service struct {
	repo      *Repository
	jwtSecret string
	jwtExpiry time.Duration
}

// NewService constructs a Service with its dependencies.
func NewService(repo *Repository, jwtSecret string, jwtExpiry time.Duration) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

// RegisterInput is the data needed to create a new user.
// Keeping inputs as named structs (vs many parameters) makes future fields easy.
type RegisterInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// AuthResult is what both Register and Login return: the user and a fresh token.
type AuthResult struct {
	User  *User
	Token string
}

// Register creates a new user account and returns an auth token.
func (s *Service) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	// Normalize email: lowercase + trim. Prevents duplicate accounts that differ
	// only in case ("Alice@x.com" vs "alice@x.com").
	email := strings.ToLower(strings.TrimSpace(in.Email))

	// Basic validation. We'll add stricter validation (length, format) at the
	// handler layer with the validator library; this is the service's safety net.
	if email == "" || in.Password == "" || in.FirstName == "" || in.LastName == "" {
		return nil, fmt.Errorf("all fields are required")
	}
	if len(in.Password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}

	user, err := s.repo.Create(ctx, email, hash, strings.TrimSpace(in.FirstName), strings.TrimSpace(in.LastName))
	if err != nil {
		// Repository's ErrDuplicateEmail bubbles up unchanged for the handler to map to a 409.
		return nil, err
	}

	token, err := auth.GenerateToken(user.ID, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}

	return &AuthResult{User: user, Token: token}, nil
}

// Login verifies credentials and returns an auth token on success.
func (s *Service) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		// "User not found" and "wrong password" both become ErrInvalidCredentials.
		// The original error is dropped here on purpose; if you want it in logs,
		// log it inside this branch before returning.
		if errors.Is(err, ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("login: %w", err)
	}

	if err := auth.VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := auth.GenerateToken(user.ID, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}

	return &AuthResult{User: user, Token: token}, nil
}

// GetByID is a thin pass-through used by protected handlers to look up "me".
// Kept on the service (not exposing the repo directly) so business rules can
// be added later without changing handler code.
func (s *Service) GetByID(ctx context.Context, id int64) (*User, error) {
	return s.repo.GetByID(ctx, id)
}
