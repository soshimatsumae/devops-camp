package auth

import (
	"testing"
	"time"
)

func TestGenerateAndParseToken(t *testing.T) {
	secret := []byte("test-secret")

	token, err := GenerateToken(secret, 42, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	userID, err := ParseToken(secret, token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if userID != 42 {
		t.Errorf("ParseToken() userID = %d, want 42", userID)
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, err := GenerateToken([]byte("secret-a"), 1, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if _, err := ParseToken([]byte("secret-b"), token); err == nil {
		t.Error("ParseToken() with wrong secret should fail, got nil error")
	}
}

func TestParseToken_Expired(t *testing.T) {
	secret := []byte("test-secret")
	token, err := GenerateToken(secret, 1, -time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if _, err := ParseToken(secret, token); err == nil {
		t.Error("ParseToken() with expired token should fail, got nil error")
	}
}

func TestParseToken_Garbage(t *testing.T) {
	if _, err := ParseToken([]byte("secret"), "not-a-jwt"); err == nil {
		t.Error("ParseToken() with garbage input should fail, got nil error")
	}
}
