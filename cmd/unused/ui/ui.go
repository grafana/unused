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
	Verbose      bool
}

type UI interface {
	Display(ctx context.Context, options Options) error
}
