package models

import "time"

type APIRequest struct {
	ID             int       `json:"id" db:"id"`
	Method         string    `json:"method" db:"method"`
	Path           string    `json:"path" db:"path"`
	ResponseStatus int       `json:"response" db:"response_status"`
	ResponseTimeMs int64     `json:"response_time" db:"response_time_ms"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
