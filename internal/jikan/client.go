package jikan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const BaseURL = "https://api.jikan.moe/v4"

// JikanClient is an interface for making requests to the Jikan API
// This allows for easy mocking in tests
type JikanClient interface {
	ProxyRequest(path string) (*RequestMetrics, error)
}

type Client struct {
	httpClient *http.Client
	baseURL    string
}

// Ensure Client implements JikanClient
var _ JikanClient = (*Client)(nil)

type RequestMetrics struct {
	Method         string
	Path           string
	ResponseStatus int
	ResponseTimeMs int64
	ResponseBody   []byte
	Error          error
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: BaseURL,
	}
}

func (c *Client) ProxyRequest(path string) (*RequestMetrics, error) {
	metrics := &RequestMetrics{
		Method: "GET",
		Path:   path,
	}

	url := c.baseURL + path
	startTime := time.Now()

	resp, err := c.httpClient.Get(url)
	elapsed := time.Since(startTime)
	metrics.ResponseTimeMs = elapsed.Milliseconds()

	if err != nil {
		metrics.Error = err
		metrics.ResponseStatus = 0
		return metrics, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	metrics.ResponseStatus = resp.StatusCode

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		metrics.Error = err
		return metrics, fmt.Errorf("failed to read response body: %w", err)
	}

	metrics.ResponseBody = body

	// Validate JSON response
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var jsonCheck any
		if err := json.Unmarshal(body, &jsonCheck); err != nil {
			metrics.Error = err
			return metrics, fmt.Errorf("invalid JSON response: %w", err)
		}
	}

	return metrics, nil
}
