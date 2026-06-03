package main

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/grafana/unused/cmd/internal"
)

func TestRealMain(t *testing.T) {
	t.Run("fails with no providers", func(t *testing.T) {
		ctx := context.Background()
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		cfg := config{
			Logger: logger,
		}
		cfg.Collector.Timeout = time.Second
		cfg.Collector.PollInterval = time.Hour
		cfg.Web.Address = "localhost:0"
		cfg.Web.Path = "/metrics"
		cfg.Web.Timeout = time.Second

		err := realMain(ctx, cfg)
		if err != internal.ErrNoProviders {
			t.Errorf("expected ErrNoProviders, got %v", err)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		cfg := config{
			Logger: logger,
		}
		cfg.Collector.Timeout = time.Second
		cfg.Collector.PollInterval = time.Hour
		cfg.Web.Address = "localhost:0"
		cfg.Web.Path = "/metrics"
		cfg.Web.Timeout = time.Second

		// With cancelled context, CreateProviders should fail fast
		err := realMain(ctx, cfg)
		if err == nil {
			t.Error("expected error with cancelled context")
		}
	})
}
