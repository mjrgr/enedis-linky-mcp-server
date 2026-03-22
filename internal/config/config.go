package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	// Token is the Conso API bearer token (required).
	Token string
	// Port is the HTTP port for SSE transport (default: 8080).
	Port int
	// LogLevel controls log verbosity.
	LogLevel slog.Level
	// BaseURL is the Conso API base URL (override for testing).
	BaseURL string
	// Transport selects the MCP transport: "stdio" or "sse" (default: stdio).
	Transport string
	// PRM is the default PRM meter identifier (optional, overridable per-call).
	PRM string
	// Dashboard enables the web dashboard (default: false).
	Dashboard bool
	// DashAddr is the listen address for the web dashboard (default: :8081).
	DashAddr string
}

// Load reads configuration from environment variables and validates required fields.
func Load() (*Config, error) {
	token := os.Getenv("CONSO_API_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("CONSO_API_TOKEN environment variable is required")
	}

	port := 8080
	if portStr := os.Getenv("PORT"); portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT value %q: %w", portStr, err)
		}
		if p < 1 || p > 65535 {
			return nil, fmt.Errorf("PORT must be between 1 and 65535, got %d", p)
		}
		port = p
	}

	logLevel := slog.LevelInfo
	if ll := strings.ToLower(os.Getenv("LOG_LEVEL")); ll != "" {
		switch ll {
		case "debug":
			logLevel = slog.LevelDebug
		case "info":
			logLevel = slog.LevelInfo
		case "warn", "warning":
			logLevel = slog.LevelWarn
		case "error":
			logLevel = slog.LevelError
		default:
			return nil, fmt.Errorf("invalid LOG_LEVEL %q: must be debug, info, warn, or error", ll)
		}
	}

	baseURL := "https://conso.boris.sh/api"
	if u := os.Getenv("CONSO_API_BASE_URL"); u != "" {
		baseURL = strings.TrimRight(u, "/")
	}

	transport := "stdio"
	if t := strings.ToLower(os.Getenv("MCP_TRANSPORT")); t != "" {
		switch t {
		case "stdio", "sse":
			transport = t
		default:
			return nil, fmt.Errorf("invalid MCP_TRANSPORT %q: must be stdio or sse", t)
		}
	}

	prm := os.Getenv("LINKY_PRM")

	dashboard := strings.EqualFold(os.Getenv("LINKY_DASHBOARD"), "true")
	dashAddr := ":8081"
	if a := os.Getenv("LINKY_DASH_ADDR"); a != "" {
		dashAddr = a
	}

	return &Config{
		Token:     token,
		Port:      port,
		LogLevel:  logLevel,
		BaseURL:   baseURL,
		Transport: transport,
		PRM:       prm,
		Dashboard: dashboard,
		DashAddr:  dashAddr,
	}, nil
}
