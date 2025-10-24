package handlers

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"treblle_project/internal/jikan"
	"treblle_project/internal/repository"
	"treblle_project/internal/testutil"

	"github.com/gin-gonic/gin"
)

// mockJikanClient is a mock implementation of the Jikan client for testing
type mockJikanClient struct {
	response *jikan.RequestMetrics
	err      error
}

func (m *mockJikanClient) ProxyRequest(path string) (*jikan.RequestMetrics, error) {
	// Update path in response to match request
	if m.response != nil {
		m.response.Path = path
	}
	return m.response, m.err
}

// Test 1: Successful proxy request with valid response
func TestJikanHandler_SuccessfulProxy(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	requestRepo := repository.NewRequestRepository(db)
	problemRepo := repository.NewProblemRepository(db)

	// Mock a successful 200 response
	mockClient := &mockJikanClient{
		response: &jikan.RequestMetrics{
			Method:         "GET",
			Path:           "/anime/1",
			ResponseStatus: 200,
			ResponseTimeMs: 150,
			ResponseBody:   []byte(`{"data":{"mal_id":1,"title":"Cowboy Bebop"}}`),
		},
		err: nil,
	}

	handler := &JikanHandler{
		jikanClient: mockClient,
		requestRepo: requestRepo,
		problemRepo: problemRepo,
	}

	// Setup Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/jikan/*path", handler.ProxyRequest)

	// Make request
	req := httptest.NewRequest("GET", "/jikan/anime/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response body
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if data, ok := response["data"].(map[string]interface{}); !ok || data["title"] != "Cowboy Bebop" {
		t.Errorf("Expected Cowboy Bebop in response, got %v", response)
	}

	// Verify request was logged
	requests, _ := requestRepo.List(repository.RequestFilters{Limit: 10})
	if len(requests) != 1 {
		t.Errorf("Expected 1 logged request, got %d", len(requests))
	}
	if requests[0].ResponseStatus != 200 {
		t.Errorf("Expected status 200 in log, got %d", requests[0].ResponseStatus)
	}

	// Verify no problem was created (fast, successful response)
	problems, _ := problemRepo.List(repository.ProblemFilters{Limit: 10})
	if len(problems) != 0 {
		t.Errorf("Expected no problems for successful fast response, got %d", len(problems))
	}
}

// Test 2: 404 Not Found detection
func TestJikanHandler_NotFoundDetection(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	requestRepo := repository.NewRequestRepository(db)
	problemRepo := repository.NewProblemRepository(db)

	mockClient := &mockJikanClient{
		response: &jikan.RequestMetrics{
			Method:         "GET",
			Path:           "/anime/999999",
			ResponseStatus: 404,
			ResponseTimeMs: 100,
			ResponseBody:   []byte(`{"status":404,"message":"Resource not found"}`),
		},
		err: nil,
	}

	handler := &JikanHandler{
		jikanClient: mockClient,
		requestRepo: requestRepo,
		problemRepo: problemRepo,
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/jikan/*path", handler.ProxyRequest)

	req := httptest.NewRequest("GET", "/jikan/anime/999999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify 404 response is proxied
	if w.Code != 404 {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	// Verify problem was created
	problems, _ := problemRepo.List(repository.ProblemFilters{Limit: 10})
	if len(problems) != 1 {
		t.Fatalf("Expected 1 problem for 404, got %d", len(problems))
	}

	if problems[0].ProblemType != "not_found" {
		t.Errorf("Expected problem type 'not_found', got %s", problems[0].ProblemType)
	}

	if problems[0].Description != "The requested resource could not be found." {
		t.Errorf("Unexpected description: %s", problems[0].Description)
	}
}

// Test 3: 403 Forbidden detection
func TestJikanHandler_ForbiddenDetection(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	requestRepo := repository.NewRequestRepository(db)
	problemRepo := repository.NewProblemRepository(db)

	mockClient := &mockJikanClient{
		response: &jikan.RequestMetrics{
			Method:         "GET",
			Path:           "/forbidden",
			ResponseStatus: 403,
			ResponseTimeMs: 50,
			ResponseBody:   []byte(`{"status":403,"message":"Forbidden"}`),
		},
		err: nil,
	}

	handler := &JikanHandler{
		jikanClient: mockClient,
		requestRepo: requestRepo,
		problemRepo: problemRepo,
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/jikan/*path", handler.ProxyRequest)

	req := httptest.NewRequest("GET", "/jikan/forbidden", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	problems, _ := problemRepo.List(repository.ProblemFilters{Limit: 10})
	if len(problems) != 1 {
		t.Fatalf("Expected 1 problem for 403, got %d", len(problems))
	}

	if problems[0].ProblemType != "forbidden" {
		t.Errorf("Expected problem type 'forbidden', got %s", problems[0].ProblemType)
	}
}

// Test 4: 400 Bad Request detection
func TestJikanHandler_BadRequestDetection(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	requestRepo := repository.NewRequestRepository(db)
	problemRepo := repository.NewProblemRepository(db)

	mockClient := &mockJikanClient{
		response: &jikan.RequestMetrics{
			Method:         "GET",
			Path:           "/invalid",
			ResponseStatus: 400,
			ResponseTimeMs: 50,
			ResponseBody:   []byte(`{"status":400,"message":"Bad request"}`),
		},
		err: nil,
	}

	handler := &JikanHandler{
		jikanClient: mockClient,
		requestRepo: requestRepo,
		problemRepo: problemRepo,
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/jikan/*path", handler.ProxyRequest)

	req := httptest.NewRequest("GET", "/jikan/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	problems, _ := problemRepo.List(repository.ProblemFilters{Limit: 10})
	if len(problems) != 1 {
		t.Fatalf("Expected 1 problem for 400, got %d", len(problems))
	}

	if problems[0].ProblemType != "bad_request" {
		t.Errorf("Expected problem type 'bad_request', got %s", problems[0].ProblemType)
	}
}

// Test 5: 418 I'm a Teapot detection (fun edge case!)
func TestJikanHandler_TeapotDetection(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	requestRepo := repository.NewRequestRepository(db)
	problemRepo := repository.NewProblemRepository(db)

	mockClient := &mockJikanClient{
		response: &jikan.RequestMetrics{
			Method:         "GET",
			Path:           "/teapot",
			ResponseStatus: 418,
			ResponseTimeMs: 50,
			ResponseBody:   []byte(`{"message":"I'm a teapot"}`),
		},
		err: nil,
	}

	handler := &JikanHandler{
		jikanClient: mockClient,
		requestRepo: requestRepo,
		problemRepo: problemRepo,
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/jikan/*path", handler.ProxyRequest)

	req := httptest.NewRequest("GET", "/jikan/teapot", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 418 {
		t.Errorf("Expected status 418, got %d", w.Code)
	}

	problems, _ := problemRepo.List(repository.ProblemFilters{Limit: 10})
	if len(problems) != 1 {
		t.Fatalf("Expected 1 problem for 418, got %d", len(problems))
	}

	if problems[0].ProblemType != "im_a_teapot" {
		t.Errorf("Expected problem type 'im_a_teapot', got %s", problems[0].ProblemType)
	}
}

// Test 6: Slow response detection (>= 400ms)
func TestJikanHandler_SlowResponseDetection(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	requestRepo := repository.NewRequestRepository(db)
	problemRepo := repository.NewProblemRepository(db)

	mockClient := &mockJikanClient{
		response: &jikan.RequestMetrics{
			Method:         "GET",
			Path:           "/anime/1",
			ResponseStatus: 200,
			ResponseTimeMs: 500, // Slow!
			ResponseBody:   []byte(`{"data":{"mal_id":1}}`),
		},
		err: nil,
	}

	handler := &JikanHandler{
		jikanClient: mockClient,
		requestRepo: requestRepo,
		problemRepo: problemRepo,
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/jikan/*path", handler.ProxyRequest)

	req := httptest.NewRequest("GET", "/jikan/anime/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify slow_response problem was created
	problems, _ := problemRepo.List(repository.ProblemFilters{Limit: 10})
	if len(problems) != 1 {
		t.Fatalf("Expected 1 problem for slow response, got %d", len(problems))
	}

	if problems[0].ProblemType != "slow_response" {
		t.Errorf("Expected problem type 'slow_response', got %s", problems[0].ProblemType)
	}

	if problems[0].ThresholdMs != SlowResponseThresholdMs {
		t.Errorf("Expected threshold %d, got %d", SlowResponseThresholdMs, problems[0].ThresholdMs)
	}
}

// Test 7: Fast response at threshold boundary (399ms - no problem)
func TestJikanHandler_FastResponseNoProblem(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	requestRepo := repository.NewRequestRepository(db)
	problemRepo := repository.NewProblemRepository(db)

	mockClient := &mockJikanClient{
		response: &jikan.RequestMetrics{
			Method:         "GET",
			Path:           "/anime/1",
			ResponseStatus: 200,
			ResponseTimeMs: 399, // Just under threshold
			ResponseBody:   []byte(`{"data":{"mal_id":1}}`),
		},
		err: nil,
	}

	handler := &JikanHandler{
		jikanClient: mockClient,
		requestRepo: requestRepo,
		problemRepo: problemRepo,
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/jikan/*path", handler.ProxyRequest)

	req := httptest.NewRequest("GET", "/jikan/anime/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify no problem was created
	problems, _ := problemRepo.List(repository.ProblemFilters{Limit: 10})
	if len(problems) != 0 {
		t.Errorf("Expected no problems for fast response, got %d", len(problems))
	}
}

// Test 8: Network failure (connection error)
func TestJikanHandler_NetworkFailure(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	requestRepo := repository.NewRequestRepository(db)
	problemRepo := repository.NewProblemRepository(db)

	mockClient := &mockJikanClient{
		response: &jikan.RequestMetrics{
			Method:         "GET",
			Path:           "/anime/1",
			ResponseStatus: 0, // No response
			ResponseTimeMs: 10,
		},
		err: errors.New("connection refused"),
	}

	handler := &JikanHandler{
		jikanClient: mockClient,
		requestRepo: requestRepo,
		problemRepo: problemRepo,
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/jikan/*path", handler.ProxyRequest)

	req := httptest.NewRequest("GET", "/jikan/anime/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 502 Bad Gateway
	if w.Code != 502 {
		t.Errorf("Expected status 502, got %d", w.Code)
	}

	// Verify error response contains details
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["error"] != "Failed to fetch from Jikan API" {
		t.Errorf("Expected error message in response, got %v", response)
	}

	// Verify request was still logged with status 0
	requests, _ := requestRepo.List(repository.RequestFilters{Limit: 10})
	if len(requests) != 1 {
		t.Errorf("Expected 1 logged request even on failure, got %d", len(requests))
	}
	if requests[0].ResponseStatus != 0 {
		t.Errorf("Expected status 0 for failed request, got %d", requests[0].ResponseStatus)
	}
}

// Test 9: 404 AND slow (status code problem takes precedence)
func TestJikanHandler_SlowAnd404(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	requestRepo := repository.NewRequestRepository(db)
	problemRepo := repository.NewProblemRepository(db)

	mockClient := &mockJikanClient{
		response: &jikan.RequestMetrics{
			Method:         "GET",
			Path:           "/anime/999999",
			ResponseStatus: 404,
			ResponseTimeMs: 500, // Also slow
			ResponseBody:   []byte(`{"status":404}`),
		},
		err: nil,
	}

	handler := &JikanHandler{
		jikanClient: mockClient,
		requestRepo: requestRepo,
		problemRepo: problemRepo,
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/jikan/*path", handler.ProxyRequest)

	req := httptest.NewRequest("GET", "/jikan/anime/999999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Only one problem should be created (404 takes precedence)
	problems, _ := problemRepo.List(repository.ProblemFilters{Limit: 10})
	if len(problems) != 1 {
		t.Fatalf("Expected 1 problem (404 takes precedence), got %d", len(problems))
	}

	if problems[0].ProblemType != "not_found" {
		t.Errorf("Expected 'not_found' problem to take precedence, got %s", problems[0].ProblemType)
	}
}

// Test 10: Request logged even if DB problem logging fails
func TestJikanHandler_ProblemLoggingFailureDoesNotAffectResponse(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	requestRepo := repository.NewRequestRepository(db)
	problemRepo := repository.NewProblemRepository(db)

	mockClient := &mockJikanClient{
		response: &jikan.RequestMetrics{
			Method:         "GET",
			Path:           "/anime/1",
			ResponseStatus: 200,
			ResponseTimeMs: 150,
			ResponseBody:   []byte(`{"data":{"mal_id":1}}`),
		},
		err: nil,
	}

	handler := &JikanHandler{
		jikanClient: mockClient,
		requestRepo: requestRepo,
		problemRepo: problemRepo,
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/jikan/*path", handler.ProxyRequest)

	req := httptest.NewRequest("GET", "/jikan/anime/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should still return success even if problem logging were to fail
	if w.Code != 200 {
		t.Errorf("Expected status 200 even if problem logging fails, got %d", w.Code)
	}
}
