package repository

import (
	"strings"
	"testing"
	"time"
)

// Test 1: Basic Create
func TestProblemRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	problemRepo := NewProblemRepository(db)

	// Create a problem

	id := createTestProblem(t, db,
		int(createTestRequest(t, db, "GET", "/anime/999", 404, 100)),
		"not_found", "Resource not found", 0)

	if id <= 0 {
		t.Errorf("Expected positive ID, got %d", id)
	}

	// Verify it was saved by listing
	filters := ProblemFilters{Limit: 100}
	problems, err := problemRepo.List(filters)
	if err != nil {
		t.Fatalf("Failed to list problems: %v", err)
	}

	if len(problems) != 1 {
		t.Errorf("Expected 1 problem, got %d", len(problems))
	}

	if problems[0].ProblemType != "not_found" {
		t.Errorf("Expected problem type 'not_found', got %s", problems[0].ProblemType)
	}
}

// Test 2: Filter by Method
func TestProblemRepository_FilterByMethod(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	problemRepo := NewProblemRepository(db)

	// Create test data
	createTestProblem(t, db, int(createTestRequest(t, db, "GET", "/test1", 404, 100)), "not_found", "Not found", 0)
	createTestProblem(t, db, int(createTestRequest(t, db, "POST", "/test2", 404, 200)), "not_found", "Not found", 0)
	createTestProblem(t, db, int(createTestRequest(t, db, "GET", "/test3", 404, 300)), "not_found", "Not found", 0)

	// Filter by GET method
	filters := ProblemFilters{Method: "GET", Limit: 100}
	results, err := problemRepo.List(filters)
	if err != nil {
		t.Fatalf("Failed to list problems: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 GET problems, got %d", len(results))
	}

	for _, p := range results {
		if p.Method != "GET" {
			t.Errorf("Expected method GET, got %s", p.Method)
		}
	}
}

// Test 3: Filter by Response Status
func TestProblemRepository_FilterByResponseStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	problemRepo := NewProblemRepository(db)

	// Create test data with different status codes
	createTestProblem(t, db, int(createTestRequest(t, db, "GET", "/missing", 404, 100)), "not_found", "Not found", 0)
	createTestProblem(t, db, int(createTestRequest(t, db, "GET", "/forbidden", 403, 200)), "forbidden", "Forbidden", 0)

	// Filter by 404 status
	filters := ProblemFilters{Response: 404, Limit: 100}
	results, err := problemRepo.List(filters)
	if err != nil {
		t.Fatalf("Failed to list problems: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 problem with 404 status, got %d", len(results))
	}

	if results[0].ResponseStatus != 404 {
		t.Errorf("Expected response status 404, got %d", results[0].ResponseStatus)
	}
}

// Test 4: Search functionality
func TestProblemRepository_Search(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	problemRepo := NewProblemRepository(db)

	// Create test data with different paths
	createTestProblem(t, db,
		int(createTestRequest(t, db, "GET", "/anime/1", 404, 100)),
		"not_found", "Anime not found", 0)
	createTestProblem(t, db,
		int(createTestRequest(t, db, "GET", "/manga/1", 404, 200)),
		"not_found", "Manga not found", 0)
	createTestProblem(t, db,
		int(createTestRequest(t, db, "GET", "/anime/characters", 404, 300)),
		"not_found", "Characters not found", 0)

	// Search for "anime"
	filters := ProblemFilters{Search: "anime", Limit: 100}
	results, err := problemRepo.List(filters)
	if err != nil {
		t.Fatalf("Failed to search problems: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 anime-related problems, got %d", len(results))
	}

	for _, p := range results {
		if !strings.Contains(p.Path, "anime") {
			t.Errorf("Expected path to contain 'anime', got %s", p.Path)
		}
	}
}

// Test 5: Filter by Response Time
func TestProblemRepository_FilterByResponseTime(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	problemRepo := NewProblemRepository(db)

	// Create test data with different response times
	createTestProblem(t, db,
		int(createTestRequest(t, db, "GET", "/slow", 200, 3000)),
		"slow_response", "Too slow", 2000)
	createTestProblem(t, db,
		int(createTestRequest(t, db, "GET", "/fast", 200, 100)),
		"slow_response", "Fast but logged", 2000)

	// Filter for slow responses (>2000ms)
	filters := ProblemFilters{MinTime: 2000, Limit: 100}
	results, err := problemRepo.List(filters)
	if err != nil {
		t.Fatalf("Failed to list problems: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 slow problem, got %d", len(results))
	}

	if results[0].ResponseTimeMs < 2000 {
		t.Errorf("Expected response time >= 2000, got %d", results[0].ResponseTimeMs)
	}
}

// Test 6: Sort by Response Time
func TestProblemRepository_SortByResponseTime(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	problemRepo := NewProblemRepository(db)

	// Create test data with different response times
	createTestProblem(t, db,
		int(createTestRequest(t, db, "GET", "/test1", 404, 300)),
		"not_found", "Problem 1", 0)
	time.Sleep(10 * time.Millisecond)
	createTestProblem(t, db,
		int(createTestRequest(t, db, "GET", "/test2", 404, 100)),
		"not_found", "Problem 2", 0)
	time.Sleep(10 * time.Millisecond)
	createTestProblem(t, db,
		int(createTestRequest(t, db, "GET", "/test3", 404, 200)),
		"not_found", "Problem 3", 0)

	// Sort by response time (descending)
	filters := ProblemFilters{SortBy: "response_time", Limit: 100}
	results, err := problemRepo.List(filters)
	if err != nil {
		t.Fatalf("Failed to list problems: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Verify descending order
	if results[0].ResponseTimeMs != 300 {
		t.Errorf("Expected first result to be slowest (300ms), got %d", results[0].ResponseTimeMs)
	}

	if results[1].ResponseTimeMs != 200 {
		t.Errorf("Expected second result to be 200ms, got %d", results[1].ResponseTimeMs)
	}

	if results[2].ResponseTimeMs != 100 {
		t.Errorf("Expected third result to be fastest (100ms), got %d", results[2].ResponseTimeMs)
	}
}

// Test 7: Verify Joined Fields
func TestProblemRepository_JoinedFields(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	problemRepo := NewProblemRepository(db)

	// Create test data
	createTestProblem(t, db,
		int(createTestRequest(t, db, "GET", "/anime/404", 404, 150)),
		"not_found", "Anime not found", 0)

	// List and verify joined fields are populated
	filters := ProblemFilters{Limit: 100}
	results, err := problemRepo.List(filters)
	if err != nil {
		t.Fatalf("Failed to list problems: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 problem, got %d", len(results))
	}

	p := results[0]

	// Verify joined fields from request
	if p.Method != "GET" {
		t.Errorf("Expected joined method 'GET', got %s", p.Method)
	}

	if p.Path != "/anime/404" {
		t.Errorf("Expected joined path '/anime/404', got %s", p.Path)
	}

	if p.ResponseStatus != 404 {
		t.Errorf("Expected joined response status 404, got %d", p.ResponseStatus)
	}

	if p.ResponseTimeMs != 150 {
		t.Errorf("Expected joined response time 150, got %d", p.ResponseTimeMs)
	}
}
