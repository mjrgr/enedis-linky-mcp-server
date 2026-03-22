package mcp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/server"

	"github.com/mjrgr/enedis-linky-mcp-server/internal/service"
)

const (
	serverName    = "enedis-linky-mcp-server"
	serverVersion = "1.0.0"
)

// Server wraps the MCP server and its registered tools.
type Server struct {
	mcp     *server.MCPServer
	service *service.ConsumptionService
	logger  *slog.Logger
	prm     string
}

// New creates a new MCP server with all tools registered.
func New(svc *service.ConsumptionService, logger *slog.Logger, prm string) *Server {
	s := &Server{
		mcp:     server.NewMCPServer(serverName, serverVersion),
		service: svc,
		logger:  logger,
		prm:     prm,
	}
	s.registerTools()
	return s
}

// ServeStdio runs the MCP server over stdio (for use with Claude Desktop, etc.).
func (s *Server) ServeStdio(_ context.Context) error {
	s.logger.Info("starting MCP server", "transport", "stdio")
	return server.ServeStdio(s.mcp)
}

// ServeSSE runs the MCP server over HTTP/SSE on the given address.
func (s *Server) ServeSSE(ctx context.Context, addr string) error {
	s.logger.Info("starting MCP server", "transport", "sse", "addr", addr)

	sseServer := server.NewSSEServer(s.mcp,
		server.WithBaseURL(fmt.Sprintf("http://%s", addr)),
	)

	errCh := make(chan error, 1)
	go func() {
		errCh <- sseServer.Start(addr)
	}()

	select {
	case <-ctx.Done():
		return sseServer.Shutdown(ctx)
	case err := <-errCh:
		return err
	}
}
