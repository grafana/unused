package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	// defer profile.Start(profile.CPUProfile, profile.MemProfile).Stop()

	var gcpProjects, awsProfiles, azureSubs stringSlice
	flag.Var(&gcpProjects, "gcp.project", "GCP project ID (can be specified multiple times)")
	flag.Var(&awsProfiles, "aws.profile", "AWS profile (can be specified multiple times)")
	flag.Var(&azureSubs, "azure.sub", "Azure subscription (can be specified multiple times)")

	var (
		interval = flag.Duration("metrics.interval", 15*time.Second, "polling interval to query providers for unused disks")
		path     = flag.String("metrics.path", "/metrics", "path on which to expose metris")
		address  = flag.String("web.address", ":8080", "address to expose metrics and web interface")
	)

	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := realMain(ctx, gcpProjects, awsProfiles, azureSubs, *address, *path, *interval); err != nil {
		cancel() // cleanup resources
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func realMain(ctx context.Context, gcpProjects, awsProfiles, azureSubs []string, address, path string, interval time.Duration) error {
	providers, err := createProviders(ctx, gcpProjects, awsProfiles, azureSubs)
	if err != nil {
		return err
	}

	l := logger{os.Stdout}

	ms, err := newMetrics(l)
	if err != nil {
		return fmt.Errorf("creating metrics: %w", err)
	}

	go func() {
		l.Log("starting collection loop", "interval", interval)

		t := time.NewTicker(interval)

		for {
			ms.Collect(ctx, providers)

			select {
			case <-ctx.Done():
				t.Stop()
				l.Log("stopping collection loop")
				break
			case <-t.C:
				continue // unnecessary but expressive
			}
		}
	}()

	if err := runWebServer(ctx, l, address, path); err != nil {
		return fmt.Errorf("running web server: %w", err)
	}

	return nil
}