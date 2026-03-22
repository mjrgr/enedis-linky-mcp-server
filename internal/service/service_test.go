package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mjrgr/enedis-linky-mcp-server/internal/client"
)

func newTestService(handler http.HandlerFunc) (*ConsumptionService, *httptest.Server) {
	srv := httptest.NewServer(handler)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	c := client.New(srv.URL, "test-token", logger)
	return New(c), srv
}

func dailyResponse() client.ConsumptionResponse {
	return client.ConsumptionResponse{
		ReadingType: client.ReadingType{
			Unit:            "Wh",
			Aggregate:       "sum",
			MeasurementKind: "energy",
			MeasuringPeriod: "P1D",
		},
		IntervalReading: []client.IntervalReading{
			{Date: "2025-01-01", Value: "12340"},
			{Date: "2025-01-02", Value: "9870"},
			{Date: "2025-01-03", Value: "15200"},
		},
	}
}

func TestGetDailyConsumption(t *testing.T) {
	svc, srv := newTestService(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/daily_consumption" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("prm") != "12345678901234" {
			t.Errorf("unexpected prm: %s", r.URL.Query().Get("prm"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dailyResponse())
	})
	defer srv.Close()

	result, err := svc.GetDailyConsumption(context.Background(), "12345678901234", "2025-01-01", "2025-01-03")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PRM != "12345678901234" {
		t.Errorf("PRM = %q, want 12345678901234", result.PRM)
	}
	if result.Count != 3 {
		t.Errorf("Count = %d, want 3", result.Count)
	}
	if result.Unit != "Wh" {
		t.Errorf("Unit = %q, want Wh", result.Unit)
	}
}

func TestGetConsumptionSummary(t *testing.T) {
	svc, srv := newTestService(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dailyResponse())
	})
	defer srv.Close()

	result, err := svc.GetConsumptionSummary(context.Background(), "12345678901234", "2025-01-01", "2025-01-03")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 12340 + 9870 + 15200 = 37410.
	if result.TotalWh != 37410 {
		t.Errorf("TotalWh = %v, want 37410", result.TotalWh)
	}
	if result.TotalKWh != 37.41 {
		t.Errorf("TotalKWh = %v, want 37.41", result.TotalKWh)
	}
	// avg = 37410 / 3 = 12470.
	if result.AveragePerDay != 12470 {
		t.Errorf("AveragePerDay = %v, want 12470", result.AveragePerDay)
	}
	if result.PeakDay != "2025-01-03" {
		t.Errorf("PeakDay = %q, want 2025-01-03", result.PeakDay)
	}
	if result.PeakValue != 15200 {
		t.Errorf("PeakValue = %v, want 15200", result.PeakValue)
	}
	if result.ReadingCount != 3 {
		t.Errorf("ReadingCount = %d, want 3", result.ReadingCount)
	}
}

func TestGetConsumptionSummaryEmpty(t *testing.T) {
	svc, srv := newTestService(func(w http.ResponseWriter, _ *http.Request) {
		resp := client.ConsumptionResponse{
			ReadingType:     client.ReadingType{Unit: "Wh"},
			IntervalReading: []client.IntervalReading{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	result, err := svc.GetConsumptionSummary(context.Background(), "12345678901234", "2025-01-01", "2025-01-03")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ReadingCount != 0 {
		t.Errorf("ReadingCount = %d, want 0", result.ReadingCount)
	}
	if result.TotalWh != 0 {
		t.Errorf("TotalWh = %v, want 0", result.TotalWh)
	}
}

func TestGetLoadCurve(t *testing.T) {
	resp := client.ConsumptionResponse{
		ReadingType: client.ReadingType{
			Unit:            "W",
			MeasuringPeriod: "PT30M",
		},
		IntervalReading: []client.IntervalReading{
			{Date: "2025-01-01 00:00", Value: "450"},
			{Date: "2025-01-01 00:30", Value: "420"},
		},
	}

	svc, srv := newTestService(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/load_curve" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	result, err := svc.GetLoadCurve(context.Background(), "12345678901234", "2025-01-01", "2025-01-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Unit != "W" {
		t.Errorf("Unit = %q, want W", result.Unit)
	}
	if result.Count != 2 {
		t.Errorf("Count = %d, want 2", result.Count)
	}
}

func TestGetMaxPower(t *testing.T) {
	resp := client.ConsumptionResponse{
		ReadingType: client.ReadingType{Unit: "VA"},
		IntervalReading: []client.IntervalReading{
			{Date: "2025-01-01", Value: "6000"},
		},
	}

	svc, srv := newTestService(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/max_power" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	result, err := svc.GetMaxPower(context.Background(), "12345678901234", "2025-01-01", "2025-01-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Unit != "VA" {
		t.Errorf("Unit = %q, want VA", result.Unit)
	}
}

func TestHealthCheck(t *testing.T) {
	svc, srv := newTestService(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	result := svc.HealthCheck(context.Background())
	if result.Status != "healthy" {
		t.Errorf("Status = %q, want healthy", result.Status)
	}
}

func TestHealthCheckUnhealthy(t *testing.T) {
	svc := New(client.New("http://127.0.0.1:1", "tok", slog.New(slog.NewTextHandler(os.Stderr, nil))))

	result := svc.HealthCheck(context.Background())
	if result.Status != "unhealthy" {
		t.Errorf("Status = %q, want unhealthy", result.Status)
	}
}

func TestValidateDateRange(t *testing.T) {
	tests := []struct {
		name    string
		start   string
		end     string
		wantErr bool
	}{
		{
			name:    "valid range",
			start:   "2025-01-01",
			end:     "2025-01-31",
			wantErr: false,
		},
		{
			name:    "same day",
			start:   "2025-06-15",
			end:     "2025-06-15",
			wantErr: false,
		},
		{
			name:    "invalid start format",
			start:   "01-01-2025",
			end:     "2025-01-31",
			wantErr: true,
		},
		{
			name:    "invalid end format",
			start:   "2025-01-01",
			end:     "not-a-date",
			wantErr: true,
		},
		{
			name:    "start after end",
			start:   "2025-02-01",
			end:     "2025-01-01",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDateRange(tt.start, tt.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDateRange(%q, %q) error = %v, wantErr %v",
					tt.start, tt.end, err, tt.wantErr)
			}
		})
	}
}

func TestRoundTo(t *testing.T) {
	cases := []struct {
		input    float64
		decimals int
		want     float64
	}{
		{1234.5678, 2, 1234.57},
		{1000.0, 3, 1000.0},
		{0.0, 2, 0.0},
		{-5.555, 2, -5.56},
	}

	for _, c := range cases {
		got := roundTo(c.input, c.decimals)
		if got != c.want {
			t.Errorf("roundTo(%v, %d) = %v, want %v", c.input, c.decimals, got, c.want)
		}
	}
}

func TestInvalidDateRangeReturnsError(t *testing.T) {
	svc, srv := newTestService(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	_, err := svc.GetDailyConsumption(context.Background(), "123", "2025-02-01", "2025-01-01")
	if err == nil {
		t.Error("expected error for start > end")
	}
}
