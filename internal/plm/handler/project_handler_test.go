package handler

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/bitfantasy/nimo/internal/plm/entity"
	"github.com/bitfantasy/nimo/internal/plm/repository"
	"github.com/bitfantasy/nimo/internal/plm/service"
	"github.com/bitfantasy/nimo/internal/plm/testutil"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func setupProjectTest(t *testing.T) (*testutil.TestEnv, *ProjectHandler) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	router := testutil.SetupRouter()

	// Need Redis for some services
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "127.0.0.1"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisHost + ":" + redisPort,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("Skipping project tests: Redis unavailable: %v", err)
	}
	t.Cleanup(func() { rdb.Close() })

	// Create repositories and service
	repos := repository.NewRepositories(db)
	projectSvc := service.NewProjectService(
		repos.Project,
		repos.Task,
		repos.Product,
		nil, // feishuSvc
		repos.TaskForm,
	)
	handler := NewProjectHandler(projectSvc)

	api := testutil.AuthGroup(router, "/api/v1")
	api.GET("/projects", handler.ListProjects)
	api.GET("/projects/:id", handler.GetProject)
	api.POST("/projects", handler.CreateProject)
	api.PUT("/projects/:id", handler.UpdateProject)
	api.DELETE("/projects/:id", handler.DeleteProject)
	api.PUT("/projects/:id/status", handler.UpdateProjectStatus)

	return &testutil.TestEnv{DB: db, Router: router, T: t}, handler
}

func seedTestProject(t *testing.T, db *gorm.DB, id, name, managerID string) *entity.Project {
	t.Helper()
	project := &entity.Project{
		ID:        id,
		Code:      "PRJ-" + id,
		Name:      name,
		Status:    "planning",
		Phase:     "concept",
		ManagerID: managerID,
		CreatedBy: managerID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("Failed to seed test project: %v", err)
	}
	return project
}

func TestProjectList(t *testing.T) {
	env, _ := setupProjectTest(t)
	token := testutil.DefaultTestToken()

	// Seed user and projects
	testutil.SeedTestUser(t, env.DB, "test-user-001", "Test Admin", "admin@test.com")
	seedTestProject(t, env.DB, "proj-001", "测试项目1", "test-user-001")
	seedTestProject(t, env.DB, "proj-002", "测试项目2", "test-user-001")

	w := testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/projects", nil, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := testutil.ParseResponse(w)
	if resp["code"].(float64) != 0 {
		t.Fatalf("expected code 0, got %v", resp["code"])
	}
}

func TestProjectCreate(t *testing.T) {
	env, _ := setupProjectTest(t)
	token := testutil.DefaultTestToken()

	// Seed the manager user
	testutil.SeedTestUser(t, env.DB, "test-user-001", "Test Admin", "admin@test.com")

	// Create a sequence for project code generation
	env.DB.Exec("CREATE SEQUENCE IF NOT EXISTS project_code_seq")

	body := map[string]interface{}{
		"name":        "新建测试项目",
		"description": "这是一个测试项目",
	}

	w := testutil.DoRequest(env.Router, http.MethodPost, "/api/v1/projects", body, token)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	resp := testutil.ParseResponse(w)
	if resp["code"].(float64) != 0 {
		t.Fatalf("expected code 0, got %v: %s", resp["code"], w.Body.String())
	}
}

func TestProjectGet(t *testing.T) {
	env, _ := setupProjectTest(t)
	token := testutil.DefaultTestToken()

	testutil.SeedTestUser(t, env.DB, "test-user-001", "Test Admin", "admin@test.com")
	seedTestProject(t, env.DB, "proj-get-001", "获取项目", "test-user-001")

	w := testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/projects/proj-get-001", nil, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestProjectGetNotFound(t *testing.T) {
	env, _ := setupProjectTest(t)
	token := testutil.DefaultTestToken()

	w := testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/projects/nonexistent", nil, token)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestProjectUpdate(t *testing.T) {
	env, _ := setupProjectTest(t)
	token := testutil.DefaultTestToken()

	testutil.SeedTestUser(t, env.DB, "test-user-001", "Test Admin", "admin@test.com")
	seedTestProject(t, env.DB, "proj-upd-001", "更新前", "test-user-001")

	body := map[string]interface{}{
		"name": "更新后的项目",
	}

	w := testutil.DoRequest(env.Router, http.MethodPut, "/api/v1/projects/proj-upd-001", body, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestProjectDelete(t *testing.T) {
	env, _ := setupProjectTest(t)
	token := testutil.DefaultTestToken()

	testutil.SeedTestUser(t, env.DB, "test-user-001", "Test Admin", "admin@test.com")
	seedTestProject(t, env.DB, "proj-del-001", "删除项目", "test-user-001")

	w := testutil.DoRequest(env.Router, http.MethodDelete, "/api/v1/projects/proj-del-001", nil, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestProjectUnauthorized(t *testing.T) {
	env, _ := setupProjectTest(t)

	w := testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/projects", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestProjectCreateBadRequest(t *testing.T) {
	env, _ := setupProjectTest(t)
	token := testutil.DefaultTestToken()

	// Missing required "name" field
	body := map[string]interface{}{
		"description": "no name",
	}

	w := testutil.DoRequest(env.Router, http.MethodPost, "/api/v1/projects", body, token)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
