package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequireAuth_MissingHeader(t *testing.T) {
	handler := RequireAuth([]byte("secret"))(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called without a token")
	})

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_ValidToken(t *testing.T) {
	secret := []byte("secret")
	token, err := GenerateToken(secret, 7, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	var gotUserID int64
	handler := RequireAuth(secret)(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, _ = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotUserID != 7 {
		t.Errorf("userID in context = %d, want 7", gotUserID)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	handler := RequireAuth([]byte("secret"))(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called with an invalid token")
	})

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
