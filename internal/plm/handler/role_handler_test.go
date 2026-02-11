package handler

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bitfantasy/nimo/internal/plm/entity"
	"github.com/bitfantasy/nimo/internal/plm/testutil"
)

func setupRoleTest(t *testing.T) (*testutil.TestEnv, *RoleHandler) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	router := testutil.SetupRouter()
	handler := NewRoleHandler(db, nil)

	api := testutil.AuthGroup(router, "/api/v1")
	api.GET("/roles", handler.List)
	api.GET("/roles/:id", handler.Get)
	api.POST("/roles", handler.Create)
	api.PUT("/roles/:id", handler.Update)
	api.DELETE("/roles/:id", handler.Delete)
	api.GET("/roles/:id/members", handler.ListMembers)
	api.POST("/roles/:id/members", handler.AddMembers)
	api.DELETE("/roles/:id/members", handler.RemoveMembers)
	api.GET("/departments", handler.ListDepartments)

	return &testutil.TestEnv{DB: db, Router: router, T: t}, handler
}

func TestRoleList(t *testing.T) {
	env, _ := setupRoleTest(t)
	token := testutil.DefaultTestToken()

	// Seed some roles
	testutil.SeedTestRole(t, env.DB, "role-001", "role_admin", "管理员", true)
	testutil.SeedTestRole(t, env.DB, "role-002", "role_pm", "项目经理", false)

	w := testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/roles", nil, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := testutil.ParseResponse(w)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object in response")
	}
	items, ok := data["items"].([]interface{})
	if !ok {
		t.Fatal("expected items array in data")
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(items))
	}
}

func TestRoleCreate(t *testing.T) {
	env, _ := setupRoleTest(t)
	token := testutil.DefaultTestToken()

	body := map[string]interface{}{
		"name":        "测试角色",
		"description": "这是一个测试角色",
	}

	w := testutil.DoRequest(env.Router, http.MethodPost, "/api/v1/roles", body, token)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	resp := testutil.ParseResponse(w)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object in response")
	}
	if data["name"] != "测试角色" {
		t.Fatalf("expected name '测试角色', got '%v'", data["name"])
	}
	if data["status"] != "active" {
		t.Fatalf("expected status 'active', got '%v'", data["status"])
	}
}

func TestRoleCreateBadRequest(t *testing.T) {
	env, _ := setupRoleTest(t)
	token := testutil.DefaultTestToken()

	// Missing required "name" field
	body := map[string]interface{}{
		"description": "no name",
	}

	w := testutil.DoRequest(env.Router, http.MethodPost, "/api/v1/roles", body, token)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRoleGet(t *testing.T) {
	env, _ := setupRoleTest(t)
	token := testutil.DefaultTestToken()

	testutil.SeedTestRole(t, env.DB, "role-get-001", "role_get_test", "获取测试", false)

	w := testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/roles/role-get-001", nil, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := testutil.ParseResponse(w)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object in response")
	}
	if data["name"] != "获取测试" {
		t.Fatalf("expected name '获取测试', got '%v'", data["name"])
	}
}

func TestRoleGetNotFound(t *testing.T) {
	env, _ := setupRoleTest(t)
	token := testutil.DefaultTestToken()

	w := testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/roles/nonexistent", nil, token)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRoleUpdate(t *testing.T) {
	env, _ := setupRoleTest(t)
	token := testutil.DefaultTestToken()

	testutil.SeedTestRole(t, env.DB, "role-upd-001", "role_upd_test", "更新前", false)

	body := map[string]interface{}{
		"name": "更新后",
	}

	w := testutil.DoRequest(env.Router, http.MethodPut, "/api/v1/roles/role-upd-001", body, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := testutil.ParseResponse(w)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object in response")
	}
	if data["name"] != "更新后" {
		t.Fatalf("expected name '更新后', got '%v'", data["name"])
	}
}

func TestRoleDelete(t *testing.T) {
	env, _ := setupRoleTest(t)
	token := testutil.DefaultTestToken()

	testutil.SeedTestRole(t, env.DB, "role-del-001", "role_del_test", "删除测试", false)

	w := testutil.DoRequest(env.Router, http.MethodDelete, "/api/v1/roles/role-del-001", nil, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify it's actually deleted
	var count int64
	env.DB.Model(&entity.Role{}).Where("id = ?", "role-del-001").Count(&count)
	if count != 0 {
		t.Fatal("expected role to be deleted from database")
	}
}

func TestRoleDeleteSystemRole(t *testing.T) {
	env, _ := setupRoleTest(t)
	token := testutil.DefaultTestToken()

	testutil.SeedTestRole(t, env.DB, "role-sys-001", "role_system", "系统角色", true)

	w := testutil.DoRequest(env.Router, http.MethodDelete, "/api/v1/roles/role-sys-001", nil, token)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for system role delete, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRoleMembers(t *testing.T) {
	env, _ := setupRoleTest(t)
	token := testutil.DefaultTestToken()

	// Setup: create role and user
	testutil.SeedTestRole(t, env.DB, "role-mem-001", "role_member_test", "成员测试", false)
	testutil.SeedTestUser(t, env.DB, "user-mem-001", "测试用户", "test@test.com")

	// Add member
	addBody := map[string]interface{}{
		"user_ids": []string{"user-mem-001"},
	}
	w := testutil.DoRequest(env.Router, http.MethodPost, "/api/v1/roles/role-mem-001/members", addBody, token)
	if w.Code != http.StatusOK {
		t.Fatalf("add member: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// List members
	w = testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/roles/role-mem-001/members", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("list members: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := testutil.ParseResponse(w)
	data := resp["data"].(map[string]interface{})
	items := data["items"].([]interface{})
	if len(items) != 1 {
		t.Fatalf("expected 1 member, got %d", len(items))
	}

	// Remove member
	removeBody := map[string]interface{}{
		"user_ids": []string{"user-mem-001"},
	}
	w = testutil.DoRequest(env.Router, http.MethodDelete, "/api/v1/roles/role-mem-001/members", removeBody, token)
	if w.Code != http.StatusOK {
		t.Fatalf("remove member: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify member removed
	w = testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/roles/role-mem-001/members", nil, token)
	resp = testutil.ParseResponse(w)
	data = resp["data"].(map[string]interface{})
	items = data["items"].([]interface{})
	if len(items) != 0 {
		t.Fatalf("expected 0 members after removal, got %d", len(items))
	}
}

func TestRoleCRUDFullFlow(t *testing.T) {
	env, _ := setupRoleTest(t)
	token := testutil.DefaultTestToken()

	// 1. Create
	createBody := map[string]interface{}{
		"name":        "全流程角色",
		"description": "测试完整流程",
	}
	w := testutil.DoRequest(env.Router, http.MethodPost, "/api/v1/roles", createBody, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	resp := testutil.ParseResponse(w)
	roleData := resp["data"].(map[string]interface{})
	roleID := roleData["id"].(string)

	// 2. Get
	w = testutil.DoRequest(env.Router, http.MethodGet, fmt.Sprintf("/api/v1/roles/%s", roleID), nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d", w.Code)
	}

	// 3. List (should have 1 role)
	w = testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/roles", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", w.Code)
	}

	// 4. Update
	updateBody := map[string]interface{}{
		"name": "更新后的角色",
	}
	w = testutil.DoRequest(env.Router, http.MethodPut, fmt.Sprintf("/api/v1/roles/%s", roleID), updateBody, token)
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d", w.Code)
	}

	// 5. Delete
	w = testutil.DoRequest(env.Router, http.MethodDelete, fmt.Sprintf("/api/v1/roles/%s", roleID), nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d", w.Code)
	}

	// 6. Verify deleted
	w = testutil.DoRequest(env.Router, http.MethodGet, fmt.Sprintf("/api/v1/roles/%s", roleID), nil, token)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected 404, got %d", w.Code)
	}
}

func TestRoleUnauthorized(t *testing.T) {
	env, _ := setupRoleTest(t)

	// No token
	w := testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/roles", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}

	// Invalid token
	w = testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/roles", nil, "invalid-token")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid token, got %d", w.Code)
	}
}

func TestDepartmentList(t *testing.T) {
	env, _ := setupRoleTest(t)
	token := testutil.DefaultTestToken()

	// Seed department and user
	dept := &entity.Department{
		ID:     "dept-001",
		Name:   "工程部",
		Status: "active",
	}
	env.DB.Create(dept)

	testutil.SeedTestUser(t, env.DB, "user-dept-001", "部门用户", "dept@test.com")
	env.DB.Model(&entity.User{}).Where("id = ?", "user-dept-001").Update("department_id", "dept-001")

	w := testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/departments", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
