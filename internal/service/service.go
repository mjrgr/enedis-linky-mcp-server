package service

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/mjrgr/enedis-linky-mcp-server/internal/client"
)

const dateLayout = "2006-01-02"

// ConsumptionService orchestrates calls to the Conso API client and applies
// business logic such as aggregation and validation.
type ConsumptionService struct {
	client *client.Client
}

// New creates a new ConsumptionService backed by the given client.
func New(c *client.Client) *ConsumptionService {
	return &ConsumptionService{client: c}
}

// DailyConsumptionResult holds the raw response from the daily_consumption endpoint.
type DailyConsumptionResult struct {
	PRM             string                   `json:"prm"`
	StartDate       string                   `json:"start_date"`
	EndDate         string                   `json:"end_date"`
	Unit            string                   `json:"unit"`
	MeasuringPeriod string                   `json:"measuring_period"`
	Readings        []client.IntervalReading  `json:"readings"`
	Count           int                      `json:"count"`
}

// LoadCurveResult holds the raw response from the load_curve endpoint.
type LoadCurveResult struct {
	PRM             string                   `json:"prm"`
	StartDate       string                   `json:"start_date"`
	EndDate         string                   `json:"end_date"`
	Unit            string                   `json:"unit"`
	MeasuringPeriod string                   `json:"measuring_period"`
	Readings        []client.IntervalReading  `json:"readings"`
	Count           int                      `json:"count"`
}

// MaxPowerResult holds the raw response from the max_power endpoint.
type MaxPowerResult struct {
	PRM             string                   `json:"prm"`
	StartDate       string                   `json:"start_date"`
	EndDate         string                   `json:"end_date"`
	Unit            string                   `json:"unit"`
	MeasuringPeriod string                   `json:"measuring_period"`
	Readings        []client.IntervalReading  `json:"readings"`
	Count           int                      `json:"count"`
}

// DailyProductionResult holds the raw response from the daily_production endpoint.
type DailyProductionResult struct {
	PRM             string                   `json:"prm"`
	StartDate       string                   `json:"start_date"`
	EndDate         string                   `json:"end_date"`
	Unit            string                   `json:"unit"`
	MeasuringPeriod string                   `json:"measuring_period"`
	Readings        []client.IntervalReading  `json:"readings"`
	Count           int                      `json:"count"`
}

// ProductionLoadCurveResult holds the raw response from the production_load_curve endpoint.
type ProductionLoadCurveResult struct {
	PRM             string                   `json:"prm"`
	StartDate       string                   `json:"start_date"`
	EndDate         string                   `json:"end_date"`
	Unit            string                   `json:"unit"`
	MeasuringPeriod string                   `json:"measuring_period"`
	Readings        []client.IntervalReading  `json:"readings"`
	Count           int                      `json:"count"`
}

// SummaryResult contains aggregated statistics over a date range.
type SummaryResult struct {
	PRM          string  `json:"prm"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
	Unit         string  `json:"unit"`
	TotalWh      float64 `json:"total_wh"`
	TotalKWh     float64 `json:"total_kwh"`
	AveragePerDay float64 `json:"average_per_day_wh"`
	PeakDay      string  `json:"peak_day"`
	PeakValue    float64 `json:"peak_value_wh"`
	ReadingCount int     `json:"reading_count"`
}

// HealthResult contains the API health check result.
type HealthResult struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

// GetDailyConsumption fetches and returns daily consumption readings.
func (s *ConsumptionService) GetDailyConsumption(ctx context.Context, prm, start, end string) (*DailyConsumptionResult, error) {
	if err := validateDateRange(start, end); err != nil {
		return nil, err
	}

	resp, err := s.client.GetData(ctx, "daily_consumption", client.QueryParams{
		PRM:   prm,
		Start: start,
		End:   end,
	})
	if err != nil {
		return nil, fmt.Errorf("fetching daily consumption: %w", err)
	}

	return &DailyConsumptionResult{
		PRM:             prm,
		StartDate:       start,
		EndDate:         end,
		Unit:            resp.ReadingType.Unit,
		MeasuringPeriod: resp.ReadingType.MeasuringPeriod,
		Readings:        resp.IntervalReading,
		Count:           len(resp.IntervalReading),
	}, nil
}

// GetLoadCurve fetches 30-minute average power readings.
func (s *ConsumptionService) GetLoadCurve(ctx context.Context, prm, start, end string) (*LoadCurveResult, error) {
	if err := validateDateRange(start, end); err != nil {
		return nil, err
	}

	resp, err := s.client.GetData(ctx, "load_curve", client.QueryParams{
		PRM:   prm,
		Start: start,
		End:   end,
	})
	if err != nil {
		return nil, fmt.Errorf("fetching load curve: %w", err)
	}

	return &LoadCurveResult{
		PRM:             prm,
		StartDate:       start,
		EndDate:         end,
		Unit:            resp.ReadingType.Unit,
		MeasuringPeriod: resp.ReadingType.MeasuringPeriod,
		Readings:        resp.IntervalReading,
		Count:           len(resp.IntervalReading),
	}, nil
}

// GetMaxPower fetches the maximum power reached each day.
func (s *ConsumptionService) GetMaxPower(ctx context.Context, prm, start, end string) (*MaxPowerResult, error) {
	if err := validateDateRange(start, end); err != nil {
		return nil, err
	}

	resp, err := s.client.GetData(ctx, "max_power", client.QueryParams{
		PRM:   prm,
		Start: start,
		End:   end,
	})
	if err != nil {
		return nil, fmt.Errorf("fetching max power: %w", err)
	}

	return &MaxPowerResult{
		PRM:             prm,
		StartDate:       start,
		EndDate:         end,
		Unit:            resp.ReadingType.Unit,
		MeasuringPeriod: resp.ReadingType.MeasuringPeriod,
		Readings:        resp.IntervalReading,
		Count:           len(resp.IntervalReading),
	}, nil
}

// GetDailyProduction fetches daily solar/production readings.
func (s *ConsumptionService) GetDailyProduction(ctx context.Context, prm, start, end string) (*DailyProductionResult, error) {
	if err := validateDateRange(start, end); err != nil {
		return nil, err
	}

	resp, err := s.client.GetData(ctx, "daily_production", client.QueryParams{
		PRM:   prm,
		Start: start,
		End:   end,
	})
	if err != nil {
		return nil, fmt.Errorf("fetching daily production: %w", err)
	}

	return &DailyProductionResult{
		PRM:             prm,
		StartDate:       start,
		EndDate:         end,
		Unit:            resp.ReadingType.Unit,
		MeasuringPeriod: resp.ReadingType.MeasuringPeriod,
		Readings:        resp.IntervalReading,
		Count:           len(resp.IntervalReading),
	}, nil
}

// GetProductionLoadCurve fetches 30-minute average production power readings.
func (s *ConsumptionService) GetProductionLoadCurve(ctx context.Context, prm, start, end string) (*ProductionLoadCurveResult, error) {
	if err := validateDateRange(start, end); err != nil {
		return nil, err
	}

	resp, err := s.client.GetData(ctx, "production_load_curve", client.QueryParams{
		PRM:   prm,
		Start: start,
		End:   end,
	})
	if err != nil {
		return nil, fmt.Errorf("fetching production load curve: %w", err)
	}

	return &ProductionLoadCurveResult{
		PRM:             prm,
		StartDate:       start,
		EndDate:         end,
		Unit:            resp.ReadingType.Unit,
		MeasuringPeriod: resp.ReadingType.MeasuringPeriod,
		Readings:        resp.IntervalReading,
		Count:           len(resp.IntervalReading),
	}, nil
}

// GetConsumptionSummary aggregates daily consumption into statistics.
func (s *ConsumptionService) GetConsumptionSummary(ctx context.Context, prm, start, end string) (*SummaryResult, error) {
	if err := validateDateRange(start, end); err != nil {
		return nil, err
	}

	resp, err := s.client.GetData(ctx, "daily_consumption", client.QueryParams{
		PRM:   prm,
		Start: start,
		End:   end,
	})
	if err != nil {
		return nil, fmt.Errorf("fetching data for summary: %w", err)
	}

	if len(resp.IntervalReading) == 0 {
		return &SummaryResult{
			PRM:          prm,
			StartDate:    start,
			EndDate:      end,
			Unit:         resp.ReadingType.Unit,
			ReadingCount: 0,
		}, nil
	}

	var total float64
	peakDay := ""
	peakValue := math.Inf(-1)

	for _, r := range resp.IntervalReading {
		v, err := strconv.ParseFloat(r.Value, 64)
		if err != nil {
			continue
		}
		total += v
		if v > peakValue {
			peakValue = v
			peakDay = r.Date
		}
	}

	count := float64(len(resp.IntervalReading))
	avg := 0.0
	if count > 0 {
		avg = total / count
	}

	return &SummaryResult{
		PRM:           prm,
		StartDate:     start,
		EndDate:       end,
		Unit:          resp.ReadingType.Unit,
		TotalWh:       total,
		TotalKWh:      roundTo(total/1000, 3),
		AveragePerDay: roundTo(avg, 2),
		PeakDay:       peakDay,
		PeakValue:     peakValue,
		ReadingCount:  len(resp.IntervalReading),
	}, nil
}

// HealthCheck verifies connectivity to the Conso API.
func (s *ConsumptionService) HealthCheck(ctx context.Context) *HealthResult {
	err := s.client.Ping(ctx)
	if err != nil {
		return &HealthResult{
			Status:    "unhealthy",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Message:   fmt.Sprintf("Conso API unreachable: %s", err.Error()),
		}
	}
	return &HealthResult{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Message:   "Conso API is reachable",
	}
}

// validateDateRange checks that start and end dates are in YYYY-MM-DD format
// and that start is not after end.
func validateDateRange(start, end string) error {
	s, err := time.Parse(dateLayout, start)
	if err != nil {
		return fmt.Errorf("invalid start date %q: must be YYYY-MM-DD", start)
	}
	e, err := time.Parse(dateLayout, end)
	if err != nil {
		return fmt.Errorf("invalid end date %q: must be YYYY-MM-DD", end)
	}
	if s.After(e) {
		return fmt.Errorf("start date %s must not be after end date %s", start, end)
	}
	return nil
}

func roundTo(v float64, decimals int) float64 {
	p := math.Pow(10, float64(decimals))
	return math.Round(v*p) / p
}
