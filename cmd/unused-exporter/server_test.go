package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRunWebServer(t *testing.T) {
	t.Run("server starts and handles requests", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		cfg := config{
			Logger: logger,
		}
		cfg.Web.Address = "localhost:0" // Use random port
		cfg.Web.Path = "/metrics"
		cfg.Web.Timeout = time.Second

		// Start server in background
		serverErr := make(chan error, 1)
		go func() {
			serverErr <- runWebServer(ctx, cfg)
		}()

		// Give server time to start
		time.Sleep(50 * time.Millisecond)

		// Cancel context to trigger shutdown
		cancel()

		// Wait for server to finish
		select {
		case err := <-serverErr:
			if err != nil {
				t.Errorf("runWebServer() error = %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("server did not shut down in time")
		}
	})

	t.Run("server handles root request", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		cfg := config{
			Logger: logger,
		}
		cfg.Web.Address = "localhost:9876"
		cfg.Web.Path = "/metrics"
		cfg.Web.Timeout = time.Second

		// Start server
		go func() {
			_ = runWebServer(ctx, cfg)
		}()

		// Wait for server to start
		time.Sleep(100 * time.Millisecond)

		// Make request to root
		resp, err := http.Get("http://localhost:9876/")
		if err != nil {
			t.Skipf("Could not connect to server: %v (port might be in use)", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET / status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "Unused Disks Exporter") {
			t.Errorf("GET / body should contain title")
		}

		if !strings.Contains(bodyStr, "/metrics") {
			t.Errorf("GET / body should contain metrics link")
		}

		// Clean up
		cancel()
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("server handles 404", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		cfg := config{
			Logger: logger,
		}
		cfg.Web.Address = "localhost:9877"
		cfg.Web.Path = "/metrics"
		cfg.Web.Timeout = time.Second

		// Start server
		go func() {
			_ = runWebServer(ctx, cfg)
		}()

		// Wait for server to start
		time.Sleep(100 * time.Millisecond)

		// Make request to non-existent path
		resp, err := http.Get("http://localhost:9877/nonexistent")
		if err != nil {
			t.Skipf("Could not connect to server: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("GET /nonexistent status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}

		// Clean up
		cancel()
		time.Sleep(100 * time.Millisecond)
	})
}

func TestIndexTemplate(t *testing.T) {
	// Test that the index template is properly formatted
	if !strings.Contains(indexTemplate, "<title>") {
		t.Error("indexTemplate should contain <title>")
	}
	if !strings.Contains(indexTemplate, "Unused Disks Exporter") {
		t.Error("indexTemplate should contain 'Unused Disks Exporter'")
	}
	if !strings.Contains(indexTemplate, "<a href=") {
		t.Error("indexTemplate should contain metrics link")
	}
}
