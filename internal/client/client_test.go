package client

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newTestClient(baseURL string) *Client {
	return New(baseURL, "test-token", slog.New(slog.NewTextHandler(os.Stderr, nil)))
}

func TestGetData_Success(t *testing.T) {
	want := ConsumptionResponse{
		ReadingType: ReadingType{
			Unit:            "Wh",
			Aggregate:       "sum",
			MeasurementKind: "energy",
			MeasuringPeriod: "P1D",
		},
		IntervalReading: []IntervalReading{
			{Date: "2025-01-01", Value: "12340"},
			{Date: "2025-01-02", Value: "9870"},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("User-Agent") != userAgent {
			http.Error(w, "bad user-agent", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	got, err := c.GetData(context.Background(), "daily_consumption", QueryParams{
		PRM:   "12345678901234",
		Start: "2025-01-01",
		End:   "2025-01-02",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.IntervalReading) != 2 {
		t.Errorf("got %d readings, want 2", len(got.IntervalReading))
	}
	if got.ReadingType.Unit != "Wh" {
		t.Errorf("unit = %q, want Wh", got.ReadingType.Unit)
	}
}

func TestGetData_ClientError_NoRetry(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		http.Error(w, `{"error":"bad_request"}`, http.StatusBadRequest)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	_, err := c.GetData(context.Background(), "daily_consumption", QueryParams{})
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
	// Should NOT retry on 4xx.
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1 (no retries for client errors)", callCount)
	}
}

func TestGetData_ServerError_Retries(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		// Return 500 on first 2 calls, success on 3rd.
		if callCount < 3 {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		resp := ConsumptionResponse{
			ReadingType:     ReadingType{Unit: "Wh"},
			IntervalReading: []IntervalReading{{Date: "2025-01-01", Value: "100"}},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	// Use a very short backoff by creating a custom client.
	c.httpClient.Timeout = 5 * httpTimeout / 5

	got, err := c.GetData(context.Background(), "daily_consumption", QueryParams{})
	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if len(got.IntervalReading) != 1 {
		t.Errorf("got %d readings, want 1", len(got.IntervalReading))
	}
	if callCount != 3 {
		t.Errorf("callCount = %d, want 3", callCount)
	}
}

func TestBuildURL(t *testing.T) {
	c := newTestClient("https://conso.boris.sh/api")

	tests := []struct {
		endpoint string
		params   QueryParams
		want     string
	}{
		{
			endpoint: "daily_consumption",
			params:   QueryParams{PRM: "12345", Start: "2025-01-01", End: "2025-01-31"},
			want:     "https://conso.boris.sh/api/daily_consumption?end=2025-01-31&prm=12345&start=2025-01-01",
		},
		{
			endpoint: "health",
			params:   QueryParams{},
			want:     "https://conso.boris.sh/api/health",
		},
	}

	for _, tt := range tests {
		got := c.buildURL(tt.endpoint, tt.params)
		if got != tt.want {
			t.Errorf("buildURL(%q) = %q, want %q", tt.endpoint, got, tt.want)
		}
	}
}

func TestAPIError(t *testing.T) {
	e := &APIError{StatusCode: 401, RawBody: "unauthorized"}
	if !e.IsClientError() {
		t.Error("expected 401 to be a client error")
	}

	e500 := &APIError{StatusCode: 500, RawBody: "server error"}
	if e500.IsClientError() {
		t.Error("expected 500 to not be a client error")
	}
}
