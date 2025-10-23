package models

import "time"

type Problem struct {
	ID          int       `json:"id" db:"id"`
	RequestID   int       `json:"request_id" db:"request_id"`
	ProblemType string    `json:"problem_type" db:"problem_type"`
	Description string    `json:"description" db:"description"`
	ThresholdMs int64     `json:"threshold_ms" db:"threshold_ms"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`

	// Joined fields from api_requests
	Method         string `json:"method,omitempty" db:"method"`
	Path           string `json:"path,omitempty" db:"path"`
	ResponseStatus int    `json:"response,omitempty" db:"response_status"`
	ResponseTimeMs int64  `json:"response_time,omitempty" db:"response_time_ms"`
}
