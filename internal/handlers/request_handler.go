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
