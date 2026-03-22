// Package client provides a typed HTTP client for the Conso API with retry and backoff.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

const (
	// userAgent identifies this client to the Conso API (required by API author).
	userAgent = "enedis-linky-mcp-server/1.0 (+https://github.com/mjrgr/enedis-linky-mcp-server)"

	maxRetries          = 3
	initialRetryBackoff = 500 * time.Millisecond
	httpTimeout         = 30 * time.Second
)

// Client is a typed HTTP client for the Conso API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	logger     *slog.Logger
}

// ConsumptionResponse is the standard response envelope from the Conso API.
type ConsumptionResponse struct {
	ReadingType     ReadingType       `json:"reading_type"`
	IntervalReading []IntervalReading `json:"interval_reading"`
}

// ReadingType describes the unit and measurement kind of the returned data.
type ReadingType struct {
	Unit            string `json:"unit"`
	Aggregate       string `json:"aggregate"`
	MeasurementKind string `json:"measurement_kind"`
	MeasuringPeriod string `json:"measuring_period"`
}

// IntervalReading is a single data point with a date and a value.
type IntervalReading struct {
	Date  string `json:"date"`
	Value string `json:"value"`
}

// APIError represents a non-2xx response from the Conso API.
type APIError struct {
	StatusCode int
	RawBody    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("conso API responded with status %d: %s", e.StatusCode, e.RawBody)
}

// IsClientError returns true for 4xx errors that should not be retried.
func (e *APIError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// New creates a new Conso API client.
func New(baseURL, token string, logger *slog.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
		logger: logger,
	}
}

// QueryParams holds the common query parameters for Conso API endpoints.
type QueryParams struct {
	PRM   string
	Start string
	End   string
}

// GetData fetches data from the given endpoint with retry and backoff.
func (c *Client) GetData(ctx context.Context, endpoint string, params QueryParams) (*ConsumptionResponse, error) {
	rawURL := c.buildURL(endpoint, params)

	var lastErr error
	backoff := initialRetryBackoff

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Debug("retrying request",
				"endpoint", endpoint,
				"attempt", attempt,
				"backoff_ms", backoff.Milliseconds(),
			)
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff):
			}
			backoff *= 2
		}

		result, err := c.doRequest(ctx, rawURL)
		if err == nil {
			return result, nil
		}

		// Do not retry on client errors (4xx).
		if apiErr, ok := err.(*APIError); ok && apiErr.IsClientError() {
			return nil, err
		}

		lastErr = err
		c.logger.Warn("request failed, will retry",
			"endpoint", endpoint,
			"attempt", attempt,
			"error", err,
		)
	}

	return nil, fmt.Errorf("all %d attempts failed for %s: %w", maxRetries+1, endpoint, lastErr)
}

// Ping checks that the Conso API base URL is reachable.
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL, nil)
	if err != nil {
		return fmt.Errorf("building ping request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("API unreachable: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

// buildURL constructs the full URL with query parameters.
func (c *Client) buildURL(endpoint string, params QueryParams) string {
	q := url.Values{}
	if params.PRM != "" {
		q.Set("prm", params.PRM)
	}
	if params.Start != "" {
		q.Set("start", params.Start)
	}
	if params.End != "" {
		q.Set("end", params.End)
	}

	u := fmt.Sprintf("%s/%s", c.baseURL, endpoint)
	if len(q) > 0 {
		u = u + "?" + q.Encode()
	}
	return u
}

// doRequest performs a single HTTP GET request and decodes the response.
func (c *Client) doRequest(ctx context.Context, rawURL string) (*ConsumptionResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	c.logger.Debug("sending request", "url", rawURL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MB cap
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			RawBody:    string(body),
		}
	}

	var result ConsumptionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decoding response JSON: %w", err)
	}

	c.logger.Debug("received response",
		"endpoint", rawURL,
		"readings", len(result.IntervalReading),
	)

	return &result, nil
}
