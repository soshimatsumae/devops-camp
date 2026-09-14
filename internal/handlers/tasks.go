package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/smatsumae/devops-camp/internal/auth"
	"github.com/smatsumae/devops-camp/internal/httpx"
	"github.com/smatsumae/devops-camp/internal/models"
)

type TaskHandler struct {
	DB *sql.DB
}

var (
	errInvalidDateFormat = errors.New("invalid date format")
	errInvalidDateRange  = errors.New("invalid date range")
)

var validStatuses = map[string]bool{
	models.StatusTodo:       true,
	models.StatusInProgress: true,
	models.StatusDone:       true,
}

func currentUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return 0, false
	}
	return userID, true
}

func scanTask(row interface {
	Scan(dest ...any) error
}) (models.Task, error) {
	var t models.Task
	var (
		parentID     sql.NullInt64
		description  sql.NullString
		actualWeight sql.NullInt64
		dueDate      sql.NullString
		completedAt  sql.NullTime
	)
	err := row.Scan(
		&t.ID, &t.UserID, &parentID, &t.Title, &description, &t.Status,
		&t.EstimatedWeight, &actualWeight, &dueDate, &completedAt,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return models.Task{}, err
	}
	if parentID.Valid {
		t.ParentID = &parentID.Int64
	}
	if description.Valid {
		t.Description = &description.String
	}
	if actualWeight.Valid {
		v := int(actualWeight.Int64)
		t.ActualWeight = &v
	}
	if dueDate.Valid {
		d := dueDate.String[:10]
		t.DueDate = &d
	}
	if completedAt.Valid {
		t.CompletedAt = &completedAt.Time
	}
	return t, nil
}

const taskColumns = `id, user_id, parent_id, title, description, status,
	estimated_weight, actual_weight, due_date, completed_at, created_at, updated_at`

// List handles GET /tasks
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	q := r.URL.Query()

	page := 1
	if v, err := strconv.Atoi(q.Get("page")); err == nil && v > 0 {
		page = v
	}
	perPage := 20
	if v, err := strconv.Atoi(q.Get("per_page")); err == nil && v > 0 && v <= 100 {
		perPage = v
	}

	where := []string{"user_id = ?"}
	args := []any{userID}

	if status := q.Get("status"); status != "" {
		if !validStatuses[status] {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_STATUS", "status must be one of todo, in_progress, done")
			return
		}
		where = append(where, "status = ?")
		args = append(args, status)
	}

	if q.Has("parent_id") {
		raw := q.Get("parent_id")
		if raw == "null" {
			where = append(where, "parent_id IS NULL")
		} else {
			parentID, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "INVALID_PARENT_ID", "parent_id must be an integer or \"null\"")
				return
			}
			where = append(where, "parent_id = ?")
			args = append(args, parentID)
		}
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := "SELECT COUNT(*) FROM tasks WHERE " + whereClause
	if err := h.DB.QueryRowContext(r.Context(), countQuery, args...).Scan(&total); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to count tasks")
		return
	}

	listArgs := append(append([]any{}, args...), perPage, (page-1)*perPage)
	listQuery := "SELECT " + taskColumns + " FROM tasks WHERE " + whereClause +
		" ORDER BY created_at DESC LIMIT ? OFFSET ?"

	rows, err := h.DB.QueryContext(r.Context(), listQuery, listArgs...)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list tasks")
		return
	}
	defer rows.Close()

	tasks := []models.Task{}
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to read task row")
			return
		}
		tasks = append(tasks, t)
	}

	totalPages := (total + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"data": tasks,
		"meta": map[string]any{
			"total_count": total,
			"page":        page,
			"per_page":    perPage,
			"total_pages": totalPages,
		},
	})
}

type createTaskRequest struct {
	Title           string  `json:"title"`
	Description     *string `json:"description"`
	ParentID        *int64  `json:"parent_id"`
	EstimatedWeight *int    `json:"estimated_weight"`
	DueDate         *string `json:"due_date"`
}

// Create handles POST /tasks
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "request body is not valid JSON")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "title is required")
		return
	}
	if req.EstimatedWeight == nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "ESTIMATED_WEIGHT_REQUIRED", "estimated_weight is required")
		return
	}

	if req.ParentID != nil {
		var owner int64
		err := h.DB.QueryRowContext(r.Context(), `SELECT user_id FROM tasks WHERE id = ?`, *req.ParentID).Scan(&owner)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && owner != userID) {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "INVALID_PARENT_ID", "invalid parent_id")
			return
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to validate parent_id")
			return
		}
	}

	res, err := h.DB.ExecContext(r.Context(),
		`INSERT INTO tasks (user_id, parent_id, title, description, status, estimated_weight, due_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, req.ParentID, req.Title, req.Description, models.StatusTodo, *req.EstimatedWeight, req.DueDate,
	)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create task")
		return
	}
	id, _ := res.LastInsertId()

	task, err := h.loadTask(r, id, userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load created task")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) loadTask(r *http.Request, id, userID int64) (models.Task, error) {
	row := h.DB.QueryRowContext(r.Context(),
		"SELECT "+taskColumns+" FROM tasks WHERE id = ? AND user_id = ?", id, userID)
	return scanTask(row)
}

func parseTaskID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

// Get handles GET /tasks/{id}
func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	id, err := parseTaskID(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "id must be an integer")
		return
	}

	task, err := h.loadTask(r, id, userID)
	if errors.Is(err, sql.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "TASK_NOT_FOUND", "task not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load task")
		return
	}

	rows, err := h.DB.QueryContext(r.Context(),
		"SELECT "+taskColumns+" FROM tasks WHERE parent_id = ? AND user_id = ? ORDER BY created_at ASC", id, userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load child tasks")
		return
	}
	defer rows.Close()

	children := []models.Task{}
	for rows.Next() {
		child, err := scanTask(rows)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to read child task")
			return
		}
		children = append(children, child)
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"id":               task.ID,
		"parent_id":        task.ParentID,
		"title":            task.Title,
		"description":      task.Description,
		"status":           task.Status,
		"estimated_weight": task.EstimatedWeight,
		"actual_weight":    task.ActualWeight,
		"due_date":         task.DueDate,
		"completed_at":     task.CompletedAt,
		"created_at":       task.CreatedAt,
		"updated_at":       task.UpdatedAt,
		"children":         children,
	})
}

type updateTaskRequest struct {
	Title        *string `json:"title"`
	Description  *string `json:"description"`
	Status       *string `json:"status"`
	ActualWeight *int    `json:"actual_weight"`
	DueDate      *string `json:"due_date"`
}

// Update handles PATCH /tasks/{id}
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	id, err := parseTaskID(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "id must be an integer")
		return
	}

	existing, err := h.loadTask(r, id, userID)
	if errors.Is(err, sql.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "TASK_NOT_FOUND", "task not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load task")
		return
	}

	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "request body is not valid JSON")
		return
	}

	if req.Status != nil && !validStatuses[*req.Status] {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "status must be one of todo, in_progress, done")
		return
	}

	completingNow := req.Status != nil && *req.Status == models.StatusDone && existing.Status != models.StatusDone
	if completingNow && req.ActualWeight == nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "ACTUAL_WEIGHT_REQUIRED", "actual_weight is required when completing a task")
		return
	}

	sets := []string{}
	args := []any{}

	if req.Title != nil {
		sets = append(sets, "title = ?")
		args = append(args, strings.TrimSpace(*req.Title))
	}
	if req.Description != nil {
		sets = append(sets, "description = ?")
		args = append(args, *req.Description)
	}
	if req.Status != nil {
		sets = append(sets, "status = ?")
		args = append(args, *req.Status)
	}
	if req.ActualWeight != nil {
		sets = append(sets, "actual_weight = ?")
		args = append(args, *req.ActualWeight)
	}
	if req.DueDate != nil {
		sets = append(sets, "due_date = ?")
		args = append(args, *req.DueDate)
	}
	if completingNow {
		sets = append(sets, "completed_at = ?")
		args = append(args, time.Now().UTC())
	}

	if len(sets) > 0 {
		args = append(args, id, userID)
		query := "UPDATE tasks SET " + strings.Join(sets, ", ") + " WHERE id = ? AND user_id = ?"
		if _, err := h.DB.ExecContext(r.Context(), query, args...); err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update task")
			return
		}
	}

	updated, err := h.loadTask(r, id, userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load updated task")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, updated)
}

// Delete handles DELETE /tasks/{id}
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	id, err := parseTaskID(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "id must be an integer")
		return
	}

	res, err := h.DB.ExecContext(r.Context(), `DELETE FROM tasks WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete task")
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		httpx.WriteError(w, http.StatusNotFound, "TASK_NOT_FOUND", "task not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Calendar handles GET /tasks/calendar
const calendarDateLayout = "2006-01-02"

// parseCalendarRange resolves the from/to query params into a concrete date
// range. Missing values default to [today-365d, today] so that the response
// size stays bounded even for long-lived accounts.
func parseCalendarRange(q url.Values) (from, to time.Time, err error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	from = today.AddDate(0, 0, -365)
	to = today

	if raw := q.Get("from"); raw != "" {
		from, err = time.Parse(calendarDateLayout, raw)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("from: %w", errInvalidDateFormat)
		}
	}
	if raw := q.Get("to"); raw != "" {
		to, err = time.Parse(calendarDateLayout, raw)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("to: %w", errInvalidDateFormat)
		}
	}
	if from.After(to) {
		return time.Time{}, time.Time{}, errInvalidDateRange
	}
	return from, to, nil
}

func (h *TaskHandler) Calendar(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	from, to, err := parseCalendarRange(r.URL.Query())
	if errors.Is(err, errInvalidDateFormat) {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_DATE_FORMAT", "from/to must be in YYYY-MM-DD format")
		return
	}
	if errors.Is(err, errInvalidDateRange) {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "INVALID_DATE_RANGE", "from must not be after to")
		return
	}

	rows, err := h.DB.QueryContext(r.Context(), `
		SELECT DATE(completed_at) AS d, SUM(actual_weight)
		FROM tasks
		WHERE user_id = ? AND status = ? AND completed_at IS NOT NULL
			AND DATE(completed_at) BETWEEN ? AND ?
		GROUP BY d
		ORDER BY d ASC`,
		userID, models.StatusDone, from.Format(calendarDateLayout), to.Format(calendarDateLayout),
	)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to aggregate calendar")
		return
	}
	defer rows.Close()

	entries := []models.CalendarEntry{}
	for rows.Next() {
		var date string
		var total int
		if err := rows.Scan(&date, &total); err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to read calendar row")
			return
		}
		entries = append(entries, models.CalendarEntry{Date: date, TotalWeight: total})
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"data": entries})
}
