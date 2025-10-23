package handlers

import (
	"fmt"
	"net/http"
	"time"
	"treblle_project/internal/jikan"
	"treblle_project/internal/models"
	"treblle_project/internal/repository"

	"github.com/gin-gonic/gin"
)

const SlowResponseThresholdMs = 2000

type JikanHandler struct {
	jikanClient *jikan.Client
	requestRepo *repository.RequestRepository
	problemRepo *repository.ProblemRepository
}

func NewJikanHandler(
	jikanClient *jikan.Client,
	requestRepo *repository.RequestRepository,
	problemRepo *repository.ProblemRepository,
) *JikanHandler {
	return &JikanHandler{
		jikanClient: jikanClient,
		requestRepo: requestRepo,
		problemRepo: problemRepo,
	}
}

func (h *JikanHandler) ProxyRequest(c *gin.Context) {
	path := c.Param("path")
	if path == "" {
		path = "/"
	}

	// Make request to Jikan API and measure
	metrics, err := h.jikanClient.ProxyRequest(path)

	// Log the request regardless of success/failure
	apiRequest := &models.APIRequest{
		Method:         metrics.Method,
		ResponseStatus: metrics.ResponseStatus,
		Path:           metrics.Path,
		ResponseTimeMs: metrics.ResponseTimeMs,
		CreatedAt:      time.Now(),
	}

	// If request failed completely (no response), set status to 0
	if err != nil && metrics.ResponseStatus == 0 {
		apiRequest.ResponseStatus = 0
	}

	requestID, dbErr := h.requestRepo.Create(apiRequest)
	if dbErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to log request",
			"details": dbErr.Error(),
		})
		return
	}

	// Check if this is a slow response and create a problem
	if metrics.ResponseTimeMs >= SlowResponseThresholdMs {
		problem := &models.Problem{
			RequestID:   int(requestID),
			ProblemType: "slow_response",
			Description: fmt.Sprintf("Response time (%dms) exceeded threshold (%dms)", metrics.ResponseTimeMs, SlowResponseThresholdMs),
			ThresholdMs: SlowResponseThresholdMs,
			CreatedAt:   time.Now(),
		}
		_, _ = h.problemRepo.Create(problem) // Don't fail the request if problem logging fails
	}

	// If the Jikan API request failed, return error
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "Failed to fetch from Jikan API",
			"details": err.Error(),
			"metrics": gin.H{
				"response_time_ms": metrics.ResponseTimeMs,
				"request_id":       requestID,
			},
		})
		return
	}

	// Return the proxied response
	c.Data(metrics.ResponseStatus, "application/json", metrics.ResponseBody)
}
