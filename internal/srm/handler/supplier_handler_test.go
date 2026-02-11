package handler

import (
	"net/http"
	"testing"
	"time"

	"github.com/bitfantasy/nimo/internal/plm/testutil"
	"github.com/bitfantasy/nimo/internal/srm/entity"
	"github.com/bitfantasy/nimo/internal/srm/repository"
	"github.com/bitfantasy/nimo/internal/srm/service"
)

func setupSupplierTest(t *testing.T) (*testutil.TestEnv, *SupplierHandler) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	// Migrate SRM supplier table in test schema
	if err := db.AutoMigrate(&entity.Supplier{}, &entity.SupplierContact{}); err != nil {
		t.Fatalf("Failed to migrate SRM tables: %v", err)
	}

	repo := repository.NewSupplierRepository(db)
	svc := service.NewSupplierService(repo)
	handler := NewSupplierHandler(svc)

	router := testutil.SetupRouter()
	api := testutil.AuthGroup(router, "/api/v1/srm")
	api.GET("/suppliers", handler.ListSuppliers)
	api.GET("/suppliers/:id", handler.GetSupplier)

	return &testutil.TestEnv{DB: db, Router: router, T: t}, handler
}

func seedTestSupplier(t *testing.T, env *testutil.TestEnv, id, code, name, category, status string) {
	t.Helper()
	supplier := &entity.Supplier{
		ID:        id,
		Code:      code,
		Name:      name,
		Category:  category,
		Status:    status,
		CreatedBy: "test-user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := env.DB.Create(supplier).Error; err != nil {
		t.Fatalf("Failed to seed test supplier: %v", err)
	}
}

// TestSupplierListNoFilter verifies that listing suppliers without status filter
// returns ALL suppliers regardless of status. This is the scenario the kanban board uses.
func TestSupplierListNoFilter(t *testing.T) {
	env, _ := setupSupplierTest(t)
	token := testutil.DefaultTestToken()

	// Seed suppliers with different statuses
	seedTestSupplier(t, env, "sup-001", "SUP-0001", "供应商A", "electronic", "pending")
	seedTestSupplier(t, env, "sup-002", "SUP-0002", "供应商B", "structural", "active")
	seedTestSupplier(t, env, "sup-003", "SUP-0003", "供应商C", "optical", "suspended")

	// Request without status filter (same as kanban does after fix)
	w := testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/srm/suppliers?page_size=200", nil, token)

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
	if len(items) != 3 {
		t.Fatalf("expected 3 suppliers (all statuses), got %d", len(items))
	}
}

// TestSupplierListWithStatusFilter verifies that filtering by status works correctly.
func TestSupplierListWithStatusFilter(t *testing.T) {
	env, _ := setupSupplierTest(t)
	token := testutil.DefaultTestToken()

	seedTestSupplier(t, env, "sup-001", "SUP-0001", "供应商A", "electronic", "pending")
	seedTestSupplier(t, env, "sup-002", "SUP-0002", "供应商B", "structural", "active")
	seedTestSupplier(t, env, "sup-003", "SUP-0003", "供应商C", "optical", "suspended")

	// Filter by active status
	w := testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/srm/suppliers?status=active&page_size=200", nil, token)

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
	if len(items) != 1 {
		t.Fatalf("expected 1 active supplier, got %d", len(items))
	}

	// The one result should be the active supplier
	item := items[0].(map[string]interface{})
	if item["name"] != "供应商B" {
		t.Fatalf("expected active supplier '供应商B', got '%v'", item["name"])
	}
}

// TestSupplierListPendingOnly verifies that when all suppliers are pending
// (common initial state), the no-filter query still returns them.
// This was the original bug: kanban filtered by active, but all suppliers were pending.
func TestSupplierListPendingOnly(t *testing.T) {
	env, _ := setupSupplierTest(t)
	token := testutil.DefaultTestToken()

	// Seed all suppliers as pending (common initial state)
	seedTestSupplier(t, env, "sup-001", "SUP-0001", "待审核供应商A", "electronic", "pending")
	seedTestSupplier(t, env, "sup-002", "SUP-0002", "待审核供应商B", "structural", "pending")

	// Without status filter, should return all pending suppliers
	w := testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/srm/suppliers?page_size=200", nil, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := testutil.ParseResponse(w)
	data := resp["data"].(map[string]interface{})
	items := data["items"].([]interface{})
	if len(items) != 2 {
		t.Fatalf("expected 2 pending suppliers, got %d", len(items))
	}

	// With active filter, should return 0 (this was the old buggy behavior)
	w2 := testutil.DoRequest(env.Router, http.MethodGet, "/api/v1/srm/suppliers?status=active&page_size=200", nil, token)
	resp2 := testutil.ParseResponse(w2)
	data2 := resp2["data"].(map[string]interface{})
	items2 := data2["items"].([]interface{})
	if len(items2) != 0 {
		t.Fatalf("expected 0 active suppliers when all are pending, got %d", len(items2))
	}
}
