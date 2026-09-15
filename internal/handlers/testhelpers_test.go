package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/smatsumae/devops-camp/internal/auth"
)

// setupTestDB connects to a MySQL instance configured via TEST_MYSQL_DSN and
// resets the schema before each test. Tests are skipped when the env var is
// not set, since this sandbox cannot reach a local MySQL server -- run these
// against a real database from your own terminal:
//
//	TEST_MYSQL_DSN="root@tcp(127.0.0.1:3306)/devops_camp_test?parseTime=true&charset=utf8mb4" go test ./...
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN not set; skipping test that requires a live MySQL database")
	}

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	schema, err := os.ReadFile("../../schema.sql")
	if err != nil {
		t.Fatalf("failed to read schema.sql: %v", err)
	}
	for _, stmt := range strings.Split(string(schema), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := conn.Exec(stmt); err != nil {
			t.Fatalf("failed to apply schema: %v", err)
		}
	}

	for _, stmt := range []string{
		"SET FOREIGN_KEY_CHECKS=0",
		"TRUNCATE TABLE tasks",
		"TRUNCATE TABLE users",
		"SET FOREIGN_KEY_CHECKS=1",
	} {
		if _, err := conn.Exec(stmt); err != nil {
			t.Fatalf("failed to reset tables (%q): %v", stmt, err)
		}
	}

	return conn
}

func newAuthedRequest(t *testing.T, secret []byte, userID int64, method, target string, body []byte) *http.Request {
	t.Helper()

	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, target, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, target, nil)
	}

	token, err := auth.GenerateToken(secret, userID, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Tests call handlers directly rather than through the RequireAuth
	// middleware, so inject the authenticated user into the context the
	// same way the middleware would after validating the token above.
	req = req.WithContext(auth.ContextWithUserID(req.Context(), userID))
	return req
}

func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("failed to decode JSON response: %v (body=%s)", err, rec.Body.String())
	}
}
