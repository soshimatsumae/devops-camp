package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/smatsumae/devops-camp/internal/handlers"
)

func newTaskHandlerWithUser(t *testing.T) (*handlers.TaskHandler, []byte, int64) {
	db := setupTestDB(t)
	secret := []byte("test-secret")

	res, err := db.Exec(`INSERT INTO users (email, password_hash, name) VALUES (?, ?, ?)`,
		"taro@example.com", "dummy-hash", "太郎")
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	userID, _ := res.LastInsertId()

	return &handlers.TaskHandler{DB: db}, secret, userID
}

func TestTaskCreate_RequiresEstimatedWeight(t *testing.T) {
	h, secret, userID := newTaskHandlerWithUser(t)

	req := newAuthedRequest(t, secret, userID, http.MethodPost, "/tasks", []byte(`{"title":"タスク"}`))
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusUnprocessableEntity, rec.Body.String())
	}
	var got map[string]any
	decodeJSON(t, rec, &got)
	if got["error"].(map[string]any)["code"] != "ESTIMATED_WEIGHT_REQUIRED" {
		t.Errorf("error.code = %v, want ESTIMATED_WEIGHT_REQUIRED", got["error"])
	}
}

func TestTaskCreate_InvalidParentID(t *testing.T) {
	h, secret, userID := newTaskHandlerWithUser(t)

	req := newAuthedRequest(t, secret, userID, http.MethodPost, "/tasks",
		[]byte(`{"title":"子タスク","estimated_weight":3,"parent_id":9999}`))
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusUnprocessableEntity, rec.Body.String())
	}
}

func TestTaskLifecycle_CreateGetUpdateDelete(t *testing.T) {
	h, secret, userID := newTaskHandlerWithUser(t)

	// Create a parent task.
	createReq := newAuthedRequest(t, secret, userID, http.MethodPost, "/tasks",
		[]byte(`{"title":"親タスク","estimated_weight":5}`))
	createRec := httptest.NewRecorder()
	h.Create(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d, body=%s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	var created map[string]any
	decodeJSON(t, createRec, &created)
	parentID := int64(created["id"].(float64))

	// Create a child task under it.
	childReq := newAuthedRequest(t, secret, userID, http.MethodPost, "/tasks",
		[]byte(`{"title":"子タスク","estimated_weight":2,"parent_id":`+strconv.FormatInt(parentID, 10)+`}`))
	childRec := httptest.NewRecorder()
	h.Create(childRec, childReq)
	if childRec.Code != http.StatusCreated {
		t.Fatalf("child create status = %d, want %d, body=%s", childRec.Code, http.StatusCreated, childRec.Body.String())
	}

	// GET the parent and confirm the child shows up.
	getReq := newAuthedRequest(t, secret, userID, http.MethodGet, "/tasks/"+strconv.FormatInt(parentID, 10), nil)
	getReq.SetPathValue("id", strconv.FormatInt(parentID, 10))
	getRec := httptest.NewRecorder()
	h.Get(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d, body=%s", getRec.Code, http.StatusOK, getRec.Body.String())
	}
	var fetched map[string]any
	decodeJSON(t, getRec, &fetched)
	children, _ := fetched["children"].([]any)
	if len(children) != 1 {
		t.Fatalf("expected 1 child task, got %d", len(children))
	}

	// Completing the task without actual_weight should fail.
	badUpdateReq := newAuthedRequest(t, secret, userID, http.MethodPatch, "/tasks/"+strconv.FormatInt(parentID, 10),
		[]byte(`{"status":"done"}`))
	badUpdateReq.SetPathValue("id", strconv.FormatInt(parentID, 10))
	badUpdateRec := httptest.NewRecorder()
	h.Update(badUpdateRec, badUpdateReq)
	if badUpdateRec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("update status = %d, want %d, body=%s", badUpdateRec.Code, http.StatusUnprocessableEntity, badUpdateRec.Body.String())
	}

	// Completing with actual_weight should succeed and persist.
	updateReq := newAuthedRequest(t, secret, userID, http.MethodPatch, "/tasks/"+strconv.FormatInt(parentID, 10),
		[]byte(`{"status":"done","actual_weight":8}`))
	updateReq.SetPathValue("id", strconv.FormatInt(parentID, 10))
	updateRec := httptest.NewRecorder()
	h.Update(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d, body=%s", updateRec.Code, http.StatusOK, updateRec.Body.String())
	}
	var updated map[string]any
	decodeJSON(t, updateRec, &updated)
	if updated["status"] != "done" {
		t.Errorf("status = %v, want done", updated["status"])
	}
	if updated["completed_at"] == nil {
		t.Error("completed_at should be set once a task is marked done")
	}

	// Re-fetch to confirm persistence, independent of the update response.
	reGetReq := newAuthedRequest(t, secret, userID, http.MethodGet, "/tasks/"+strconv.FormatInt(parentID, 10), nil)
	reGetReq.SetPathValue("id", strconv.FormatInt(parentID, 10))
	reGetRec := httptest.NewRecorder()
	h.Get(reGetRec, reGetReq)
	var reFetched map[string]any
	decodeJSON(t, reGetRec, &reFetched)
	if reFetched["actual_weight"].(float64) != 8 {
		t.Errorf("persisted actual_weight = %v, want 8", reFetched["actual_weight"])
	}

	// Delete and confirm it is gone.
	delReq := newAuthedRequest(t, secret, userID, http.MethodDelete, "/tasks/"+strconv.FormatInt(parentID, 10), nil)
	delReq.SetPathValue("id", strconv.FormatInt(parentID, 10))
	delRec := httptest.NewRecorder()
	h.Delete(delRec, delReq)
	if delRec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d, body=%s", delRec.Code, http.StatusNoContent, delRec.Body.String())
	}

	afterDeleteReq := newAuthedRequest(t, secret, userID, http.MethodGet, "/tasks/"+strconv.FormatInt(parentID, 10), nil)
	afterDeleteReq.SetPathValue("id", strconv.FormatInt(parentID, 10))
	afterDeleteRec := httptest.NewRecorder()
	h.Get(afterDeleteRec, afterDeleteReq)
	if afterDeleteRec.Code != http.StatusNotFound {
		t.Fatalf("get-after-delete status = %d, want %d", afterDeleteRec.Code, http.StatusNotFound)
	}
}

func TestTaskGet_OtherUsersTaskNotFound(t *testing.T) {
	h, secret, userID := newTaskHandlerWithUser(t)

	res, err := h.DB.Exec(`INSERT INTO users (email, password_hash, name) VALUES (?, ?, ?)`,
		"jiro@example.com", "dummy-hash", "次郎")
	if err != nil {
		t.Fatalf("failed to seed second user: %v", err)
	}
	otherUserID, _ := res.LastInsertId()

	createReq := newAuthedRequest(t, secret, userID, http.MethodPost, "/tasks",
		[]byte(`{"title":"太郎のタスク","estimated_weight":3}`))
	createRec := httptest.NewRecorder()
	h.Create(createRec, createReq)
	var created map[string]any
	decodeJSON(t, createRec, &created)
	taskID := int64(created["id"].(float64))

	getReq := newAuthedRequest(t, secret, otherUserID, http.MethodGet, "/tasks/"+strconv.FormatInt(taskID, 10), nil)
	getReq.SetPathValue("id", strconv.FormatInt(taskID, 10))
	getRec := httptest.NewRecorder()
	h.Get(getRec, getReq)

	if getRec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body=%s", getRec.Code, http.StatusNotFound, getRec.Body.String())
	}
}

func TestTaskCalendar_AggregatesByCompletionDate(t *testing.T) {
	h, secret, userID := newTaskHandlerWithUser(t)

	for _, weight := range []string{"3", "5"} {
		createReq := newAuthedRequest(t, secret, userID, http.MethodPost, "/tasks",
			[]byte(`{"title":"タスク","estimated_weight":`+weight+`}`))
		createRec := httptest.NewRecorder()
		h.Create(createRec, createReq)
		var created map[string]any
		decodeJSON(t, createRec, &created)
		taskID := int64(created["id"].(float64))

		updateReq := newAuthedRequest(t, secret, userID, http.MethodPatch, "/tasks/"+strconv.FormatInt(taskID, 10),
			[]byte(`{"status":"done","actual_weight":`+weight+`}`))
		updateReq.SetPathValue("id", strconv.FormatInt(taskID, 10))
		h.Update(httptest.NewRecorder(), updateReq)
	}

	calReq := newAuthedRequest(t, secret, userID, http.MethodGet, "/tasks/calendar", nil)
	calRec := httptest.NewRecorder()
	h.Calendar(calRec, calReq)

	if calRec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", calRec.Code, http.StatusOK, calRec.Body.String())
	}
	var got map[string]any
	decodeJSON(t, calRec, &got)
	data := got["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 aggregated day, got %d (%v)", len(data), data)
	}
	entry := data[0].(map[string]any)
	if entry["total_weight"].(float64) != 8 {
		t.Errorf("total_weight = %v, want 8", entry["total_weight"])
	}
}
