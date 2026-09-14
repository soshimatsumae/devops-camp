package models

import "time"

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

const (
	StatusTodo       = "todo"
	StatusInProgress = "in_progress"
	StatusDone       = "done"
)

type Task struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"-"`
	ParentID        *int64     `json:"parent_id"`
	Title           string     `json:"title"`
	Description     *string    `json:"description"`
	Status          string     `json:"status"`
	EstimatedWeight int        `json:"estimated_weight"`
	ActualWeight    *int       `json:"actual_weight"`
	DueDate         *string    `json:"due_date"`
	CompletedAt     *time.Time `json:"completed_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CalendarEntry struct {
	Date        string `json:"date"`
	TotalWeight int    `json:"total_weight"`
}
