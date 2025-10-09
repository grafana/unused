package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/grafana/unused"
	"golang.org/x/sync/errgroup"
)

type Filters struct {
	Key, Value string
	MinAge     time.Duration
}

type UI struct {
	Filters      Filters
	Group        string
	Providers    []unused.Provider
	ExtraColumns []string
	Verbose      bool
	DryRun       bool
	CSV          bool
	Interactive  bool
	Out          io.Writer
}

func (ui UI) Filter(d unused.Disk) bool {
	minAge := ui.Filters.MinAge == 0 || time.Since(d.CreatedAt()) >= ui.Filters.MinAge
	keyVal := ui.Filters.Key == "" || d.Meta().Matches(ui.Filters.Key, ui.Filters.Value)

	return minAge && keyVal
}

func (ui UI) Run(ctx context.Context) error {
	if ui.Out == nil {
		ui.Out = os.Stdout
	}

	var display func(ctx context.Context, ui UI) error

	if ui.Interactive {
		display = Interactive
	} else if ui.Group != "" && !ui.CSV {
		display = GroupTable
	} else if ui.CSV {
		display = CSV
	} else {
		display = Table
	}

	return display(ctx, ui)
}

const (
	KubernetesNS  = "__k8s:ns__"
	KubernetesPV  = "__k8s:pv__"
	KubernetesPVC = "__k8s:pvc__"
)

func (ui UI) listUnusedDisks(ctx context.Context) (unused.Disks, error) {
	var (
		mu    sync.Mutex
		total unused.Disks
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(len(ui.Providers))

	for _, p := range ui.Providers {
		g.Go(func() error {
			disks, err := p.ListUnusedDisks(ctx)
			if err != nil {
				return fmt.Errorf("%s %s: %w", p.Name(), p.Meta(), err)
			}

			mu.Lock()
			disks = disks.Filter(ui.Filter)
			total = append(total, disks...)
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("listing disks: %w", err)
	}

	return total, nil
}
