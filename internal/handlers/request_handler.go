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

type RequestHandler struct {
	repo *repository.RequestRepository
}

func NewRequestHandler(repo *repository.RequestRepository) *RequestHandler {
	return &RequestHandler{repo: repo}
}

// ListRequests godoc
// @Summary      List of API requests successfully completed
// @Description  Get a list of logged API requests calls with optional filtering, ordering and searching
// @Tags         requests, filter, order, search, list, successful, completed
// @Accept       json
// @Produce      json
// @Param        method         query    string  false  "HTTP method filter (GET, POST, etc.)"
// @Param        response       query    int     false  "Response status code filter"
// @Param        min_time       query    int     false  "Minimum response time in milliseconds"
// @Param        max_time       query    int     false  "Maximum response time in milliseconds"
// @Param        created_after  query    string  false  "Filter requests created after date (format: 2006-01-02)"
// @Param        created_before query    string  false  "Filter requests created before date (format: 2006-01-02)"
// @Param        search         query    string  false  "Search in request path"
// @Param        sort           query    string  false  "Sort by field (response_time, created_at)"
// @Param        limit          query    int     false  "Maximum number of results (default: 100)"
// @Param        offset         query    int     false  "Number of results to skip (default: 0)"
// @Success      200  {object}  map[string]interface{}  "List of requests with metadata"
// @Failure      500  {object}  map[string]string       "Internal server error"
// @Router       /requests [get]
func (h *RequestHandler) ListRequests(c *gin.Context) {
	filters := parseRequestFilters(c)
	requests, err := h.repo.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": requests,
		"meta": gin.H{
			"count":  len(requests),
			"limit":  filters.Limit,
			"offset": filters.Offset,
		},
	})
}

// TableView godoc
// @Summary      Table of successfully completed API request calls
// @Description  Get successfully completed API request calls formatted for table display after further proccessing with columns and rows, supports ordering, filtering and searching
// @Tags         requests, table, successful, completed, search, filter, order, processing
// @Accept       json
// @Produce      json
// @Param        method         query    string  false  "HTTP method filter (GET, POST, etc.)"
// @Param        response       query    int     false  "Response status code filter"
// @Param        min_time       query    int     false  "Minimum response time in milliseconds"
// @Param        max_time       query    int     false  "Maximum response time in milliseconds"
// @Param        created_after  query    string  false  "Filter requests created after date (format: 2006-01-02)"
// @Param        created_before query    string  false  "Filter requests created before date (format: 2006-01-02)"
// @Param        search         query    string  false  "Search in request path"
// @Param        sort           query    string  false  "Sort by field (response_time, created_at)"
// @Param        limit          query    int     false  "Maximum number of results (default: 100)"
// @Param        offset         query    int     false  "Number of results to skip (default: 0)"
// @Success      200  {object}  map[string]interface{}  "Table data with columns and rows"
// @Failure      500  {object}  map[string]string       "Internal server error"
// @Router       /requests/table [get]
func (h *RequestHandler) TableView(c *gin.Context) {
	filters := parseRequestFilters(c)
	requests, err := h.repo.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format as table structure
	tableData := make([][]any, 0, len(requests))
	for _, req := range requests {
		tableData = append(tableData, []any{
			req.Method,
			req.ResponseStatus,
			req.Path,
			req.ResponseTimeMs,
			req.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"columns": []string{"method", "response", "path", "response_time", "created_at"},
		"rows":    tableData,
		"meta": gin.H{
			"count":  len(tableData),
			"limit":  filters.Limit,
			"offset": filters.Offset,
		},
	})
}

// CSVExport godoc
// @Summary      Export successfully completed API requests as CSV
// @Description  Download successfully completed API requests as a CSV file, supports filtering, ordering and searching, ready for use
// @Tags         requests, successful, completed, csv, filter, order, search
// @Accept       json
// @Produce      text/csv
// @Param        method         query    string  false  "HTTP method filter (GET, POST, etc.)"
// @Param        response       query    int     false  "Response status code filter"
// @Param        min_time       query    int     false  "Minimum response time in milliseconds"
// @Param        max_time       query    int     false  "Maximum response time in milliseconds"
// @Param        created_after  query    string  false  "Filter requests created after date (format: 2006-01-02)"
// @Param        created_before query    string  false  "Filter requests created before date (format: 2006-01-02)"
// @Param        search         query    string  false  "Search in request path"
// @Param        sort           query    string  false  "Sort by field (response_time, created_at)"
// @Param        limit          query    int     false  "Maximum number of results (default: 100)"
// @Param        offset         query    int     false  "Number of results to skip (default: 0)"
// @Success      200  {string}  string  "CSV file download"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /requests/csv [get]
func (h *RequestHandler) CSVExport(c *gin.Context) {
	filters := parseRequestFilters(c)
	requests, err := h.repo.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//Create CSV writer
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	//Build CSV
	writer.Write([]string{"method", "response", "path", "response_time", "created_at"})
	for _, req := range requests {
		writer.Write([]string{
			req.Method,
			strconv.Itoa(req.ResponseStatus),
			req.Path,
			strconv.FormatInt(req.ResponseTimeMs, 10),
			req.CreatedAt.Format(time.RFC3339),
		})
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate CSV"})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=requests.csv")
	c.String(http.StatusOK, buf.String())
}

func parseRequestFilters(c *gin.Context) repository.RequestFilters {
	filters := repository.RequestFilters{
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
