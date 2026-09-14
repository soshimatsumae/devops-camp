package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/smatsumae/devops-camp/internal/auth"
	"github.com/smatsumae/devops-camp/internal/httpx"
)

type AuthHandler struct {
	DB        *sql.DB
	JWTSecret []byte
	TokenTTL  time.Duration
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type registerResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "request body is not valid JSON")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Name = strings.TrimSpace(req.Name)

	if req.Email == "" || req.Name == "" {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "email and name are required")
		return
	}
	if len(req.Password) < 8 {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "password must be at least 8 characters")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to hash password")
		return
	}

	res, err := h.DB.ExecContext(r.Context(),
		`INSERT INTO users (email, password_hash, name) VALUES (?, ?, ?)`,
		req.Email, hash, req.Name,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "EMAIL_ALREADY_REGISTERED", "email already registered")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create user")
		return
	}

	id, _ := res.LastInsertId()

	var createdAt time.Time
	err = h.DB.QueryRowContext(r.Context(), `SELECT created_at FROM users WHERE id = ?`, id).Scan(&createdAt)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load created user")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, registerResponse{
		ID:        id,
		Email:     req.Email,
		Name:      req.Name,
		CreatedAt: createdAt,
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "request body is not valid JSON")
		return
	}

	var (
		userID int64
		hash   string
	)
	err := h.DB.QueryRowContext(r.Context(),
		`SELECT id, password_hash FROM users WHERE email = ?`, req.Email,
	).Scan(&userID, &hash)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !auth.CheckPassword(hash, req.Password)) {
		httpx.WriteError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
		return
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to look up user")
		return
	}

	token, err := auth.GenerateToken(h.JWTSecret, userID, h.TokenTTL)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to generate token")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, loginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(h.TokenTTL.Seconds()),
	})
}
