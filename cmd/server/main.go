package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/mjrgr/enedis-linky-mcp-server/internal/client"
	"github.com/mjrgr/enedis-linky-mcp-server/internal/config"
	"github.com/mjrgr/enedis-linky-mcp-server/internal/mcp"
	"github.com/mjrgr/enedis-linky-mcp-server/internal/service"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	logger := newLogger(cfg.LogLevel)
	logger.Info("enedis-linky-mcp-server starting",
		"transport", cfg.Transport,
		"base_url", cfg.BaseURL,
	)

	apiClient := client.New(cfg.BaseURL, cfg.Token, logger)
	svc := service.New(apiClient)
	mcpServer := mcp.New(svc, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	switch cfg.Transport {
	case "sse":
		addr := fmt.Sprintf(":%d", cfg.Port)
		return mcpServer.ServeSSE(ctx, addr)
	default: // stdio
		return mcpServer.ServeStdio(ctx)
	}
}

// newLogger creates a structured logger at the requested level.
func newLogger(level slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{Level: level}
	// When using stdio transport the MCP protocol owns stdout, so log to stderr.
	handler := slog.NewJSONHandler(os.Stderr, opts)
	return slog.New(handler)
}
