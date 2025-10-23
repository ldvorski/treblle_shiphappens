package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"treblle_project/internal/database"
	"treblle_project/internal/models"
)

type RequestRepository struct {
	db *database.DB
}

func NewRequestRepository(db *database.DB) *RequestRepository {
	return &RequestRepository{db: db}
}

type RequestFilters struct {
	Method        string
	Response      int
	MinTime       int64
	MaxTime       int64
	CreatedAfter  time.Time
	CreatedBefore time.Time
	Search        string
	SortBy        string
	Limit         int
	Offset        int
}

func (r *RequestRepository) Create(req *models.APIRequest) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO api_requests (method, path, response_status, response_time_ms, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		req.Method, req.Path, req.ResponseStatus, req.ResponseTimeMs, req.CreatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (r *RequestRepository) List(filters RequestFilters) ([]models.APIRequest, error) {
	query := "SELECT id, method, path, response_status, response_time_ms, created_at FROM api_requests"
	where := []string{}
	args := []interface{}{}

	if filters.Method != "" {
		where = append(where, "method = ?")
		args = append(args, filters.Method)
	}

	if filters.Response > 0 {
		where = append(where, "response_status = ?")
		args = append(args, filters.Response)
	}

	if filters.MinTime > 0 {
		where = append(where, "response_time_ms >= ?")
		args = append(args, filters.MinTime)
	}

	if filters.MaxTime > 0 {
		where = append(where, "response_time_ms <= ?")
		args = append(args, filters.MaxTime)
	}

	if !filters.CreatedAfter.IsZero() {
		where = append(where, "created_at >= ?")
		args = append(args, filters.CreatedAfter)
	}

	if !filters.CreatedBefore.IsZero() {
		where = append(where, "created_at <= ?")
		args = append(args, filters.CreatedBefore)
	}

	if filters.Search != "" {
		where = append(where, "path LIKE ?")
		args = append(args, "%"+filters.Search+"%")
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	// Sorting
	var sortBy string
	switch filters.SortBy {
	case "response_time":
		sortBy = "response_time_ms DESC"
	case "created_at":
		sortBy = "created_at DESC"
	default:
		sortBy = "created_at DESC"
	}
	query += " ORDER BY " + sortBy

	// Pagination
	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	} else {
		query += " LIMIT 100" // Default limit
	}

	if filters.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filters.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query requests: %w", err)
	}
	defer rows.Close()

	var requests []models.APIRequest
	for rows.Next() {
		var req models.APIRequest
		err := rows.Scan(
			&req.ID,
			&req.Method,
			&req.Path,
			&req.ResponseStatus,
			&req.ResponseTimeMs,
			&req.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan request: %w", err)
		}
		requests = append(requests, req)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return requests, nil
}

func (r *RequestRepository) GetByID(id int) (*models.APIRequest, error) {
	var req models.APIRequest
	err := r.db.QueryRow(
		`SELECT id, method, path, response_status, response_time_ms, created_at
		FROM api_requests WHERE id = ?`,
		id,
	).Scan(&req.ID, &req.Method, &req.Path, &req.ResponseStatus, &req.ResponseTimeMs, &req.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get request: %w", err)
	}

	return &req, nil
}
