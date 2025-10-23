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

const SlowResponseThresholdMs = 400

type JikanHandler struct {
	jikanClient jikan.JikanClient
	requestRepo *repository.RequestRepository
	problemRepo *repository.ProblemRepository
}

func NewJikanHandler(
	jikanClient jikan.JikanClient,
	requestRepo *repository.RequestRepository,
	problemRepo *repository.ProblemRepository,
) *JikanHandler {
	return &JikanHandler{
		jikanClient: jikanClient,
		requestRepo: requestRepo,
		problemRepo: problemRepo,
	}
}

// ProxyRequest godoc
// @Summary      Proxy request to Jikan API
// @Description  Forwards requests to the Jikan API, logs metrics, and detects problems (404, 403, 400, slow responses, etc.). Returns the proxied response with the same status code from Jikan.
// @Tags         jikan
// @Accept       json
// @Produce      json
// @Param        path  path  string  true  "Jikan API path (e.g., /anime/1, /manga/2)"
// @Success      200  {object}  map[string]interface{}  "Successfully proxied response from Jikan API (returns whatever status Jikan returns: 200, 404, etc.)"
// @Failure      500  {object}  map[string]interface{}  "Failed to log request to database (request not recorded)"
// @Failure      502  {object}  map[string]interface{}  "Failed to fetch from Jikan API due to network error (request is still logged with status 0)"
// @Router       /jikan/{path} [get]
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
	var problem *models.Problem
	switch metrics.ResponseStatus {
	case 400:
		problem = &models.Problem{
			RequestID:   int(requestID),
			ProblemType: "bad_request",
			Description: "The server cannot or will not process the request.",
			ThresholdMs: 0,
			CreatedAt:   time.Now(),
		}
	case 403:
		problem = &models.Problem{
			RequestID:   int(requestID),
			ProblemType: "forbidden",
			Description: "The request was valid, but the server is refusing action.",
			ThresholdMs: 0,
			CreatedAt:   time.Now(),
		}
	case 404:
		problem = &models.Problem{
			RequestID:   int(requestID),
			ProblemType: "not_found",
			Description: "The requested resource could not be found.",
			ThresholdMs: 0,
			CreatedAt:   time.Now(),
		}
	case 418:
		problem = &models.Problem{
			RequestID:   int(requestID),
			ProblemType: "im_a_teapot",
			Description: "The server is literally a teapot",
			ThresholdMs: 0,
			CreatedAt:   time.Now(),
		}

	}

	if problem == nil && metrics.ResponseTimeMs >= SlowResponseThresholdMs {
		problem = &models.Problem{
			RequestID:   int(requestID),
			ProblemType: "slow_response",
			Description: fmt.Sprintf("Response time (%dms) exceeded threshold (%dms)", metrics.ResponseTimeMs, SlowResponseThresholdMs),
			ThresholdMs: SlowResponseThresholdMs,
			CreatedAt:   time.Now(),
		}
	}
	if problem != nil {
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
