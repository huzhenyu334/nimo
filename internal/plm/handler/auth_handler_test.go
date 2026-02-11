package handler

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bitfantasy/nimo/internal/config"
	"github.com/bitfantasy/nimo/internal/plm/repository"
	"github.com/bitfantasy/nimo/internal/plm/service"
	"github.com/bitfantasy/nimo/internal/plm/testutil"
	"github.com/redis/go-redis/v9"
)

func getRedisAddr() string {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	return host + ":" + port
}

func setupAuthTest(t *testing.T) (*testutil.TestEnv, *AuthHandler) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	router := testutil.SetupRouter()

	// Setup Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: getRedisAddr(),
	})

	// Verify Redis connection
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("Skipping auth tests: Redis unavailable: %v", err)
	}
	t.Cleanup(func() { rdb.Close() })

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:             testutil.JWTSecret,
			AccessTokenExpire:  24 * time.Hour,
			RefreshTokenExpire: 7 * 24 * time.Hour,
			Issuer:             "nimo-plm-test",
		},
		Feishu: config.FeishuConfig{
			AppID:       "test-app-id",
			AppSecret:   "test-app-secret",
			RedirectURI: "http://localhost:8080/api/v1/auth/feishu/callback",
		},
	}

	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo, rdb, cfg)
	handler := NewAuthHandler(authSvc, cfg)

	// Public routes (no auth)
	router.GET("/api/v1/auth/feishu/login", handler.FeishuLogin)
	router.GET("/api/v1/auth/feishu/callback", handler.FeishuCallback)
	router.POST("/api/v1/auth/refresh", handler.RefreshToken)

	// Protected routes
	api := testutil.AuthGroup(router, "/api/v1")
	api.GET("/auth/me", handler.GetCurrentUser)
	api.POST("/auth/logout", handler.Logout)

	return &testutil.TestEnv{DB: db, Router: router, T: t}, handler
}

func TestFeishuLoginRedirect(t *testing.T) {
	env, _ := setupAuthTest(t)

	w := testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/auth/feishu/login", nil, "")

	// Should redirect to Feishu OAuth URL
	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect 302, got %d: %s", w.Code, w.Body.String())
	}

	location := w.Header().Get("Location")
	if location == "" {
		t.Fatal("expected Location header in redirect")
	}

	if !strings.Contains(location, "open.feishu.cn") {
		t.Fatalf("expected redirect to feishu, got: %s", location)
	}

	if !strings.Contains(location, "test-app-id") {
		t.Fatalf("expected app_id in redirect URL, got: %s", location)
	}
}

func TestFeishuCallbackMissingCode(t *testing.T) {
	env, _ := setupAuthTest(t)

	w := testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/auth/feishu/callback", nil, "")

	resp := testutil.ParseResponse(w)
	code, _ := resp["code"].(float64)
	if code != 40001 {
		t.Fatalf("expected error code 40001, got %v (status %d)", code, w.Code)
	}
}

func TestGetCurrentUser(t *testing.T) {
	env, _ := setupAuthTest(t)

	// Seed a test user matching the default token user ID
	testutil.SeedTestUser(t, env.DB, "test-user-001", "Test Admin", "admin@test.com")

	token := testutil.DefaultTestToken()
	w := testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/auth/me", nil, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := testutil.ParseResponse(w)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object in response")
	}
	if data["name"] != "Test Admin" {
		t.Fatalf("expected name 'Test Admin', got '%v'", data["name"])
	}
	if data["email"] != "admin@test.com" {
		t.Fatalf("expected email 'admin@test.com', got '%v'", data["email"])
	}
}

func TestGetCurrentUserUnauthorized(t *testing.T) {
	env, _ := setupAuthTest(t)

	w := testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/auth/me", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestGetCurrentUserNotFound(t *testing.T) {
	env, _ := setupAuthTest(t)

	// Token with user ID that doesn't exist in DB
	token := testutil.GenerateTestToken("nonexistent-user", "Ghost", "ghost@test.com", nil, nil)
	w := testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/auth/me", nil, token)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLogout(t *testing.T) {
	env, _ := setupAuthTest(t)

	testutil.SeedTestUser(t, env.DB, "test-user-001", "Test Admin", "admin@test.com")

	token := testutil.DefaultTestToken()
	w := testutil.DoRequest(env.Router, http.MethodPost, "/api/v1/auth/logout", nil, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRefreshTokenInvalid(t *testing.T) {
	env, _ := setupAuthTest(t)

	body := map[string]interface{}{
		"refresh_token": "invalid-token-string",
	}
	w := testutil.DoRequest(env.Router, http.MethodPost, "/api/v1/auth/refresh", body, "")

	resp := testutil.ParseResponse(w)
	code, _ := resp["code"].(float64)
	// Should be unauthorized (40100)
	if code < 40100 {
		t.Fatalf("expected unauthorized error code, got %v (status %d)", code, w.Code)
	}
}

func TestRefreshTokenMissingBody(t *testing.T) {
	env, _ := setupAuthTest(t)

	w := testutil.DoRequest(env.Router, http.MethodPost, "/api/v1/auth/refresh", nil, "")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
