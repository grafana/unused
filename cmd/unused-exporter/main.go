// unused-exporter is a Prometheus exporter with a web interface to
// expose unused disks as metrics.
//
// Provider selection is opinionated, currently accepting the
// following authentication method for each provider:
//   - GCP: pass gcp.project with a valid GCP project ID.
//   - AWS: pass aws.profile with a valid AWS shared profile.
//   - Azure: pass azure.sub with a valid Azure subscription ID.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/grafana/unused/cmd/internal"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	otelmetric "go.opentelemetry.io/otel/sdk/metric"
)

func main() {
	cfg := config{
		Logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	internal.ProviderFlags(flag.CommandLine, &cfg.Providers.GCP, &cfg.Providers.AWS, &cfg.Providers.Azure)

	flag.BoolVar(&cfg.VerboseLogging, "verbose", false, "add verbose logging information")
	flag.DurationVar(&cfg.Collector.Timeout, "collect.timeout", 30*time.Second, "timeout for collecting metrics from each provider")
	flag.StringVar(&cfg.Web.Path, "web.path", "/metrics", "path on which to expose metrics")
	flag.StringVar(&cfg.Web.Address, "web.address", ":8080", "address to expose metrics and web interface")
	flag.DurationVar(&cfg.Web.Timeout, "web.timeout", 5*time.Second, "timeout for shutting down the server")
	flag.DurationVar(&cfg.Collector.PollInterval, "collect.interval", 5*time.Minute, "interval to poll the cloud provider API for unused disks")

	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := realMain(ctx, cfg); err != nil {
		cancel() // cleanup resources
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func realMain(ctx context.Context, cfg config) error {
	providers, err := internal.CreateProviders(ctx, cfg.Logger, cfg.Providers.GCP, cfg.Providers.AWS, cfg.Providers.Azure)
	if err != nil {
		return err
	}

	if err := registerExporter(ctx, providers, cfg); err != nil {
		return fmt.Errorf("registering exporter: %w", err)
	}

	if err := setupOTLPExport(cfg.Logger); err != nil {
		return fmt.Errorf("setting up OTLP exporter: %w", err)
	}

	if err := runWebServer(ctx, cfg); err != nil {
		return fmt.Errorf("running web server: %w", err)
	}

	return nil
}

func setupOTLPExport(logger *slog.Logger) error {
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" && os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT") == "" {
		logger.Debug("OTLP exporter is not enabled")
		return nil
	}
	os.Setenv("OTEL_METRICS_PRODUCERS", "prometheus")

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(cause error) {
		logger.Error("OTLP exporter error", "err", cause)
	}))

	if os.Getenv("OTEL_SERVICE_NAME") == "" {
		os.Setenv("OTEL_SERVICE_NAME", "unused_exporter")
	}

	reader, err := autoexport.NewMetricReader(context.Background())
	if err != nil {
		return err
	}

	otelmetric.NewMeterProvider(otelmetric.WithReader(reader))

	return nil
}
