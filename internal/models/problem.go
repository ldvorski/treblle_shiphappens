package models

import "time"

// Problem represents a detected API issue (404, slow response, etc.)
type Problem struct {
	ID          int       `json:"id" db:"id" example:"1"`                                                           // Unique identifier
	RequestID   int       `json:"request_id" db:"request_id" example:"5"`                                           // Related request ID
	ProblemType string    `json:"problem_type" db:"problem_type" example:"not_found"`                               // Type of problem (not_found, slow_response, forbidden, etc.)
	Description string    `json:"description" db:"description" example:"The requested resource could not be found"` // Human-readable description
	ThresholdMs int64     `json:"threshold_ms" db:"threshold_ms" example:"400"`                                     // Threshold that triggered this problem (for slow_response)
	CreatedAt   time.Time `json:"created_at" db:"created_at" example:"2024-01-15T10:30:00Z"`                        // When the problem was detected

	// Joined fields from api_requests
	Method         string `json:"method,omitempty" db:"method" example:"GET"`                  // HTTP method from related request
	Path           string `json:"path,omitempty" db:"path" example:"/anime/999"`               // Request path from related request
	ResponseStatus int    `json:"response,omitempty" db:"response_status" example:"404"`       // Response status from related request
	ResponseTimeMs int64  `json:"response_time,omitempty" db:"response_time_ms" example:"150"` // Response time from related request
}
