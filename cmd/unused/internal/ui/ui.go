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
}

type DisplayFunc func(ctx context.Context, options Options) error

const (
	KubernetesNS  = "__k8s:ns__"
	KubernetesPV  = "__k8s:pv__"
	KubernetesPVC = "__k8s:pvc__"
)
