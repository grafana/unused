package ui

import (
	"context"

	"github.com/grafana/unused"
)

type Displayer interface {
	Display(ctx context.Context, disks unused.Disks) error
}
