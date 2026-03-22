package config

import (
	"log/slog"
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("missing token returns error", func(t *testing.T) {
		os.Unsetenv("CONSO_API_TOKEN")
		_, err := Load()
		if err == nil {
			t.Fatal("expected error when CONSO_API_TOKEN is missing")
		}
	})

	t.Run("valid minimal config", func(t *testing.T) {
		os.Setenv("CONSO_API_TOKEN", "test-token")
		defer os.Unsetenv("CONSO_API_TOKEN")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Token != "test-token" {
			t.Errorf("Token = %q, want %q", cfg.Token, "test-token")
		}
		if cfg.Port != 8080 {
			t.Errorf("Port = %d, want 8080", cfg.Port)
		}
		if cfg.LogLevel != slog.LevelInfo {
			t.Errorf("LogLevel = %v, want Info", cfg.LogLevel)
		}
		if cfg.Transport != "stdio" {
			t.Errorf("Transport = %q, want stdio", cfg.Transport)
		}
	})

	t.Run("custom port", func(t *testing.T) {
		os.Setenv("CONSO_API_TOKEN", "tok")
		os.Setenv("PORT", "9090")
		defer os.Unsetenv("CONSO_API_TOKEN")
		defer os.Unsetenv("PORT")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Port != 9090 {
			t.Errorf("Port = %d, want 9090", cfg.Port)
		}
	})

	t.Run("invalid port returns error", func(t *testing.T) {
		os.Setenv("CONSO_API_TOKEN", "tok")
		os.Setenv("PORT", "not-a-number")
		defer os.Unsetenv("CONSO_API_TOKEN")
		defer os.Unsetenv("PORT")

		_, err := Load()
		if err == nil {
			t.Fatal("expected error for invalid PORT")
		}
	})

	t.Run("debug log level", func(t *testing.T) {
		os.Setenv("CONSO_API_TOKEN", "tok")
		os.Setenv("LOG_LEVEL", "debug")
		defer os.Unsetenv("CONSO_API_TOKEN")
		defer os.Unsetenv("LOG_LEVEL")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.LogLevel != slog.LevelDebug {
			t.Errorf("LogLevel = %v, want Debug", cfg.LogLevel)
		}
	})

	t.Run("sse transport", func(t *testing.T) {
		os.Setenv("CONSO_API_TOKEN", "tok")
		os.Setenv("MCP_TRANSPORT", "sse")
		defer os.Unsetenv("CONSO_API_TOKEN")
		defer os.Unsetenv("MCP_TRANSPORT")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Transport != "sse" {
			t.Errorf("Transport = %q, want sse", cfg.Transport)
		}
	})
}
