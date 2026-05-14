package auth

import (
	"fmt"
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	plaintext := "supersecret123"

	hash, err := HashPassword(plaintext)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == "" {
		t.Fatal("hash should not be empty")
	}

	if err := VerifyPassword(hash, plaintext); err != nil {
		fmt.Printf("hash: %s\n", hash)
		t.Errorf("correct password should verify: %v", err)
	}

	if err := VerifyPassword(hash, "wrongpassword"); err == nil {
		t.Error("wrong password should fail verification")
	}
}
