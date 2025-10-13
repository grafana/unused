// unused is a CLI tool to query the given providers for unused disks.
//
// In its default operation mode it outputs a table listing all the
// unused disks. I also supports an interactive mode where the user
// can see mark unused disks from the listing tables to individually
// delete them.
//
// Provider selection is opinionated, currently accepting the
// following authentication method for each provider:
//   - GCP: pass gcp.project with a valid GCP project ID.
//   - AWS: pass aws.profile with a valid AWS shared profile.
//   - Azure: pass azure.sub with a valid Azure subscription ID.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/grafana/unused/cmd/internal"
	"github.com/grafana/unused/cmd/unused/internal/ui"
)

func main() {
	var (
		gcpProjects, awsProfiles, azureSubs internal.StringSliceFlag

		out ui.UI
	)

	internal.ProviderFlags(flag.CommandLine, &gcpProjects, &awsProfiles, &azureSubs)

	flag.BoolVar(&out.Interactive, "i", false, "Interactive UI mode")
	flag.BoolVar(&out.Verbose, "v", false, "Verbose mode")
	flag.BoolVar(&out.DryRun, "n", false, "Do not delete disks in interactive mode")
	flag.BoolVar(&out.CSV, "csv", false, "Output results in CSV form")

	flag.Func("filter", "Filter by disk metadata; use k8s:ns, k8s:pvc or k8s:pv for Kubernetes metadata", func(v string) error {
		ps := strings.SplitN(v, "=", 2)

		if len(ps) == 0 || ps[0] == "" {
			return errors.New("invalid filter format")
		}

		out.Filters.Key = ps[0]

		if len(ps) == 2 {
			out.Filters.Value = ps[1]
		}

		return nil
	})

	flag.Func("min-age", "Minimum age of the disk to be listed (ex: 365d or 36h)", func(s string) error {
		dur, err := internal.ParseAge(s)
		if err != nil {
			return err
		}

		out.Filters.MinAge = dur

		return nil
	})

	flag.Func("add-column", "Display additional column with metadata", func(c string) error {
		out.ExtraColumns = append(out.ExtraColumns, c)
		return nil
	})

	flag.Func("add-k8s-column", "Add Kubernetes metadata column; valid values are: ns, pvc, pv", func(c string) error {
		switch c {
		case "ns":
			out.ExtraColumns = append(out.ExtraColumns, ui.KubernetesNS)
		case "pvc":
			out.ExtraColumns = append(out.ExtraColumns, ui.KubernetesPVC)
		case "pv":
			out.ExtraColumns = append(out.ExtraColumns, ui.KubernetesPV)
		default:
			return errors.New("valid values are ns, pvc, pv")
		}

		return nil
	})

	flag.Func("group-by", "Group by disk metadata values; use k8s:ns, k8s:pvc, or k8s:pv for Kubernetes metadata", func(s string) error {
		out.Group = s
		return nil
	})

	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	providers, err := internal.CreateProviders(ctx, logger, gcpProjects, awsProfiles, azureSubs)
	if err != nil {
		cancel()
		fmt.Fprintln(os.Stderr, "creating providers:", err)
		os.Exit(1)
	}

	out.Providers = providers

	if err := out.Run(ctx); err != nil {
		cancel() // cleanup resources
		fmt.Fprintln(os.Stderr, "displaying output:", err)
		os.Exit(1)
	}
}
