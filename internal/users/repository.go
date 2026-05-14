package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("User not found")
var ErrDuplicateEmail = errors.New("Email already exists")

// Repository wraps the database pool and exposes user-related database operations
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new Repository with the given database pool
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new user into the database and returns the created user with its ID
// Uses RETURNING to get the generated ID and timestamps in one query
func (r *Repository) Create(ctx context.Context, email, passwordHash, firstName, lastName string) (*User, error) {
	const query = `
		INSERT INTO users (email, password_hash, first_name, last_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, password_hash, first_name, last_name, created_at, updated_at
	`

	var u User
	err := r.pool.QueryRow(ctx, query, email, passwordHash, firstName, lastName).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateEmail
		}
		return nil, fmt.Errorf("create user: %w", err)

	}

	return &u, nil
}

// GetByEmail looks up a User by email. Returns ErrNotFound if no row exists
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	const query = `
		SELECT id, email, password_hash, first_name, last_name, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var u User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &u, nil

}

// GetByID looks up a user by ID. Returns ErrNotFound if no row exists.
func (r *Repository) GetByID(ctx context.Context, id int64) (*User, error) {
	const query = `
		SELECT id, email, password_hash, first_name, last_name, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var u User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

// isUniqueViolation checks if the error is a unique constraint violation (e.g., duplicate email)
func isUniqueViolation(err error) bool {
	var pgErr interface {
		SQLSlate() string
	}
	if errors.As(err, &pgErr) {
		return pgErr.SQLSlate() == "23505"
	}
	return false
}
