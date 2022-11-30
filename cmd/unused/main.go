package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/grafana/unused/cmd/internal"
	"github.com/grafana/unused/cmd/unused/ui"
	// "github.com/mmcloughlin/profile"
)

func main() {
	// defer profile.Start(profile.CPUProfile, profile.MemProfile).Stop()

	var (
		gcpProjects, awsProfiles, azureSubs internal.StringSliceFlag

		interactiveMode bool

		options ui.Options
	)

	internal.ProviderFlags(flag.CommandLine, &gcpProjects, &awsProfiles, &azureSubs)

	flag.BoolVar(&interactiveMode, "i", false, "Interactive UI mode")
	flag.BoolVar(&options.Verbose, "v", false, "Verbose mode")

	flag.Func("filter", "Filter by disk metadata", func(v string) error {
		ps := strings.SplitN(v, "=", 2)

		if len(ps) == 0 || ps[0] == "" {
			return errors.New("invalid filter format")
		}

		options.Filter.Key = ps[0]

		if len(ps) == 2 {
			options.Filter.Value = ps[1]
		}

		return nil
	})

	flag.Func("add-column", "Display additional column with metadata", func(c string) error {
		options.ExtraColumns = append(options.ExtraColumns, c)
		return nil
	})

	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	providers, err := internal.CreateProviders(ctx, gcpProjects, awsProfiles, azureSubs)
	if err != nil {
		cancel()
		fmt.Fprintln(os.Stderr, "creating providers:", err)
		os.Exit(1)
	}

	options.Providers = providers

	var display ui.DisplayFunc = ui.Table
	if interactiveMode {
		display = ui.Interactive
	}

	if err := display(ctx, options); err != nil {
		cancel() // cleanup resources
		fmt.Fprintln(os.Stderr, "displaying output:", err)
		os.Exit(1)
	}
}
