package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"treblle_project/internal/repository"
	"treblle_project/internal/testutil"

	"github.com/gin-gonic/gin"
)

// Test 1: ListProblems returns JSON with data
func TestListProblems_ReturnsJSON(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	requestRepo := repository.NewRequestRepository(db)
	problemRepo := repository.NewProblemRepository(db)
	handler := NewProblemHandler(problemRepo)

	// Create test data
	testutil.CreateTestProblem(t, problemRepo,
		int(testutil.CreateTestRequest(t, requestRepo, "GET", "/anime/999", 404, 150)),
		"not_found", "Resource not found", 0)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/problems", handler.ListProblems)

	req := httptest.NewRequest("GET", "/api/problems", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]any
	json.Unmarshal(w.Body.Bytes(), &response)

	data, ok := response["data"].([]any)
	if !ok || len(data) != 1 {
		t.Errorf("Expected 1 problem in data, got %v", response)
	}

	meta, ok := response["meta"].(map[string]any)
	if !ok {
		t.Errorf("Expected meta object, got %v", response["meta"])
	}

	if count, ok := meta["count"].(float64); !ok || count != 1 {
		t.Errorf("Expected count=1 in meta, got %v", meta["count"])
	}
}

// Test 2: Filter by method
func TestListProblems_FilterByMethod(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	requestRepo := repository.NewRequestRepository(db)
	problemRepo := repository.NewProblemRepository(db)
	handler := NewProblemHandler(problemRepo)

	// Create test data
	testutil.CreateTestProblem(t, problemRepo,
		int(testutil.CreateTestRequest(t, requestRepo, "GET", "/test1", 404, 100)),
		"not_found", "Not found", 0)
	testutil.CreateTestProblem(t, problemRepo,
		int(testutil.CreateTestRequest(t, requestRepo, "POST", "/test2", 404, 200)),
		"not_found", "Not found", 0)
	testutil.CreateTestProblem(t, problemRepo,
		int(testutil.CreateTestRequest(t, requestRepo, "GET", "/test3", 404, 300)),
		"not_found", "Not found", 0)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/problems", handler.ListProblems)

	req := httptest.NewRequest("GET", "/api/problems?method=GET", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]any
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].([]any)
	if len(data) != 2 {
		t.Errorf("Expected 2 GET problems, got %d", len(data))
	}

	// Verify all returned problems are GET
	for _, item := range data {
		problem := item.(map[string]any)
		if problem["method"] != "GET" {
			t.Errorf("Expected method GET, got %v", problem["method"])
		}
	}
}

// Test 3: Search functionality
func TestListProblems_Search(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	requestRepo := repository.NewRequestRepository(db)
	problemRepo := repository.NewProblemRepository(db)
	handler := NewProblemHandler(problemRepo)

	// Create test data
	testutil.CreateTestProblem(t, problemRepo,
		int(testutil.CreateTestRequest(t, requestRepo, "GET", "/anime/1", 404, 100)),
		"not_found", "Anime not found", 0)
	testutil.CreateTestProblem(t, problemRepo,
		int(testutil.CreateTestRequest(t, requestRepo, "GET", "/manga/1", 404, 200)),
		"not_found", "Manga not found", 0)
	testutil.CreateTestProblem(t, problemRepo,
		int(testutil.CreateTestRequest(t, requestRepo, "GET", "/anime/characters", 404, 300)),
		"not_found", "Characters not found", 0)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/problems", handler.ListProblems)

	req := httptest.NewRequest("GET", "/api/problems?search=anime", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]any
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].([]any)
	if len(data) != 2 {
		t.Errorf("Expected 2 anime-related problems, got %d", len(data))
	}
}

// Test 4: TableView returns columns and rows
func TestProblemTableView_ReturnsStructuredData(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	requestRepo := repository.NewRequestRepository(db)
	problemRepo := repository.NewProblemRepository(db)
	handler := NewProblemHandler(problemRepo)

	// Create test data
	testutil.CreateTestProblem(t, problemRepo,
		int(testutil.CreateTestRequest(t, requestRepo, "GET", "/anime/404", 404, 150)),
		"not_found", "Anime not found", 0)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/problems/table", handler.TableView)

	req := httptest.NewRequest("GET", "/api/problems/table", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]any
	json.Unmarshal(w.Body.Bytes(), &response)

	// Verify columns
	columns, ok := response["columns"].([]any)
	if !ok {
		t.Errorf("Expected columns array, got %v", response["columns"])
	}

	expectedColumns := []string{"problem_type", "description", "method", "response", "path", "response_time", "threshold_ms", "created_at"}
	if len(columns) != len(expectedColumns) {
		t.Errorf("Expected %d columns, got %d", len(expectedColumns), len(columns))
	}

	// Verify rows
	rows, ok := response["rows"].([]any)
	if !ok || len(rows) != 1 {
		t.Errorf("Expected 1 row, got %v", response["rows"])
	}

	// Verify first row structure (array of values)
	row := rows[0].([]any)
	if len(row) != len(expectedColumns) {
		t.Errorf("Expected row to have %d values, got %d", len(expectedColumns), len(row))
	}

	// Verify first value is the problem type
	if row[0] != "not_found" {
		t.Errorf("Expected first value to be 'not_found', got %v", row[0])
	}
}

// Test 5: CSV Export returns CSV format
func TestProblemCSVExport_ReturnsCSV(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	requestRepo := repository.NewRequestRepository(db)
	problemRepo := repository.NewProblemRepository(db)
	handler := NewProblemHandler(problemRepo)

	// Create test data
	testutil.CreateTestProblem(t, problemRepo,
		int(testutil.CreateTestRequest(t, requestRepo, "GET", "/slow", 200, 3000)),
		"slow_response", "Response too slow", 2000)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/problems/csv", handler.CSVExport)

	req := httptest.NewRequest("GET", "/api/problems/csv", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/csv" {
		t.Errorf("Expected Content-Type text/csv, got %s", contentType)
	}

	// Verify CSV content
	body := w.Body.String()
	if !strings.Contains(body, "problem_type") {
		t.Errorf("Expected CSV to contain header 'problem_type'")
	}

	if !strings.Contains(body, "slow_response") {
		t.Errorf("Expected CSV to contain 'slow_response'")
	}
}

// Test 6: Filter by response status
func TestListProblems_FilterByResponseStatus(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	requestRepo := repository.NewRequestRepository(db)
	problemRepo := repository.NewProblemRepository(db)
	handler := NewProblemHandler(problemRepo)

	// Create test data with different status codes
	testutil.CreateTestProblem(t, problemRepo,
		int(testutil.CreateTestRequest(t, requestRepo, "GET", "/missing", 404, 100)),
		"not_found", "Not found", 0)
	testutil.CreateTestProblem(t, problemRepo,
		int(testutil.CreateTestRequest(t, requestRepo, "GET", "/forbidden", 403, 200)),
		"forbidden", "Forbidden", 0)
	testutil.CreateTestProblem(t, problemRepo,
		int(testutil.CreateTestRequest(t, requestRepo, "GET", "/teapot", 418, 300)),
		"im_a_teapot", "I'm a teapot", 0)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/problems", handler.ListProblems)

	req := httptest.NewRequest("GET", "/api/problems?response=404", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]any
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].([]any)
	if len(data) != 1 {
		t.Errorf("Expected 1 problem with 404 status, got %d", len(data))
	}

	problem := data[0].(map[string]any)
	if problem["response"].(float64) != 404 {
		t.Errorf("Expected response status 404, got %v", problem["response"])
	}
}
