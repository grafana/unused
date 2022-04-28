package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/grafana/unused/cli"
	"github.com/grafana/unused/cmd/unused/ui"
	"github.com/grafana/unused/cmd/unused/ui/interactive"
	"github.com/grafana/unused/cmd/unused/ui/table"
	// "github.com/mmcloughlin/profile"
)

func main() {
	// defer profile.Start(profile.CPUProfile, profile.MemProfile).Stop()

	var gcpProjects, awsProfiles, azureSubs cli.StringSliceFlag
	flag.Var(&gcpProjects, "gcp.project", "GCP project ID (can be specified multiple times)")
	flag.Var(&awsProfiles, "aws.profile", "AWS profile (can be specified multiple times)")
	flag.Var(&azureSubs, "azure.sub", "Azure subscription (can be specified multiple times)")

	var interactiveMode, verbose bool

	flag.BoolVar(&interactiveMode, "i", false, "Interactive UI mode")
	flag.BoolVar(&verbose, "v", false, "Verbose mode")

	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var out ui.UI
	if interactiveMode {
		out = interactive.New(verbose)
	} else {
		out = table.New(os.Stdout, verbose)
	}

	if err := realMain(ctx, out, gcpProjects, awsProfiles, azureSubs); err != nil {
		cancel() // cleanup resources
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func realMain(ctx context.Context, out ui.UI, gcpProjects, awsProfiles, azureSubs []string) error {
	providers, err := cli.CreateProviders(ctx, gcpProjects, awsProfiles, azureSubs)
	if err != nil {
		return err
	}

	disks, err := listUnusedDisks(ctx, providers)
	if err != nil {
		return err
	}

	return out.Display(ctx, disks)
}
