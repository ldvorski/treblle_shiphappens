package repository

import (
	"testing"
	"time"
	"treblle_project/internal/models"
)

// createTestRequest creates a test API request and returns its ID
func createTestRequest(t *testing.T, repo *RequestRepository, method, path string, status int, responseTime int64) int64 {
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
func createTestProblem(t *testing.T, repo *ProblemRepository, requestID int, problemType, description string, threshold int64) int64 {
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
