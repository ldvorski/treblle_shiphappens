package testutil

import (
	"testing"
	"treblle_project/internal/database"
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
