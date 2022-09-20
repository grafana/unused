package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/inkel/logfmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func runWebServer(ctx context.Context, logger *logfmt.Logger, addr, metricsPath string) error {
	mux := http.NewServeMux()
	h := promhttp.Handler()
	mux.HandleFunc(metricsPath, func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, req)
		logger.Log("Prometheus query", logfmt.Labels{
			"path": metricsPath,
			"dur":  time.Since(start),
		})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			fmt.Fprintf(w, indexTemplate, metricsPath)
		} else {
			http.NotFound(w, req)
		}
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	var closeErr error

	go func() {
		<-ctx.Done()

		timeout := 5 * time.Second // TODO move this to a configuration value
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		logger.Log("shutting down server", nil)
		closeErr = srv.Shutdown(ctx)
	}()

	logger.Log("starting server", logfmt.Labels{"addr": addr, "metricspath": metricsPath})
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("running server: %w", err)
	}

	if closeErr != nil {
		return fmt.Errorf("shutting down server: %w", closeErr)
	}

	return nil
}

const indexTemplate = `<!doctype html>
<html>
  <head><title>Unused Disks Exporter</title></head>
  <body>
    <h1>Unused Disks Exporter</h1>
    <p><a href=%q'>Metrics</a></p>
  </body>
</html>`
