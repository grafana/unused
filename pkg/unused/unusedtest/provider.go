package unusedtest

import (
	"context"

	"github.com/grafana/unused-pds/pkg/unused"
)

var _ unused.Provider = &provider{}

type provider struct {
	name  string
	disks unused.Disks
}

func NewProvider(name string, disks ...unused.Disk) *provider {
	return &provider{name, disks}
}

func (p *provider) Name() string { return p.name }

func (p *provider) ListUnusedDisks(ctx context.Context) (unused.Disks, error) {
	return p.disks, nil
}
