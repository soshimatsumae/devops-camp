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

// validSorts maps the `sort` query param to a trusted ORDER BY clause. Using
// a whitelist (rather than building the clause from user input directly)
// avoids SQL injection via the sort column/direction.
var validSorts = map[string]string{
	"due_date_asc":         "due_date ASC",
	"due_date_desc":        "due_date DESC",
	"estimated_hours_asc":  "estimated_hours ASC",
	"estimated_hours_desc": "estimated_hours DESC",
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
		parentID    sql.NullInt64
		listID      sql.NullInt64
		description sql.NullString
		actualHours sql.NullFloat64
		dueDate     sql.NullString
		completedAt sql.NullTime
	)
	err := row.Scan(
		&t.ID, &t.UserID, &parentID, &listID, &t.Title, &description, &t.Status,
		&t.EstimatedHours, &actualHours, &dueDate, &completedAt,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return models.Task{}, err
	}
	if parentID.Valid {
		t.ParentID = &parentID.Int64
	}
	if listID.Valid {
		t.ListID = &listID.Int64
	}
	if description.Valid {
		t.Description = &description.String
	}
	if actualHours.Valid {
		v := actualHours.Float64
		t.ActualHours = &v
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

const taskColumns = `id, user_id, parent_id, list_id, title, description, status,
	estimated_hours, actual_hours, due_date, completed_at, created_at, updated_at`

const (
	minHours float64 = 0.5
	maxHours float64 = 9999
)

func validHours(h float64) bool {
	return h >= minHours && h <= maxHours
}

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

	if q.Has("list_id") {
		raw := q.Get("list_id")
		if raw == "null" {
			where = append(where, "list_id IS NULL")
		} else {
			listID, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "INVALID_LIST_ID", "list_id must be an integer or \"null\"")
				return
			}
			where = append(where, "list_id = ?")
			args = append(args, listID)
		}
	}

	if keyword := q.Get("q"); keyword != "" {
		where = append(where, "(title LIKE ? OR description LIKE ?)")
		likePattern := "%" + strings.NewReplacer("%", "\\%", "_", "\\_").Replace(keyword) + "%"
		args = append(args, likePattern, likePattern)
	}

	orderBy := "created_at DESC"
	if sort := q.Get("sort"); sort != "" {
		clause, ok := validSorts[sort]
		if !ok {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_SORT", "sort must be one of due_date_asc, due_date_desc, estimated_hours_asc, estimated_hours_desc")
			return
		}
		orderBy = clause
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := "SELECT COUNT(*) FROM tasks WHERE " + whereClause
	if err := h.DB.QueryRowContext(r.Context(), countQuery, args...).Scan(&total); err != nil {
		httpx.WriteInternalError(w, err)
		return
	}

	listArgs := append(append([]any{}, args...), perPage, (page-1)*perPage)
	listQuery := "SELECT " + taskColumns + " FROM tasks WHERE " + whereClause +
		" ORDER BY " + orderBy + " LIMIT ? OFFSET ?"

	rows, err := h.DB.QueryContext(r.Context(), listQuery, listArgs...)
	if err != nil {
		httpx.WriteInternalError(w, err)
		return
	}
	defer rows.Close()

	tasks := []models.Task{}
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			httpx.WriteInternalError(w, err)
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
	Title          string   `json:"title"`
	Description    *string  `json:"description"`
	ParentID       *int64   `json:"parent_id"`
	ListID         *int64   `json:"list_id"`
	EstimatedHours *float64 `json:"estimated_hours"`
	DueDate        *string  `json:"due_date"`
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
	if req.EstimatedHours == nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "ESTIMATED_HOURS_REQUIRED", "estimated_hours is required")
		return
	}
	if !validHours(*req.EstimatedHours) {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", fmt.Sprintf("estimated_hours must be between %g and %g", minHours, maxHours))
		return
	}
	if req.DueDate != nil {
		if _, err := time.Parse(calendarDateLayout, *req.DueDate); err != nil {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "due_date must be in YYYY-MM-DD format")
			return
		}
	}

	if req.ParentID != nil {
		var owner int64
		err := h.DB.QueryRowContext(r.Context(), `SELECT user_id FROM tasks WHERE id = ?`, *req.ParentID).Scan(&owner)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && owner != userID) {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "INVALID_PARENT_ID", "invalid parent_id")
			return
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			httpx.WriteInternalError(w, err)
			return
		}
	}

	if req.ListID != nil {
		if req.ParentID != nil {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "LIST_ID_NOT_ALLOWED_FOR_CHILD_TASK", "list_id can only be set on a parent task")
			return
		}
		var owner int64
		err := h.DB.QueryRowContext(r.Context(), `SELECT user_id FROM lists WHERE id = ?`, *req.ListID).Scan(&owner)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && owner != userID) {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "INVALID_LIST_ID", "invalid list_id")
			return
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			httpx.WriteInternalError(w, err)
			return
		}
	}

	res, err := h.DB.ExecContext(r.Context(),
		`INSERT INTO tasks (user_id, parent_id, list_id, title, description, status, estimated_hours, due_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, req.ParentID, req.ListID, req.Title, req.Description, models.StatusTodo, *req.EstimatedHours, req.DueDate,
	)
	if err != nil {
		httpx.WriteInternalError(w, err)
		return
	}
	id, _ := res.LastInsertId()

	task, err := h.loadTask(r, id, userID)
	if err != nil {
		httpx.WriteInternalError(w, err)
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
		httpx.WriteInternalError(w, err)
		return
	}

	rows, err := h.DB.QueryContext(r.Context(),
		"SELECT "+taskColumns+" FROM tasks WHERE parent_id = ? AND user_id = ? ORDER BY created_at ASC", id, userID)
	if err != nil {
		httpx.WriteInternalError(w, err)
		return
	}
	defer rows.Close()

	children := []models.Task{}
	for rows.Next() {
		child, err := scanTask(rows)
		if err != nil {
			httpx.WriteInternalError(w, err)
			return
		}
		children = append(children, child)
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"id":              task.ID,
		"parent_id":       task.ParentID,
		"list_id":         task.ListID,
		"title":           task.Title,
		"description":     task.Description,
		"status":          task.Status,
		"estimated_hours": task.EstimatedHours,
		"actual_hours":    task.ActualHours,
		"due_date":        task.DueDate,
		"completed_at":    task.CompletedAt,
		"created_at":      task.CreatedAt,
		"updated_at":      task.UpdatedAt,
		"children":        children,
	})
}

type updateTaskRequest struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Status      *string  `json:"status"`
	ListID      *int64   `json:"list_id"`
	ActualHours *float64 `json:"actual_hours"`
	DueDate     *string  `json:"due_date"`
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
		httpx.WriteInternalError(w, err)
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

	var trimmedTitle string
	if req.Title != nil {
		trimmedTitle = strings.TrimSpace(*req.Title)
		if trimmedTitle == "" {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "title cannot be empty")
			return
		}
	}
	if req.DueDate != nil {
		if _, err := time.Parse(calendarDateLayout, *req.DueDate); err != nil {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "due_date must be in YYYY-MM-DD format")
			return
		}
	}

	if req.ActualHours != nil && !validHours(*req.ActualHours) {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", fmt.Sprintf("actual_hours must be between %g and %g", minHours, maxHours))
		return
	}

	if req.ListID != nil {
		if existing.ParentID != nil {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "LIST_ID_NOT_ALLOWED_FOR_CHILD_TASK", "list_id can only be set on a parent task")
			return
		}
		var owner int64
		err := h.DB.QueryRowContext(r.Context(), `SELECT user_id FROM lists WHERE id = ?`, *req.ListID).Scan(&owner)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && owner != userID) {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "INVALID_LIST_ID", "invalid list_id")
			return
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			httpx.WriteInternalError(w, err)
			return
		}
	}

	completingNow := req.Status != nil && *req.Status == models.StatusDone && existing.Status != models.StatusDone
	if completingNow && req.ActualHours == nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "ACTUAL_HOURS_REQUIRED", "actual_hours is required when completing a task")
		return
	}
	uncompletingNow := req.Status != nil && *req.Status != models.StatusDone && existing.Status == models.StatusDone

	sets := []string{}
	args := []any{}

	if req.Title != nil {
		sets = append(sets, "title = ?")
		args = append(args, trimmedTitle)
	}
	if req.Description != nil {
		sets = append(sets, "description = ?")
		args = append(args, *req.Description)
	}
	if req.Status != nil {
		sets = append(sets, "status = ?")
		args = append(args, *req.Status)
	}
	if req.ActualHours != nil {
		sets = append(sets, "actual_hours = ?")
		args = append(args, *req.ActualHours)
	}
	if req.ListID != nil {
		sets = append(sets, "list_id = ?")
		args = append(args, *req.ListID)
	}
	if req.DueDate != nil {
		sets = append(sets, "due_date = ?")
		args = append(args, *req.DueDate)
	}
	if completingNow {
		sets = append(sets, "completed_at = ?")
		args = append(args, time.Now().UTC())
	}
	if uncompletingNow {
		sets = append(sets, "completed_at = NULL")
	}

	if len(sets) > 0 {
		args = append(args, id, userID)
		query := "UPDATE tasks SET " + strings.Join(sets, ", ") + " WHERE id = ? AND user_id = ?"
		if _, err := h.DB.ExecContext(r.Context(), query, args...); err != nil {
			httpx.WriteInternalError(w, err)
			return
		}
	}

	updated, err := h.loadTask(r, id, userID)
	if err != nil {
		httpx.WriteInternalError(w, err)
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
		httpx.WriteInternalError(w, err)
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
// range. Missing values default to a fixed 53-week (371 day) window starting
// on the first Sunday of the current year, so the heatmap grid's weekday
// columns stay aligned year over year regardless of what weekday Jan 1 falls
// on (see ADR-0010).
func parseCalendarRange(q url.Values) (from, to time.Time, err error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	jan1 := time.Date(today.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	daysUntilSunday := (7 - int(jan1.Weekday())) % 7
	from = jan1.AddDate(0, 0, daysUntilSunday)
	to = from.AddDate(0, 0, 370)

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
		SELECT DATE(completed_at) AS d, SUM(actual_hours)
		FROM tasks
		WHERE user_id = ? AND status = ? AND completed_at IS NOT NULL
			AND DATE(completed_at) BETWEEN ? AND ?
		GROUP BY d
		ORDER BY d ASC`,
		userID, models.StatusDone, from.Format(calendarDateLayout), to.Format(calendarDateLayout),
	)
	if err != nil {
		httpx.WriteInternalError(w, err)
		return
	}
	defer rows.Close()

	entries := []models.CalendarEntry{}
	for rows.Next() {
		var date string
		var total float64
		if err := rows.Scan(&date, &total); err != nil {
			httpx.WriteInternalError(w, err)
			return
		}
		entries = append(entries, models.CalendarEntry{Date: date, TotalHours: total})
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"data": entries})
}
