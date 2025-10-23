package handlers

import (
	"net/http"
	"strconv"
	"time"
	"treblle_project/internal/repository"

	"github.com/gin-gonic/gin"
)

type ProblemHandler struct {
	repo *repository.ProblemRepository
}

func NewProblemHandler(repo *repository.ProblemRepository) *ProblemHandler {
	return &ProblemHandler{repo: repo}
}

func (h *ProblemHandler) ListProblems(c *gin.Context) {
	filters := parseProblemFilters(c)
	problems, err := h.repo.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": problems,
		"meta": gin.H{
			"limit":  filters.Limit,
			"offset": filters.Offset,
		},
	})
}

func (h *ProblemHandler) TableView(c *gin.Context) {
	filters := parseProblemFilters(c)
	problems, err := h.repo.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format as table structure
	tableData := make([][]any, 0, len(problems))
	for _, p := range problems {
		tableData = append(tableData, []any{
			p.ProblemType,
			p.Description,
			p.Method,
			p.ResponseStatus,
			p.Path,
			p.ResponseTimeMs,
			p.ThresholdMs,
			p.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"columns": []string{"problem_type", "description", "method", "response", "path", "response_time", "threshold_ms", "created_at"},
		"rows":    tableData,
		"meta": gin.H{
			"limit":  filters.Limit,
			"offset": filters.Offset,
		},
	})
}

func parseProblemFilters(c *gin.Context) repository.ProblemFilters {
	filters := repository.ProblemFilters{
		Method: c.Query("method"),
		Search: c.Query("search"),
		SortBy: c.Query("sort"),
		Limit:  100,
		Offset: 0,
	}

	if response := c.Query("response"); response != "" {
		if val, err := strconv.Atoi(response); err == nil {
			filters.Response = val
		}
	}

	if minTime := c.Query("min_time"); minTime != "" {
		if val, err := strconv.ParseInt(minTime, 10, 64); err == nil {
			filters.MinTime = val
		}
	}

	if maxTime := c.Query("max_time"); maxTime != "" {
		if val, err := strconv.ParseInt(maxTime, 10, 64); err == nil {
			filters.MaxTime = val
		}
	}

	if createdAfter := c.Query("created_after"); createdAfter != "" {
		if t, err := time.Parse("2006-01-02", createdAfter); err == nil {
			filters.CreatedAfter = t
		}
	}

	if createdBefore := c.Query("created_before"); createdBefore != "" {
		if t, err := time.Parse("2006-01-02", createdBefore); err == nil {
			filters.CreatedBefore = t
		}
	}

	if limit := c.Query("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil && val > 0 {
			filters.Limit = val
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if val, err := strconv.Atoi(offset); err == nil && val >= 0 {
			filters.Offset = val
		}
	}

	return filters
}
