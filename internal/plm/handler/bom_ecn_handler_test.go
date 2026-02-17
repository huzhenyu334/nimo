package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitfantasy/nimo/internal/plm/entity"
	"github.com/bitfantasy/nimo/internal/plm/repository"
	"github.com/bitfantasy/nimo/internal/plm/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupECNTestDB() (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(
		&entity.User{},
		&entity.Project{},
		&entity.ProjectBOM{},
		&entity.ProjectBOMItem{},
		&entity.BOMDraft{},
		&entity.BOMECN{},
	)

	cleanup := func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	return db, cleanup
}

func createTestBOM(db *gorm.DB, status string) *entity.ProjectBOM {
	projectID := uuid.New().String()[:32]
	bomID := uuid.New().String()[:32]
	userID := uuid.New().String()[:32]

	project := &entity.Project{
		ID:          projectID,
		Name:        "Test Project",
		Status:      "active",
		CreatedBy:   userID,
	}
	db.Create(project)

	bom := &entity.ProjectBOM{
		ID:        bomID,
		ProjectID: projectID,
		Name:      "Test BOM",
		BOMType:   "EBOM",
		Version:   "v1.0",
		Status:    status,
		CreatedBy: userID,
	}
	db.Create(bom)

	return bom
}

func TestStartEditing(t *testing.T) {
	db, cleanup := setupECNTestDB()
	defer cleanup()

	bom := createTestBOM(db, "released")

	bomRepo := repository.NewProjectBOMRepository(db)
	draftRepo := repository.NewBOMDraftRepository(db)
	ecnRepo := repository.NewBOMECNRepository(db)
	svc := service.NewBOMECNService(bomRepo, draftRepo, ecnRepo)
	handler := NewBOMECNHandler(svc)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", bom.CreatedBy)
		c.Next()
	})
	router.POST("/api/v1/bom/:id/edit", handler.StartEditing)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/bom/"+bom.ID+"/edit", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "editing", data["status"])
}

func TestSaveDraft(t *testing.T) {
	db, cleanup := setupECNTestDB()
	defer cleanup()

	bom := createTestBOM(db, "editing")

	bomRepo := repository.NewProjectBOMRepository(db)
	draftRepo := repository.NewBOMDraftRepository(db)
	ecnRepo := repository.NewBOMECNRepository(db)
	svc := service.NewBOMECNService(bomRepo, draftRepo, ecnRepo)
	handler := NewBOMECNHandler(svc)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", bom.CreatedBy)
		c.Next()
	})
	router.POST("/api/v1/bom/:id/draft", handler.SaveDraft)

	draftData := service.DraftData{
		Items: []entity.ProjectBOMItem{
			{
				ID:       uuid.New().String()[:32],
				BOMID:    bom.ID,
				Name:     "Test Item",
				Quantity: 10,
				Unit:     "pcs",
			},
		},
	}

	body, _ := json.Marshal(draftData)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/bom/"+bom.ID+"/draft", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
}

func TestSubmitECN(t *testing.T) {
	db, cleanup := setupECNTestDB()
	defer cleanup()

	bom := createTestBOM(db, "editing")

	bomRepo := repository.NewProjectBOMRepository(db)
	draftRepo := repository.NewBOMDraftRepository(db)
	ecnRepo := repository.NewBOMECNRepository(db)
	svc := service.NewBOMECNService(bomRepo, draftRepo, ecnRepo)

	// 先保存草稿
	draftData := &service.DraftData{
		Items: []entity.ProjectBOMItem{
			{
				ID:       uuid.New().String()[:32],
				BOMID:    bom.ID,
				Name:     "Test Item",
				Quantity: 10,
				Unit:     "pcs",
				Category: "electronic",
				SubCategory: "component",
			},
		},
	}
	_, err := svc.SaveDraft(nil, bom.ID, draftData, bom.CreatedBy)
	if err != nil {
		t.Fatalf("Failed to save draft: %v", err)
	}

	handler := NewBOMECNHandler(svc)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", bom.CreatedBy)
		c.Next()
	})
	router.POST("/api/v1/bom/:id/ecn", handler.SubmitECN)

	input := map[string]string{
		"title": "BOM Update ECN",
	}
	body, _ := json.Marshal(input)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/bom/"+bom.ID+"/ecn", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
		t.Logf("Response: %+v", resp)
	}

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, float64(0), resp["code"])

	if resp["data"] != nil {
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "pending", data["status"])
		assert.Contains(t, data["ecn_number"], "ECN-")
	}
}

func TestDiscardDraft(t *testing.T) {
	db, cleanup := setupECNTestDB()
	defer cleanup()

	bom := createTestBOM(db, "editing")

	bomRepo := repository.NewProjectBOMRepository(db)
	draftRepo := repository.NewBOMDraftRepository(db)
	ecnRepo := repository.NewBOMECNRepository(db)
	svc := service.NewBOMECNService(bomRepo, draftRepo, ecnRepo)

	// 先保存草稿
	draftData := &service.DraftData{
		Items: []entity.ProjectBOMItem{},
	}
	svc.SaveDraft(nil, bom.ID, draftData, bom.CreatedBy)

	handler := NewBOMECNHandler(svc)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", bom.CreatedBy)
		c.Next()
	})
	router.DELETE("/api/v1/bom/:id/draft", handler.DiscardDraft)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/bom/"+bom.ID+"/draft", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
}
