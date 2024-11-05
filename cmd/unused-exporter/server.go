package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func runWebServer(ctx context.Context, cfg config) error {
	mux := http.NewServeMux()
	promHandler := promhttp.Handler()
	mux.HandleFunc(cfg.Web.Path, func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		promHandler.ServeHTTP(w, req)
		cfg.Logger.Info("Prometheus query",
			slog.String("path", cfg.Web.Path),
			slog.Duration("dur", time.Since(start)),
		)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			fmt.Fprintf(w, indexTemplate, cfg.Web.Path) // nolint:errcheck
		} else {
			http.NotFound(w, req)
		}
	})

	srv := &http.Server{
		ReadTimeout: 1 * time.Second,
		Addr:        cfg.Web.Address,
		Handler:     mux,
	}

	listenErr := make(chan error)

	go func() {
		cfg.Logger.Info("starting server",
			slog.String("addr", cfg.Web.Address),
			slog.String("metricspath", cfg.Web.Path),
		)
		listenErr <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		cfg.Logger.Info("shutting down server")
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.Timeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutting down server: %w", err)
		}

	case err := <-listenErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("running server: %w", err)
		}
	}

	return nil
}

const indexTemplate = `<!doctype html>
<html>
  <head><title>Unused Disks Exporter</title></head>
  <body>
    <h1>Unused Disks Exporter</h1>
    <p><a href=%q>Metrics</a></p>
  </body>
</html>`
