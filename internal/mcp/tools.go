package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// registerTools adds all MCP tools to the server.
func (s *Server) registerTools() {
	s.registerGetDailyConsumption()
	s.registerGetLoadCurve()
	s.registerGetMaxPower()
	s.registerGetDailyProduction()
	s.registerGetProductionLoadCurve()
	s.registerGetConsumptionSummary()
	s.registerHealthCheck()
}

// ── get_daily_consumption ────────────────────────────────────────────────────

func (s *Server) registerGetDailyConsumption() {
	tool := mcp.NewTool("get_daily_consumption",
		mcp.WithDescription(
			"Fetch daily electricity consumption (in Wh) from a Linky smart meter "+
				"for a given date range. Data is updated once per day around 8 AM.",
		),
		mcp.WithString("prm",
			mcp.Required(),
			mcp.Description("PRM (Point de Référence de Mesure) — the meter's unique 14-digit identifier."),
		),
		mcp.WithString("start",
			mcp.Required(),
			mcp.Description("Start date in YYYY-MM-DD format (inclusive)."),
		),
		mcp.WithString("end",
			mcp.Required(),
			mcp.Description("End date in YYYY-MM-DD format (inclusive). Maximum range: 1 year."),
		),
	)

	s.mcp.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		prm, start, end, err := extractDateRangeParams(req)
		if err != nil {
			return toolError(err), nil
		}

		result, err := s.service.GetDailyConsumption(ctx, prm, start, end)
		if err != nil {
			s.logger.Error("get_daily_consumption failed", "error", err)
			return toolError(err), nil
		}

		return toolJSON(result)
	})
}

// ── get_load_curve ───────────────────────────────────────────────────────────

func (s *Server) registerGetLoadCurve() {
	tool := mcp.NewTool("get_load_curve",
		mcp.WithDescription(
			"Fetch the electricity load curve — average power consumed in Watts "+
				"over 30-minute intervals — for a given date range.",
		),
		mcp.WithString("prm",
			mcp.Required(),
			mcp.Description("PRM (Point de Référence de Mesure) — the meter's unique 14-digit identifier."),
		),
		mcp.WithString("start",
			mcp.Required(),
			mcp.Description("Start date in YYYY-MM-DD format (inclusive)."),
		),
		mcp.WithString("end",
			mcp.Required(),
			mcp.Description("End date in YYYY-MM-DD format (inclusive). Maximum range: 7 days recommended."),
		),
	)

	s.mcp.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		prm, start, end, err := extractDateRangeParams(req)
		if err != nil {
			return toolError(err), nil
		}

		result, err := s.service.GetLoadCurve(ctx, prm, start, end)
		if err != nil {
			s.logger.Error("get_load_curve failed", "error", err)
			return toolError(err), nil
		}

		return toolJSON(result)
	})
}

// ── get_max_power ────────────────────────────────────────────────────────────

func (s *Server) registerGetMaxPower() {
	tool := mcp.NewTool("get_max_power",
		mcp.WithDescription(
			"Fetch the maximum power (in VA) reached each day on a Linky smart meter "+
				"for a given date range.",
		),
		mcp.WithString("prm",
			mcp.Required(),
			mcp.Description("PRM (Point de Référence de Mesure) — the meter's unique 14-digit identifier."),
		),
		mcp.WithString("start",
			mcp.Required(),
			mcp.Description("Start date in YYYY-MM-DD format (inclusive)."),
		),
		mcp.WithString("end",
			mcp.Required(),
			mcp.Description("End date in YYYY-MM-DD format (inclusive)."),
		),
	)

	s.mcp.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		prm, start, end, err := extractDateRangeParams(req)
		if err != nil {
			return toolError(err), nil
		}

		result, err := s.service.GetMaxPower(ctx, prm, start, end)
		if err != nil {
			s.logger.Error("get_max_power failed", "error", err)
			return toolError(err), nil
		}

		return toolJSON(result)
	})
}

// ── get_daily_production ─────────────────────────────────────────────────────

func (s *Server) registerGetDailyProduction() {
	tool := mcp.NewTool("get_daily_production",
		mcp.WithDescription(
			"Fetch daily electricity production (in Wh) from a solar installation "+
				"connected to a Linky smart meter for a given date range.",
		),
		mcp.WithString("prm",
			mcp.Required(),
			mcp.Description("PRM (Point de Référence de Mesure) — the meter's unique 14-digit identifier."),
		),
		mcp.WithString("start",
			mcp.Required(),
			mcp.Description("Start date in YYYY-MM-DD format (inclusive)."),
		),
		mcp.WithString("end",
			mcp.Required(),
			mcp.Description("End date in YYYY-MM-DD format (inclusive)."),
		),
	)

	s.mcp.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		prm, start, end, err := extractDateRangeParams(req)
		if err != nil {
			return toolError(err), nil
		}

		result, err := s.service.GetDailyProduction(ctx, prm, start, end)
		if err != nil {
			s.logger.Error("get_daily_production failed", "error", err)
			return toolError(err), nil
		}

		return toolJSON(result)
	})
}

// ── get_production_load_curve ────────────────────────────────────────────────

func (s *Server) registerGetProductionLoadCurve() {
	tool := mcp.NewTool("get_production_load_curve",
		mcp.WithDescription(
			"Fetch the electricity production load curve — average power produced in Watts "+
				"over 30-minute intervals — for a solar installation connected to a Linky meter.",
		),
		mcp.WithString("prm",
			mcp.Required(),
			mcp.Description("PRM (Point de Référence de Mesure) — the meter's unique 14-digit identifier."),
		),
		mcp.WithString("start",
			mcp.Required(),
			mcp.Description("Start date in YYYY-MM-DD format (inclusive)."),
		),
		mcp.WithString("end",
			mcp.Required(),
			mcp.Description("End date in YYYY-MM-DD format (inclusive). Maximum range: 7 days recommended."),
		),
	)

	s.mcp.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		prm, start, end, err := extractDateRangeParams(req)
		if err != nil {
			return toolError(err), nil
		}

		result, err := s.service.GetProductionLoadCurve(ctx, prm, start, end)
		if err != nil {
			s.logger.Error("get_production_load_curve failed", "error", err)
			return toolError(err), nil
		}

		return toolJSON(result)
	})
}

// ── get_consumption_summary ──────────────────────────────────────────────────

func (s *Server) registerGetConsumptionSummary() {
	tool := mcp.NewTool("get_consumption_summary",
		mcp.WithDescription(
			"Compute aggregated statistics for daily electricity consumption over a date range: "+
				"total consumption in Wh and kWh, daily average, and peak consumption day.",
		),
		mcp.WithString("prm",
			mcp.Required(),
			mcp.Description("PRM (Point de Référence de Mesure) — the meter's unique 14-digit identifier."),
		),
		mcp.WithString("start",
			mcp.Required(),
			mcp.Description("Start date in YYYY-MM-DD format (inclusive)."),
		),
		mcp.WithString("end",
			mcp.Required(),
			mcp.Description("End date in YYYY-MM-DD format (inclusive)."),
		),
	)

	s.mcp.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		prm, start, end, err := extractDateRangeParams(req)
		if err != nil {
			return toolError(err), nil
		}

		result, err := s.service.GetConsumptionSummary(ctx, prm, start, end)
		if err != nil {
			s.logger.Error("get_consumption_summary failed", "error", err)
			return toolError(err), nil
		}

		return toolJSON(result)
	})
}

// ── health_check ─────────────────────────────────────────────────────────────

func (s *Server) registerHealthCheck() {
	tool := mcp.NewTool("health_check",
		mcp.WithDescription(
			"Check whether the Conso API (conso.boris.sh) is currently reachable. "+
				"Returns status, timestamp, and a human-readable message.",
		),
	)

	s.mcp.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result := s.service.HealthCheck(ctx)
		return toolJSON(result)
	})
}

// ── helpers ───────────────────────────────────────────────────────────────────

// extractDateRangeParams pulls prm, start, and end from tool arguments.
func extractDateRangeParams(req mcp.CallToolRequest) (prm, start, end string, err error) {
	prm = req.GetString("prm", "")
	if prm == "" {
		return "", "", "", fmt.Errorf("parameter 'prm' is required and must be a non-empty string")
	}

	start = req.GetString("start", "")
	if start == "" {
		return "", "", "", fmt.Errorf("parameter 'start' is required and must be a non-empty string")
	}

	end = req.GetString("end", "")
	if end == "" {
		return "", "", "", fmt.Errorf("parameter 'end' is required and must be a non-empty string")
	}

	return prm, start, end, nil
}

// toolJSON marshals v to pretty JSON and wraps it in a text tool result.
func toolJSON(v any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshalling result: %w", err)
	}
	return mcp.NewToolResultText(string(data)), nil
}

// toolError wraps an error as a tool result with isError=true.
func toolError(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError(err.Error())
}

