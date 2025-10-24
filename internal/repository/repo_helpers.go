package repository

import (
	"testing"
	"time"
	"treblle_project/internal/database"
	"treblle_project/internal/models"
)

// setupTestDB creates an in-memory SQLite database for testing
// This is for repository package tests only
func setupTestDB(t *testing.T) *database.DB {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
	return db
}

// createTestRequest creates a test API request and returns its ID
// This helper is for repository package tests only
func createTestRequest(t *testing.T, db *database.DB, method, path string, status int, responseTime int64) int64 {
	repo := NewRequestRepository(db)
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

// createTestProblem creates a test problem and returns its ID
// This helper is for repository package tests only
func createTestProblem(t *testing.T, db *database.DB, requestID int, problemType, description string, threshold int64) int64 {
	repo := NewProblemRepository(db)
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
