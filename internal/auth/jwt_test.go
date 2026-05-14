package auth

import (
	"testing"
	"time"
)

func TestGenerateAndValidateToken(t *testing.T) {
	secret := "test-secret-do-not-use-in-prod"
	userID := int64(42)

	token, err := GenerateToken(userID, secret, 1*time.Hour)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}

	gotID, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if gotID != userID {
		t.Errorf("got user ID %d, want %d", gotID, userID)
	}
}

func TestValidateTokenWithWrongSecret(t *testing.T) {
	token, err := GenerateToken(1, "right-secret", time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ValidateToken(token, "wrong-secret"); err == nil {
		t.Error("validation should fail with wrong secret")
	}
}

func TestValidateExpiredToken(t *testing.T) {
	// Negative duration = already expired
	token, err := GenerateToken(1, "secret", -1*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ValidateToken(token, "secret"); err == nil {
		t.Error("validation should fail for expired token")
	}
}

func TestValidateMalformedToken(t *testing.T) {
	if _, err := ValidateToken("not-a-real-jwt", "secret"); err == nil {
		t.Error("validation should fail for malformed token")
	}
}
