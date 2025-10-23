package repository

import (
	"fmt"
	"strings"
	"time"
	"treblle_project/internal/database"
	"treblle_project/internal/models"
)

type ProblemRepository struct {
	db *database.DB
}

func NewProblemRepository(db *database.DB) *ProblemRepository {
	return &ProblemRepository{db: db}
}

type ProblemFilters struct {
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

func (r *ProblemRepository) Create(problem *models.Problem) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO problems (request_id, problem_type, description, threshold_ms, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		problem.RequestID, problem.ProblemType, problem.Description, problem.ThresholdMs, problem.CreatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create problem: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (r *ProblemRepository) List(filters ProblemFilters) ([]models.Problem, error) {
	query := `
		SELECT 
			p.id, p.request_id, p.problem_type, p.description, p.threshold_ms, p.created_at,
			r.method, r.path, r.response_status, r.response_time_ms
		FROM problems p
		INNER JOIN api_requests r ON p.request_id = r.id
	`
	where := []string{}
	args := []interface{}{}

	if filters.Method != "" {
		where = append(where, "r.method = ?")
		args = append(args, filters.Method)
	}

	if filters.Response > 0 {
		where = append(where, "r.response_status = ?")
		args = append(args, filters.Response)
	}

	if filters.MinTime > 0 {
		where = append(where, "r.response_time_ms >= ?")
		args = append(args, filters.MinTime)
	}

	if filters.MaxTime > 0 {
		where = append(where, "r.response_time_ms <= ?")
		args = append(args, filters.MaxTime)
	}

	if !filters.CreatedAfter.IsZero() {
		where = append(where, "p.created_at >= ?")
		args = append(args, filters.CreatedAfter)
	}

	if !filters.CreatedBefore.IsZero() {
		where = append(where, "p.created_at <= ?")
		args = append(args, filters.CreatedBefore)
	}

	if filters.Search != "" {
		where = append(where, "r.path LIKE ?")
		args = append(args, "%"+filters.Search+"%")
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	// Sorting
	var sortBy string
	switch filters.SortBy {
	case "response_time":
		sortBy = "r.response_time_ms DESC"
	case "created_at":
		sortBy = "p.created_at DESC"
	default:
		sortBy = "p.created_at DESC"
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
		return nil, fmt.Errorf("failed to query problems: %w", err)
	}
	defer rows.Close()

	var problems []models.Problem
	for rows.Next() {
		var p models.Problem
		err := rows.Scan(
			&p.ID,
			&p.RequestID,
			&p.ProblemType,
			&p.Description,
			&p.ThresholdMs,
			&p.CreatedAt,
			&p.Method,
			&p.Path,
			&p.ResponseStatus,
			&p.ResponseTimeMs,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan problem: %w", err)
		}
		problems = append(problems, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return problems, nil
}
