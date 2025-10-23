package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"treblle_project/internal/repository"
	"treblle_project/internal/testutil"

	"github.com/gin-gonic/gin"
)

// Test 1: ListRequests returns JSON with data
func TestListRequests_ReturnsJSON(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	repo := repository.NewRequestRepository(db)
	handler := NewRequestHandler(repo)

	// Create test data
	testutil.CreateTestRequest(t, repo, "GET", "/anime/1", 200, 150)
	testutil.CreateTestRequest(t, repo, "POST", "/anime/2", 201, 250)

	// Setup Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/requests", handler.ListRequests)

	// Make request
	req := httptest.NewRequest("GET", "/api/requests", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Check data field exists
	data, ok := response["data"].([]interface{})
	if !ok {
		t.Fatalf("Expected 'data' field in response")
	}

	if len(data) != 2 {
		t.Errorf("Expected 2 requests in data, got %d", len(data))
	}

	// Check meta field
	meta, ok := response["meta"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected 'meta' field in response")
	}

	if count, ok := meta["count"].(float64); !ok || count != 2 {
		t.Errorf("Expected count=2 in meta, got %v", meta["count"])
	}
}

// Test 2: Filter by method
func TestListRequests_FilterByMethod(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	repo := repository.NewRequestRepository(db)
	handler := NewRequestHandler(repo)

	// Create test data
	testutil.CreateTestRequest(t, repo, "GET", "/test1", 200, 100)
	testutil.CreateTestRequest(t, repo, "POST", "/test2", 200, 200)
	testutil.CreateTestRequest(t, repo, "GET", "/test3", 200, 300)

	// Setup Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/requests", handler.ListRequests)

	// Make request with method filter
	req := httptest.NewRequest("GET", "/api/requests?method=GET", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].([]interface{})
	if len(data) != 2 {
		t.Errorf("Expected 2 GET requests, got %d", len(data))
	}

	// Verify all results are GET
	for _, item := range data {
		req := item.(map[string]interface{})
		if req["method"] != "GET" {
			t.Errorf("Expected method GET, got %v", req["method"])
		}
	}
}

// Test 3: Search functionality
func TestListRequests_Search(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	repo := repository.NewRequestRepository(db)
	handler := NewRequestHandler(repo)

	// Create test data
	testutil.CreateTestRequest(t, repo, "GET", "/anime/1", 200, 100)
	testutil.CreateTestRequest(t, repo, "GET", "/manga/1", 200, 200)
	testutil.CreateTestRequest(t, repo, "GET", "/anime/characters", 200, 300)

	// Setup Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/requests", handler.ListRequests)

	// Search for "anime"
	req := httptest.NewRequest("GET", "/api/requests?search=anime", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].([]interface{})
	if len(data) != 2 {
		t.Errorf("Expected 2 anime results, got %d", len(data))
	}
}

// Test 4: TableView returns columns and rows
func TestTableView_ReturnsStructuredData(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	repo := repository.NewRequestRepository(db)
	handler := NewRequestHandler(repo)

	// Create test data
	testutil.CreateTestRequest(t, repo, "GET", "/test", 200, 150)

	// Setup Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/requests/table", handler.TableView)

	// Make request
	req := httptest.NewRequest("GET", "/api/requests/table", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check columns field
	columns, ok := response["columns"].([]interface{})
	if !ok {
		t.Fatalf("Expected 'columns' field in response")
	}

	expectedColumns := []string{"method", "response", "path", "response_time", "created_at"}
	if len(columns) != len(expectedColumns) {
		t.Errorf("Expected %d columns, got %d", len(expectedColumns), len(columns))
	}

	// Check rows field
	rows, ok := response["rows"].([]interface{})
	if !ok {
		t.Fatalf("Expected 'rows' field in response")
	}

	if len(rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(rows))
	}

	// Verify row structure (should be an array)
	row := rows[0].([]interface{})
	if len(row) != 5 {
		t.Errorf("Expected 5 values in row, got %d", len(row))
	}

	// Verify first value is the method
	if row[0] != "GET" {
		t.Errorf("Expected first value to be 'GET', got %v", row[0])
	}
}

// Test 5: CSV Export returns CSV format
func TestCSVExport_ReturnsCSV(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	repo := repository.NewRequestRepository(db)
	handler := NewRequestHandler(repo)

	// Create test data
	testutil.CreateTestRequest(t, repo, "GET", "/test", 200, 150)

	// Setup Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/requests/csv", handler.CSVExport)

	// Make request
	req := httptest.NewRequest("GET", "/api/requests/csv", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/csv" {
		t.Errorf("Expected Content-Type 'text/csv', got %s", contentType)
	}

	// Check CSV content
	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Expected CSV content, got empty body")
	}

	// Verify header line exists
	if body[:6] != "method" {
		t.Errorf("Expected CSV to start with header, got: %s", body[:20])
	}
}

// Test 6: Sorting by response time
func TestListRequests_SortByResponseTime(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	repo := repository.NewRequestRepository(db)
	handler := NewRequestHandler(repo)

	// Create test data with different response times
	testutil.CreateTestRequest(t, repo, "GET", "/slow", 200, 300)
	testutil.CreateTestRequest(t, repo, "GET", "/fast", 200, 100)
	testutil.CreateTestRequest(t, repo, "GET", "/medium", 200, 200)

	// Setup Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/requests", handler.ListRequests)

	// Make request with sort parameter
	req := httptest.NewRequest("GET", "/api/requests?sort=response_time", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].([]interface{})
	if len(data) != 3 {
		t.Errorf("Expected 3 results, got %d", len(data))
	}

	// Verify descending order (slowest first)
	first := data[0].(map[string]interface{})
	if first["response_time"].(float64) != 300 {
		t.Errorf("Expected first result to have response_time 300, got %v", first["response_time"])
	}

	last := data[2].(map[string]interface{})
	if last["response_time"].(float64) != 100 {
		t.Errorf("Expected last result to have response_time 100, got %v", last["response_time"])
	}
}
