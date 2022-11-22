package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/grafana/unused/cmd/clicommon"
	"github.com/inkel/logfmt"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	// defer profile.Start(profile.CPUProfile, profile.MemProfile).Stop()

	cfg := config{
		Logger: logfmt.NewLogger(os.Stdout),
	}

	flag.Var(&cfg.Providers.GCP, "gcp.project", "GCP project ID (can be specified multiple times)")
	flag.Var(&cfg.Providers.AWS, "aws.profile", "AWS profile (can be specified multiple times)")
	flag.Var(&cfg.Providers.Azure, "azure.sub", "Azure subscription (can be specified multiple times)")

	flag.DurationVar(&cfg.Collector.Timeout, "collect.timeout", 30*time.Second, "timeout for collecting metrics from each provider")
	flag.StringVar(&cfg.Web.Path, "web.path", "/metrics", "path on which to expose metrics")
	flag.StringVar(&cfg.Web.Address, "web.address", ":8080", "address to expose metrics and web interface")
	flag.DurationVar(&cfg.Web.Timeout, "web.timeout", 5*time.Second, "timeout for shutting down the server")

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
	providers, err := clicommon.CreateProviders(ctx, cfg.Providers.GCP, cfg.Providers.AWS, cfg.Providers.Azure)
	if err != nil {
		return err
	}

	e, err := newExporter(ctx, providers, cfg)
	if err != nil {
		return fmt.Errorf("creating exporter: %w", err)
	}

	if err := prometheus.Register(e); err != nil {
		return fmt.Errorf("registering Prometheus exporter: %w", err)
	}

	if err := runWebServer(ctx, cfg); err != nil {
		return fmt.Errorf("running web server: %w", err)
	}

	return nil
}
