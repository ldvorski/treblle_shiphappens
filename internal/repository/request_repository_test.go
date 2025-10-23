package repository

import (
	"strings"
	"testing"
	"treblle_project/internal/testutil"
)

// Test 1:Basic Create
func TestRequestRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	repo := NewRequestRepository(db)

	// Create a request and verify it was saved
	id := testutil.CreateTestRequest(t, repo, "GET", "/anime/1", 200, 1500)

	if id <= 0 {
		t.Errorf("Expected positive ID, got %d", id)
	}

	//Verify it was saved
	saved, err := repo.GetByID(int(id))
	if err != nil {
		t.Fatalf("Failed to retrieve saved request: %v", err)
	}

	if saved == nil {
		t.Fatal("Expected request to be saved, got nil")
	}

	if saved.Method != "GET" {
		t.Errorf("Expected method GET, got %s", saved.Method)
	}
	if saved.Path != "/anime/1" {
		t.Errorf("Expected path /anime/1, got %s", saved.Path)
	}
}

// Test 2: Filter by Method
func TestRequestRepository_FilterByMethod(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewRequestRepository(db)

	//Create test data
	testutil.CreateTestRequest(t, repo, "GET", "/test1", 200, 100)
	testutil.CreateTestRequest(t, repo, "POST", "/test2", 200, 200)
	testutil.CreateTestRequest(t, repo, "GET", "/test3", 200, 300)

	//Filter by GET
	filters := RequestFilters{Method: "GET", Limit: 100}
	results, err := repo.List(filters)
	if err != nil {
		t.Fatalf("Failed to list requests: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 GET requests, got %d", len(results))
	}

	for _, r := range results {
		if r.Method != "GET" {
			t.Errorf("Expected method GET, got %s", r.Method)
		}
	}
}

// Test 3: Search functionality
func TestRequestRepository_Search(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewRequestRepository(db)

	// Create test data
	testutil.CreateTestRequest(t, repo, "GET", "/anime/1", 200, 100)
	testutil.CreateTestRequest(t, repo, "GET", "/manga/1", 200, 200)
	testutil.CreateTestRequest(t, repo, "GET", "/anime/characters", 200, 300)

	// Search for "anime"
	filters := RequestFilters{Search: "anime", Limit: 100}
	results, err := repo.List(filters)
	if err != nil {
		t.Fatalf("Failed to search requests: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 anime results, got %d", len(results))
	}

	for _, r := range results {
		if !strings.Contains(r.Path, "anime") {
			t.Errorf("Expected path to contain 'anime', got %s", r.Path)
		}
	}
}
