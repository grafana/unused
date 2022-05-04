package ui

import (
	"context"

	"github.com/grafana/unused"
)

type UI interface {
	Display(ctx context.Context, disks unused.Disks, extraColumns []string) error
}
