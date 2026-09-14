package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/smatsumae/devops-camp/internal/handlers"
)

func newAuthHandler(t *testing.T) *handlers.AuthHandler {
	return &handlers.AuthHandler{
		DB:        setupTestDB(t),
		JWTSecret: []byte("test-secret"),
		TokenTTL:  time.Hour,
	}
}

func TestRegister_Success(t *testing.T) {
	h := newAuthHandler(t)

	body := []byte(`{"email":"taro@example.com","password":"password123","name":"太郎"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var got map[string]any
	decodeJSON(t, rec, &got)
	if got["email"] != "taro@example.com" {
		t.Errorf("email = %v, want taro@example.com", got["email"])
	}
	if _, ok := got["id"]; !ok {
		t.Error("response missing id")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	h := newAuthHandler(t)

	body := []byte(`{"email":"taro@example.com","password":"password123","name":"太郎"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	h.Register(httptest.NewRecorder(), req)

	req2 := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	rec2 := httptest.NewRecorder()
	h.Register(rec2, req2)

	if rec2.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d, body=%s", rec2.Code, http.StatusUnprocessableEntity, rec2.Body.String())
	}

	var got map[string]any
	decodeJSON(t, rec2, &got)
	errObj := got["error"].(map[string]any)
	if errObj["code"] != "EMAIL_ALREADY_REGISTERED" {
		t.Errorf("error.code = %v, want EMAIL_ALREADY_REGISTERED", errObj["code"])
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	h := newAuthHandler(t)

	body := []byte(`{"email":"taro@example.com","password":"short","name":"太郎"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusUnprocessableEntity, rec.Body.String())
	}
}

func TestLogin_Success(t *testing.T) {
	h := newAuthHandler(t)

	registerBody := []byte(`{"email":"taro@example.com","password":"password123","name":"太郎"}`)
	h.Register(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(registerBody)))

	loginBody := []byte(`{"email":"taro@example.com","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(loginBody))
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got map[string]any
	decodeJSON(t, rec, &got)
	if got["access_token"] == "" || got["access_token"] == nil {
		t.Error("access_token missing from login response")
	}
	if got["token_type"] != "Bearer" {
		t.Errorf("token_type = %v, want Bearer", got["token_type"])
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	h := newAuthHandler(t)

	registerBody := []byte(`{"email":"taro@example.com","password":"password123","name":"太郎"}`)
	h.Register(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(registerBody)))

	loginBody := []byte(`{"email":"taro@example.com","password":"wrong-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(loginBody))
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}
