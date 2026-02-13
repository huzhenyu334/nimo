package handler

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bitfantasy/nimo/internal/plm/testutil"
	"github.com/gin-gonic/gin"
)

func setupUploadTest(t *testing.T) *gin.Engine {
	t.Helper()
	router := testutil.SetupRouter()
	handler := NewUploadHandler()
	api := testutil.AuthGroup(router, "/api/v1")
	api.POST("/upload", handler.Upload)
	return router
}

func TestUploadFile(t *testing.T) {
	router := setupUploadTest(t)
	token := testutil.DefaultTestToken()

	// Create multipart form with a test file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files", "test-drawing.pdf")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	io.Copy(part, strings.NewReader("test file content for upload"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/v1/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := testutil.ParseResponse(w)
	if code, ok := resp["code"].(float64); !ok || code != 0 {
		t.Fatalf("Expected code 0, got %v", resp["code"])
	}

	data, ok := resp["data"].([]interface{})
	if !ok || len(data) != 1 {
		t.Fatalf("Expected data array with 1 element, got %v", resp["data"])
	}

	file := data[0].(map[string]interface{})
	if file["id"] == nil || file["id"] == "" {
		t.Error("Expected non-empty id")
	}
	if file["filename"] != "test-drawing.pdf" {
		t.Errorf("Expected filename test-drawing.pdf, got %v", file["filename"])
	}
	if file["url"] == nil || file["url"] == "" {
		t.Error("Expected non-empty url")
	}
}

func TestUploadSTPFile(t *testing.T) {
	router := setupUploadTest(t)
	token := testutil.DefaultTestToken()

	// Create multipart form with an STP file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files", "model.stp")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	io.Copy(part, strings.NewReader("fake stp data"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/v1/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := testutil.ParseResponse(w)
	data := resp["data"].([]interface{})
	file := data[0].(map[string]interface{})

	if file["filename"] != "model.stp" {
		t.Errorf("Expected filename model.stp, got %v", file["filename"])
	}

	// thumbnail_url may or may not be present depending on whether stp-thumbnail service is running
	// If present, it should be a valid path
	if thumbURL, ok := file["thumbnail_url"].(string); ok && thumbURL != "" {
		if !strings.HasPrefix(thumbURL, "/uploads/thumbnails/") {
			t.Errorf("Expected thumbnail_url to start with /uploads/thumbnails/, got %v", thumbURL)
		}
		if !strings.HasSuffix(thumbURL, ".svg") {
			t.Errorf("Expected thumbnail_url to end with .svg, got %v", thumbURL)
		}
		t.Logf("STP thumbnail generated: %s", thumbURL)
	} else {
		t.Log("STP thumbnail service not running, thumbnail_url not set (expected)")
	}
}

func TestUploadNoFile(t *testing.T) {
	router := setupUploadTest(t)
	token := testutil.DefaultTestToken()

	// Send empty multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/v1/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUploadMultipleFiles(t *testing.T) {
	router := setupUploadTest(t)
	token := testutil.DefaultTestToken()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for i := 0; i < 3; i++ {
		part, _ := writer.CreateFormFile("files", fmt.Sprintf("file%d.txt", i))
		io.Copy(part, strings.NewReader(fmt.Sprintf("content %d", i)))
	}
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/v1/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := testutil.ParseResponse(w)
	data := resp["data"].([]interface{})
	if len(data) != 3 {
		t.Errorf("Expected 3 files, got %d", len(data))
	}
}
