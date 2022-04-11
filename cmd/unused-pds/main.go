package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/grafana/unused-pds/cmd/unused-pds/ui"
	// "github.com/mmcloughlin/profile"
)

func main() {
	// defer profile.Start(profile.CPUProfile, profile.MemProfile).Stop()

	var gcpProjects, awsProfiles, azureSubs stringSlice

	flag.Var(&gcpProjects, "gcp.project", "GCP project ID (can be specified multiple times)")
	flag.Var(&awsProfiles, "aws.profile", "AWS profile (can be specified multiple times)")
	flag.Var(&azureSubs, "azure.sub", "Azure subscription (can be specified multiple times)")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := realMain(ctx, gcpProjects, awsProfiles, azureSubs); err != nil {
		cancel() // cleanup resources
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func realMain(ctx context.Context, gcpProjects, awsProfiles, azureSubs []string) error {
	providers, err := createProviders(ctx, gcpProjects, awsProfiles, azureSubs)
	if err != nil {
		return err
	}

	disks, err := listUnusedDisks(ctx, providers)
	if err != nil {
		return err
	}

	return ui.DumpAsTable(os.Stdout, disks)
}