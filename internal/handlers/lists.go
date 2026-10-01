package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/smatsumae/devops-camp/internal/httpx"
	"github.com/smatsumae/devops-camp/internal/models"
)

type ListHandler struct {
	DB *sql.DB
}

func parseListID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func scanList(row interface {
	Scan(dest ...any) error
}) (models.List, error) {
	var l models.List
	err := row.Scan(&l.ID, &l.UserID, &l.Name, &l.CreatedAt, &l.UpdatedAt)
	return l, err
}

const listColumns = `id, user_id, name, created_at, updated_at`

// List handles GET /lists
func (h *ListHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	rows, err := h.DB.QueryContext(r.Context(),
		"SELECT "+listColumns+" FROM lists WHERE user_id = ? ORDER BY created_at ASC", userID)
	if err != nil {
		httpx.WriteInternalError(w, err)
		return
	}
	defer rows.Close()

	lists := []models.List{}
	for rows.Next() {
		l, err := scanList(rows)
		if err != nil {
			httpx.WriteInternalError(w, err)
			return
		}
		lists = append(lists, l)
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"data": lists})
}

type createListRequest struct {
	Name string `json:"name"`
}

// Create handles POST /lists
func (h *ListHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	var req createListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "request body is not valid JSON")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "name is required")
		return
	}

	res, err := h.DB.ExecContext(r.Context(),
		`INSERT INTO lists (user_id, name) VALUES (?, ?)`, userID, name)
	if err != nil {
		httpx.WriteInternalError(w, err)
		return
	}
	id, _ := res.LastInsertId()

	row := h.DB.QueryRowContext(r.Context(), "SELECT "+listColumns+" FROM lists WHERE id = ?", id)
	list, err := scanList(row)
	if err != nil {
		httpx.WriteInternalError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, list)
}

type updateListRequest struct {
	Name *string `json:"name"`
}

// Update handles PATCH /lists/{id}
func (h *ListHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	id, err := parseListID(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "id must be an integer")
		return
	}

	var owner int64
	err = h.DB.QueryRowContext(r.Context(), `SELECT user_id FROM lists WHERE id = ?`, id).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && owner != userID) {
		httpx.WriteError(w, http.StatusNotFound, "LIST_NOT_FOUND", "list not found")
		return
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		httpx.WriteInternalError(w, err)
		return
	}

	var req updateListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "request body is not valid JSON")
		return
	}

	var trimmedName string
	if req.Name != nil {
		trimmedName = strings.TrimSpace(*req.Name)
		if trimmedName == "" {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "name cannot be empty")
			return
		}
		if _, err := h.DB.ExecContext(r.Context(), `UPDATE lists SET name = ? WHERE id = ? AND user_id = ?`, trimmedName, id, userID); err != nil {
			httpx.WriteInternalError(w, err)
			return
		}
	}

	row := h.DB.QueryRowContext(r.Context(), "SELECT "+listColumns+" FROM lists WHERE id = ?", id)
	list, err := scanList(row)
	if err != nil {
		httpx.WriteInternalError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, list)
}

// Delete handles DELETE /lists/{id}
func (h *ListHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	id, err := parseListID(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "id must be an integer")
		return
	}

	res, err := h.DB.ExecContext(r.Context(), `DELETE FROM lists WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		httpx.WriteInternalError(w, err)
		return
	}
	affected, err := res.RowsAffected()
	if err != nil {
		httpx.WriteInternalError(w, err)
		return
	}
	if affected == 0 {
		httpx.WriteError(w, http.StatusNotFound, "LIST_NOT_FOUND", "list not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
