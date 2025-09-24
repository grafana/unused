package ui

import (
	"context"
	"time"

	"github.com/grafana/unused"
)

type Filter struct {
	Key, Value string
}

type Options struct {
	Providers    []unused.Provider
	ExtraColumns []string
	Filter       Filter
	Group        string
	Verbose      bool
	DryRun       bool
	MinAge       time.Duration
	CSV          bool
	Interactive  bool
}

func (o Options) FilterFunc(d unused.Disk) bool {
	minAge := o.MinAge == 0 || time.Since(d.CreatedAt()) >= o.MinAge
	keyVal := o.Filter.Key == "" || d.Meta().Matches(o.Filter.Key, o.Filter.Value)

	return minAge && keyVal
}

type DisplayFunc func(ctx context.Context, options Options) error

const (
	KubernetesNS  = "__k8s:ns__"
	KubernetesPV  = "__k8s:pv__"
	KubernetesPVC = "__k8s:pvc__"
)
