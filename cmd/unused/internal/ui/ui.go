package ui

import (
	"context"
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
