package ui

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/grafana/unused"
)

type Filter struct {
	Key, Value string
	MinAge     time.Duration
}

type UI struct {
	Filter       Filter
	Group        string
	Providers    []unused.Provider
	ExtraColumns []string
	Verbose      bool
	DryRun       bool
	CSV          bool
	Interactive  bool
}

func (ui UI) FilterFunc(d unused.Disk) bool {
	minAge := ui.Filter.MinAge == 0 || time.Since(d.CreatedAt()) >= ui.Filter.MinAge
	keyVal := ui.Filter.Key == "" || d.Meta().Matches(ui.Filter.Key, ui.Filter.Value)

	return minAge && keyVal
}

func (ui UI) Run(ctx context.Context) error {
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
		wg    sync.WaitGroup
		mu    sync.Mutex
		total unused.Disks
	)

	wg.Add(len(ui.Providers))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, len(ui.Providers))

	for _, p := range ui.Providers {
		go func(p unused.Provider) {
			defer wg.Done()

			disks, err := p.ListUnusedDisks(ctx)
			if err != nil {
				cancel()
				errCh <- fmt.Errorf("%s %s: %w", p.Name(), p.Meta(), err)
				return
			}

			mu.Lock()
			disks = disks.Filter(ui.FilterFunc)
			total = append(total, disks...)
			mu.Unlock()
		}(p)
	}

	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
	}

	return total, nil
}
