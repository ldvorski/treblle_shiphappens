package handlers

import (
	"bytes"
	"encoding/csv"
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

// ListProblems godoc
// @Summary      List detected failed or problematic API calls
// @Description  Get a list of detected failed or problematic API calls with optional filtering, ordering and searching
// @Tags         problems, list, order, search, filter
// @Accept       json
// @Produce      json
// @Param        method         query    string  false  "HTTP method filter (GET, POST, etc.)"
// @Param        response       query    int     false  "Response status code filter"
// @Param        min_time       query    int     false  "Minimum response time in milliseconds"
// @Param        max_time       query    int     false  "Maximum response time in milliseconds"
// @Param        created_after  query    string  false  "Filter problems created after date (format: 2006-01-02)"
// @Param        created_before query    string  false  "Filter problems created before date (format: 2006-01-02)"
// @Param        search         query    string  false  "Search in request path"
// @Param        sort           query    string  false  "Sort by field (response_time, created_at)"
// @Param        limit          query    int     false  "Maximum number of results (default: 100)"
// @Param        offset         query    int     false  "Number of results to skip (default: 0)"
// @Success      200  {object}  map[string]interface{}  "List of problems with metadata"
// @Failure      500  {object}  map[string]string       "Internal server error"
// @Router       /problems [get]
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
			"count":  len(problems),
			"limit":  filters.Limit,
			"offset": filters.Offset,
		},
	})
}

// TableView godoc
// @Summary      Table of failed or problematic API calls
// @Description  Get an ordered table of failed or problematic external API calls, optional filtering, ordering and searching, intended for further processing
// @Tags         problems, order, filter, search, table
// @Accept       json
// @Produce      json
// @Param        method         query    string  false  "HTTP method filter (GET, POST, etc.)"
// @Param        response       query    int     false  "Response status code filter"
// @Param        min_time       query    int     false  "Minimum response time in milliseconds"
// @Param        max_time       query    int     false  "Maximum response time in milliseconds"
// @Param        created_after  query    string  false  "Filter problems created after date (format: 2006-01-02)"
// @Param        created_before query    string  false  "Filter problems created before date (format: 2006-01-02)"
// @Param        search         query    string  false  "Search in request path"
// @Param        sort           query    string  false  "Sort by field (response_time, created_at)"
// @Param        limit          query    int     false  "Maximum number of results (default: 100)"
// @Param        offset         query    int     false  "Number of results to skip (default: 0)"
// @Success      200  {object}  map[string]interface{}  "Table data with columns and rows"
// @Failure      500  {object}  map[string]string       "Internal server error"
// @Router       /problems/table [get]
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
			"count":  len(tableData),
			"limit":  filters.Limit,
			"offset": filters.Offset,
		},
	})
}

// CSVExport godoc
// @Summary      Export failed and problematic API calls as CSV
// @Description  Download detected failed or problematic API calls as a CSV file, optional filtering, ordering and searching, ready for use or storing
// @Tags         problems, csv, download, search, filter, order
// @Accept       json
// @Produce      text/csv
// @Param        method         query    string  false  "HTTP method filter (GET, POST, etc.)"
// @Param        response       query    int     false  "Response status code filter"
// @Param        min_time       query    int     false  "Minimum response time in milliseconds"
// @Param        max_time       query    int     false  "Maximum response time in milliseconds"
// @Param        created_after  query    string  false  "Filter problems created after date (format: 2006-01-02)"
// @Param        created_before query    string  false  "Filter problems created before date (format: 2006-01-02)"
// @Param        search         query    string  false  "Search in request path"
// @Param        sort           query    string  false  "Sort by field (response_time, created_at)"
// @Param        limit          query    int     false  "Maximum number of results (default: 100)"
// @Param        offset         query    int     false  "Number of results to skip (default: 0)"
// @Success      200  {string}  string  "CSV file download"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /problems/csv [get]
func (h *ProblemHandler) CSVExport(c *gin.Context) {
	filters := parseProblemFilters(c)
	problems, err := h.repo.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create CSV writer
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	writer.Write([]string{"problem_type", "description", "method", "response", "path", "response_time", "threshold_ms", "created_at"})

	// Write rows
	for _, p := range problems {
		writer.Write([]string{
			p.ProblemType,
			p.Description,
			p.Method,
			strconv.Itoa(p.ResponseStatus),
			p.Path,
			strconv.FormatInt(p.ResponseTimeMs, 10),
			strconv.FormatInt(p.ThresholdMs, 10),
			p.CreatedAt.Format(time.RFC3339),
		})
	}

	// Flush to ensure all data is written
	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate CSV"})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=requests.csv")
	c.String(http.StatusOK, buf.String())
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
