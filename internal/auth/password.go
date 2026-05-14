package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const passwordHashCost = 12

// HashPassword hashes the given password using bcrypt and returns the hash.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), passwordHashCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPasswordHash checks if the given password matches the hash.
func VerifyPassword(password, hash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("password does not match hash: %w", err)
	}
	return nil
}
