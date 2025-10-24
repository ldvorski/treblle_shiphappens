package models

import "time"

// APIRequest represents a logged API request with response metrics
type APIRequest struct {
	ID             int       `json:"id" db:"id" example:"1"`                                    // Unique identifier
	Method         string    `json:"method" db:"method" example:"GET"`                          // HTTP method
	Path           string    `json:"path" db:"path" example:"/anime/1"`                         // Request path
	ResponseStatus int       `json:"response" db:"response_status" example:"200"`               // HTTP response status code
	ResponseTimeMs int64     `json:"response_time" db:"response_time_ms" example:"150"`         // Response time in milliseconds
	CreatedAt      time.Time `json:"created_at" db:"created_at" example:"2024-01-15T10:30:00Z"` // When the request was logged
}
