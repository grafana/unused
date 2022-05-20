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
}

type UI interface {
	Display(ctx context.Context, disks unused.Disks, extraColumns []string) error
}
