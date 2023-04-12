package ui

import (
	"context"

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
}

type DisplayFunc func(ctx context.Context, options Options) error
