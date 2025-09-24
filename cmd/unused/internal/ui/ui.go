package ui

import (
	"context"
	"time"

	"github.com/grafana/unused"
)

type UI struct {
	Filter struct {
		Key, Value string
	}
	Group        string
	Providers    []unused.Provider
	ExtraColumns []string
	MinAge       time.Duration
	Verbose      bool
	DryRun       bool
	CSV          bool
	Interactive  bool
}

func (o UI) FilterFunc(d unused.Disk) bool {
	minAge := o.MinAge == 0 || time.Since(d.CreatedAt()) >= o.MinAge
	keyVal := o.Filter.Key == "" || d.Meta().Matches(o.Filter.Key, o.Filter.Value)

	return minAge && keyVal
}

func (o UI) Run(ctx context.Context) error {
	var display func(ctx context.Context, ui UI) error

	if o.Interactive {
		display = Interactive
	} else if o.Group != "" && !o.CSV {
		display = GroupTable
	} else if o.CSV {
		display = CSV
	} else {
		display = Table
	}

	return display(ctx, o)
}

const (
	KubernetesNS  = "__k8s:ns__"
	KubernetesPV  = "__k8s:pv__"
	KubernetesPVC = "__k8s:pvc__"
)
