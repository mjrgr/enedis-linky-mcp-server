package dashboard

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/mjrgr/enedis-linky-mcp-server/internal/service"
)

//go:embed dashboard.html
var page []byte

// Server serves the web dashboard and proxies data requests to the service layer.
type Server struct {
	svc    *service.ConsumptionService
	logger *slog.Logger
	mux    *http.ServeMux
	prm    string
}

// New creates a dashboard Server and registers all routes.
func New(svc *service.ConsumptionService, logger *slog.Logger, prm string) *Server {
	s := &Server{svc: svc, logger: logger, mux: http.NewServeMux(), prm: prm}
	s.mux.HandleFunc("GET /", s.handlePage)
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/config", s.handleConfig)
	s.mux.HandleFunc("GET /api/consumption", s.handleConsumption)
	s.mux.HandleFunc("GET /api/load-curve", s.handleLoadCurve)
	s.mux.HandleFunc("GET /api/max-power", s.handleMaxPower)
	s.mux.HandleFunc("GET /api/summary", s.handleSummary)
	return s
}

// Serve starts the HTTP server and blocks until ctx is cancelled.
func (s *Server) Serve(ctx context.Context, addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("dashboard listen %s: %w", addr, err)
	}
	s.logger.Info("dashboard listening", "addr", ln.Addr().String())

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) handleConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"prm": s.prm})
}

func (s *Server) handlePage(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(page)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.svc.HealthCheck(r.Context()))
}

func (s *Server) handleConsumption(w http.ResponseWriter, r *http.Request) {
	prm, start, end, ok := queryParams(w, r)
	if !ok {
		return
	}
	result, err := s.svc.GetDailyConsumption(r.Context(), prm, start, end)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleLoadCurve(w http.ResponseWriter, r *http.Request) {
	prm, start, end, ok := queryParams(w, r)
	if !ok {
		return
	}
	result, err := s.svc.GetLoadCurve(r.Context(), prm, start, end)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleMaxPower(w http.ResponseWriter, r *http.Request) {
	prm, start, end, ok := queryParams(w, r)
	if !ok {
		return
	}
	result, err := s.svc.GetMaxPower(r.Context(), prm, start, end)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	prm, start, end, ok := queryParams(w, r)
	if !ok {
		return
	}
	result, err := s.svc.GetConsumptionSummary(r.Context(), prm, start, end)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func queryParams(w http.ResponseWriter, r *http.Request) (prm, start, end string, ok bool) {
	prm = r.URL.Query().Get("prm")
	start = r.URL.Query().Get("start")
	end = r.URL.Query().Get("end")
	if prm == "" || start == "" || end == "" {
		writeError(w, fmt.Errorf("prm, start, and end query parameters are required"))
		return "", "", "", false
	}
	return prm, start, end, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadGateway)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
