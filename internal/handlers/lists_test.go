package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/smatsumae/devops-camp/internal/handlers"
)

func newListHandlerWithUser(t *testing.T) (*handlers.ListHandler, *handlers.TaskHandler, []byte, int64) {
	db := setupTestDB(t)
	secret := []byte("test-secret")

	res, err := db.Exec(`INSERT INTO users (email, password_hash, name) VALUES (?, ?, ?)`,
		"taro@example.com", "dummy-hash", "太郎")
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	userID, _ := res.LastInsertId()

	return &handlers.ListHandler{DB: db}, &handlers.TaskHandler{DB: db}, secret, userID
}

func TestListLifecycle_CreateUpdateDelete(t *testing.T) {
	lh, _, secret, userID := newListHandlerWithUser(t)

	createReq := newAuthedRequest(t, secret, userID, http.MethodPost, "/lists", []byte(`{"name":"プロジェクトA"}`))
	createRec := httptest.NewRecorder()
	lh.Create(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d, body=%s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	var created map[string]any
	decodeJSON(t, createRec, &created)
	listID := int64(created["id"].(float64))

	listReq := newAuthedRequest(t, secret, userID, http.MethodGet, "/lists", nil)
	listRec := httptest.NewRecorder()
	lh.List(listRec, listReq)
	var listed map[string]any
	decodeJSON(t, listRec, &listed)
	data, _ := listed["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 list, got %d", len(data))
	}

	renameReq := newAuthedRequest(t, secret, userID, http.MethodPatch, "/lists/"+strconv.FormatInt(listID, 10),
		[]byte(`{"name":"プロジェクトB"}`))
	renameReq.SetPathValue("id", strconv.FormatInt(listID, 10))
	renameRec := httptest.NewRecorder()
	lh.Update(renameRec, renameReq)
	if renameRec.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d, body=%s", renameRec.Code, http.StatusOK, renameRec.Body.String())
	}
	var renamed map[string]any
	decodeJSON(t, renameRec, &renamed)
	if renamed["name"] != "プロジェクトB" {
		t.Errorf("name = %v, want プロジェクトB", renamed["name"])
	}

	delReq := newAuthedRequest(t, secret, userID, http.MethodDelete, "/lists/"+strconv.FormatInt(listID, 10), nil)
	delReq.SetPathValue("id", strconv.FormatInt(listID, 10))
	delRec := httptest.NewRecorder()
	lh.Delete(delRec, delReq)
	if delRec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d, body=%s", delRec.Code, http.StatusNoContent, delRec.Body.String())
	}
}

func TestListDelete_UnsetsTaskListID(t *testing.T) {
	lh, th, secret, userID := newListHandlerWithUser(t)

	createListReq := newAuthedRequest(t, secret, userID, http.MethodPost, "/lists", []byte(`{"name":"プロジェクトA"}`))
	createListRec := httptest.NewRecorder()
	lh.Create(createListRec, createListReq)
	var list map[string]any
	decodeJSON(t, createListRec, &list)
	listID := int64(list["id"].(float64))

	createTaskReq := newAuthedRequest(t, secret, userID, http.MethodPost, "/tasks",
		[]byte(`{"title":"タスク","estimated_hours":1,"list_id":`+strconv.FormatInt(listID, 10)+`}`))
	createTaskRec := httptest.NewRecorder()
	th.Create(createTaskRec, createTaskReq)
	if createTaskRec.Code != http.StatusCreated {
		t.Fatalf("create task status = %d, want %d, body=%s", createTaskRec.Code, http.StatusCreated, createTaskRec.Body.String())
	}
	var task map[string]any
	decodeJSON(t, createTaskRec, &task)
	taskID := int64(task["id"].(float64))
	if task["list_id"].(float64) != float64(listID) {
		t.Fatalf("task list_id = %v, want %d", task["list_id"], listID)
	}

	delReq := newAuthedRequest(t, secret, userID, http.MethodDelete, "/lists/"+strconv.FormatInt(listID, 10), nil)
	delReq.SetPathValue("id", strconv.FormatInt(listID, 10))
	lh.Delete(httptest.NewRecorder(), delReq)

	getReq := newAuthedRequest(t, secret, userID, http.MethodGet, "/tasks/"+strconv.FormatInt(taskID, 10), nil)
	getReq.SetPathValue("id", strconv.FormatInt(taskID, 10))
	getRec := httptest.NewRecorder()
	th.Get(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get task status = %d, want %d, body=%s", getRec.Code, http.StatusOK, getRec.Body.String())
	}
	var refetched map[string]any
	decodeJSON(t, getRec, &refetched)
	if refetched["list_id"] != nil {
		t.Errorf("list_id = %v, want nil after list deletion", refetched["list_id"])
	}
}

func TestTaskCreate_ListIDNotAllowedForChildTask(t *testing.T) {
	lh, th, secret, userID := newListHandlerWithUser(t)

	createListReq := newAuthedRequest(t, secret, userID, http.MethodPost, "/lists", []byte(`{"name":"プロジェクトA"}`))
	createListRec := httptest.NewRecorder()
	lh.Create(createListRec, createListReq)
	var list map[string]any
	decodeJSON(t, createListRec, &list)
	listID := int64(list["id"].(float64))

	createParentReq := newAuthedRequest(t, secret, userID, http.MethodPost, "/tasks",
		[]byte(`{"title":"親タスク","estimated_hours":1}`))
	createParentRec := httptest.NewRecorder()
	th.Create(createParentRec, createParentReq)
	var parent map[string]any
	decodeJSON(t, createParentRec, &parent)
	parentID := int64(parent["id"].(float64))

	createChildReq := newAuthedRequest(t, secret, userID, http.MethodPost, "/tasks",
		[]byte(`{"title":"子タスク","estimated_hours":1,"parent_id":`+strconv.FormatInt(parentID, 10)+`,"list_id":`+strconv.FormatInt(listID, 10)+`}`))
	createChildRec := httptest.NewRecorder()
	th.Create(createChildRec, createChildReq)
	if createChildRec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d, body=%s", createChildRec.Code, http.StatusUnprocessableEntity, createChildRec.Body.String())
	}
	var got map[string]any
	decodeJSON(t, createChildRec, &got)
	if got["error"].(map[string]any)["code"] != "LIST_ID_NOT_ALLOWED_FOR_CHILD_TASK" {
		t.Errorf("error.code = %v, want LIST_ID_NOT_ALLOWED_FOR_CHILD_TASK", got["error"])
	}
}

func TestTaskCreate_InvalidListID(t *testing.T) {
	_, th, secret, userID := newListHandlerWithUser(t)

	req := newAuthedRequest(t, secret, userID, http.MethodPost, "/tasks",
		[]byte(`{"title":"タスク","estimated_hours":1,"list_id":9999}`))
	rec := httptest.NewRecorder()
	th.Create(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusUnprocessableEntity, rec.Body.String())
	}
}
