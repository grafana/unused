package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/grafana/unused/cmd/clicommon"
	"github.com/grafana/unused/cmd/unused/ui"
	// "github.com/mmcloughlin/profile"
)

func main() {
	// defer profile.Start(profile.CPUProfile, profile.MemProfile).Stop()

	var (
		gcpProjects, awsProfiles, azureSubs clicommon.StringSliceFlag

		interactiveMode, verbose bool

		filter clicommon.FilterFlag

		extraColumns clicommon.StringSliceFlag
	)

	flag.Var(&gcpProjects, "gcp.project", "GCP project ID (can be specified multiple times)")
	flag.Var(&awsProfiles, "aws.profile", "AWS profile (can be specified multiple times)")
	flag.Var(&azureSubs, "azure.sub", "Azure subscription (can be specified multiple times)")

	flag.BoolVar(&interactiveMode, "i", false, "Interactive UI mode")
	flag.BoolVar(&verbose, "v", false, "Verbose mode")

	flag.Var(&filter, "filter", "Filter by disk metadata")

	flag.Var(&extraColumns, "add-column", "Display additional column with metadata")

	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	providers, err := clicommon.CreateProviders(ctx, gcpProjects, awsProfiles, azureSubs)
	if err != nil {
		cancel()
		fmt.Fprintln(os.Stderr, "creating providers:", err)
		os.Exit(1)
	}

	opts := ui.Options{
		Providers:    providers,
		ExtraColumns: extraColumns,
		Verbose:      verbose,
		Filter: ui.Filter{
			Key:   filter.Key,
			Value: filter.Value,
		},
	}

	var display ui.DisplayFunc
	if interactiveMode {
		display = ui.Interactive
	} else {
		display = ui.Table
	}

	if err := display(ctx, opts); err != nil {
		cancel() // cleanup resources
		fmt.Fprintln(os.Stderr, "displaying output:", err)
		os.Exit(1)
	}
}
