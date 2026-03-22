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
	"github.com/mjrgr/enedis-linky-mcp-server/internal/dashboard"
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
	mcpServer := mcp.New(svc, logger, cfg.PRM)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.Dashboard {
		dash := dashboard.New(svc, logger, cfg.PRM)
		go func() {
			if err := dash.Serve(ctx, cfg.DashAddr); err != nil {
				logger.Error("dashboard error", "err", err)
			}
		}()
		logger.Info("dashboard enabled", "addr", cfg.DashAddr)

		// When the dashboard is enabled, run the MCP transport in the background
		// so the process stays alive for HTTP requests even if stdio reaches EOF.
		go func() {
			switch cfg.Transport {
			case "sse":
				addr := fmt.Sprintf(":%d", cfg.Port)
				if err := mcpServer.ServeSSE(ctx, addr); err != nil {
					logger.Error("mcp sse error", "err", err)
				}
			default:
				if err := mcpServer.ServeStdio(ctx); err != nil {
					logger.Error("mcp stdio error", "err", err)
				}
			}
		}()

		<-ctx.Done()
		return nil
	}

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
