package mcp

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/mjrgr/enedis-linky-mcp-server/internal/client"
	"github.com/mjrgr/enedis-linky-mcp-server/internal/service"
)

func setupTestServer(handler http.HandlerFunc) (*Server, *httptest.Server) {
	srv := httptest.NewServer(handler)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	c := client.New(srv.URL, "test-token", logger)
	svc := service.New(c)
	mcpSrv := New(svc, logger)
	return mcpSrv, srv
}

// callTool sends a tools/call JSON-RPC request through HandleMessage and returns the parsed result.
func callTool(t *testing.T, s *Server, name string, args map[string]any) *mcpgo.CallToolResult {
	t.Helper()

	// First initialize the server.
	initMsg, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-03-26",
			"clientInfo": map[string]any{
				"name":    "test-client",
				"version": "1.0.0",
			},
			"capabilities": map[string]any{},
		},
	})
	s.mcp.HandleMessage(context.Background(), initMsg)

	// Build tools/call request.
	reqMsg, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      name,
			"arguments": args,
		},
	})
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	resp := s.mcp.HandleMessage(context.Background(), reqMsg)

	// Parse the JSON-RPC response.
	respBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var rpcResp struct {
		Result mcpgo.CallToolResult `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBytes, &rpcResp); err != nil {
		t.Fatalf("failed to parse response: %v\nraw: %s", err, string(respBytes))
	}
	if rpcResp.Error != nil {
		t.Fatalf("JSON-RPC error: %d %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return &rpcResp.Result
}

func getTextContent(t *testing.T, result *mcpgo.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty content")
	}
	// Content items are interface values; marshal and re-parse to extract text.
	raw, err := json.Marshal(result.Content[0])
	if err != nil {
		t.Fatalf("failed to marshal content: %v", err)
	}
	var tc struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &tc); err != nil {
		t.Fatalf("failed to parse content: %v", err)
	}
	return tc.Text
}

func consoResponse() client.ConsumptionResponse {
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
	mcpSrv, srv := setupTestServer(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(consoResponse())
	})
	defer srv.Close()

	result := callTool(t, mcpSrv, "get_daily_consumption", map[string]any{
		"prm":   "12345678901234",
		"start": "2025-01-01",
		"end":   "2025-01-03",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %s", getTextContent(t, result))
	}

	text := getTextContent(t, result)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if parsed["prm"] != "12345678901234" {
		t.Errorf("prm = %v, want 12345678901234", parsed["prm"])
	}
	if parsed["unit"] != "Wh" {
		t.Errorf("unit = %v, want Wh", parsed["unit"])
	}
	count, ok := parsed["count"].(float64)
	if !ok || count != 3 {
		t.Errorf("count = %v, want 3", parsed["count"])
	}
}

func TestGetConsumptionSummary(t *testing.T) {
	mcpSrv, srv := setupTestServer(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(consoResponse())
	})
	defer srv.Close()

	result := callTool(t, mcpSrv, "get_consumption_summary", map[string]any{
		"prm":   "12345678901234",
		"start": "2025-01-01",
		"end":   "2025-01-03",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %s", getTextContent(t, result))
	}

	text := getTextContent(t, result)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	// total = 12340 + 9870 + 15200 = 37410.
	if parsed["total_wh"] != 37410.0 {
		t.Errorf("total_wh = %v, want 37410", parsed["total_wh"])
	}
	if parsed["total_kwh"] != 37.41 {
		t.Errorf("total_kwh = %v, want 37.41", parsed["total_kwh"])
	}
	if parsed["peak_day"] != "2025-01-03" {
		t.Errorf("peak_day = %v, want 2025-01-03", parsed["peak_day"])
	}
}

func TestHealthCheck(t *testing.T) {
	mcpSrv, srv := setupTestServer(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	result := callTool(t, mcpSrv, "health_check", nil)
	if result.IsError {
		t.Fatalf("tool returned error: %s", getTextContent(t, result))
	}

	text := getTextContent(t, result)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if parsed["status"] != "healthy" {
		t.Errorf("status = %v, want healthy", parsed["status"])
	}
}

func TestMissingParams(t *testing.T) {
	mcpSrv, srv := setupTestServer(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	tests := []struct {
		name string
		args map[string]any
	}{
		{"missing prm", map[string]any{"start": "2025-01-01", "end": "2025-01-31"}},
		{"missing start", map[string]any{"prm": "123", "end": "2025-01-31"}},
		{"missing end", map[string]any{"prm": "123", "start": "2025-01-01"}},
		{"empty args", map[string]any{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := callTool(t, mcpSrv, "get_daily_consumption", tt.args)
			if !result.IsError {
				t.Error("expected tool error for missing params")
			}
		})
	}
}

func TestInvalidDateRange(t *testing.T) {
	mcpSrv, srv := setupTestServer(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	result := callTool(t, mcpSrv, "get_daily_consumption", map[string]any{
		"prm":   "12345678901234",
		"start": "2025-02-01",
		"end":   "2025-01-01",
	})
	if !result.IsError {
		t.Error("expected tool error for start > end")
	}
}

func TestAPIError(t *testing.T) {
	mcpSrv, srv := setupTestServer(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
	})
	defer srv.Close()

	result := callTool(t, mcpSrv, "get_daily_consumption", map[string]any{
		"prm":   "12345678901234",
		"start": "2025-01-01",
		"end":   "2025-01-31",
	})
	if !result.IsError {
		t.Error("expected tool error for 401 API response")
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

	mcpSrv, srv := setupTestServer(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	result := callTool(t, mcpSrv, "get_load_curve", map[string]any{
		"prm":   "12345678901234",
		"start": "2025-01-01",
		"end":   "2025-01-01",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %s", getTextContent(t, result))
	}

	text := getTextContent(t, result)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if parsed["unit"] != "W" {
		t.Errorf("unit = %v, want W", parsed["unit"])
	}
}

func TestGetMaxPower(t *testing.T) {
	resp := client.ConsumptionResponse{
		ReadingType: client.ReadingType{Unit: "VA"},
		IntervalReading: []client.IntervalReading{
			{Date: "2025-01-01", Value: "6000"},
		},
	}

	mcpSrv, srv := setupTestServer(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	result := callTool(t, mcpSrv, "get_max_power", map[string]any{
		"prm":   "12345678901234",
		"start": "2025-01-01",
		"end":   "2025-01-01",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %s", getTextContent(t, result))
	}

	text := getTextContent(t, result)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if parsed["unit"] != "VA" {
		t.Errorf("unit = %v, want VA", parsed["unit"])
	}
}
