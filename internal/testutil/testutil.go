package testutil

import (
	"testing"
	"time"
	"treblle_project/internal/database"
	"treblle_project/internal/models"
	"treblle_project/internal/repository"
)

// SetupTestDB creates an in-memory SQLite database for testing
// This can be used across all test packages
func SetupTestDB(t *testing.T) *database.DB {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
	return db
}

// CreateTestRequest creates a test API request and returns its ID
func CreateTestRequest(t *testing.T, repo *repository.RequestRepository, method, path string, status int, responseTime int64) int64 {
	id, err := repo.Create(&models.APIRequest{
		Method:         method,
		Path:           path,
		ResponseStatus: status,
		ResponseTimeMs: responseTime,
		CreatedAt:      time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}
	return id
}

// CreateTestProblem creates a test problem and returns its ID
func CreateTestProblem(t *testing.T, repo *repository.ProblemRepository, requestID int, problemType, description string, threshold int64) int64 {
	id, err := repo.Create(&models.Problem{
		RequestID:   requestID,
		ProblemType: problemType,
		Description: description,
		ThresholdMs: threshold,
		CreatedAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test problem: %v", err)
	}
	return id
}
