package unusedtest

import (
	"context"

	"github.com/grafana/unused-pds/pkg/unused"
)

var _ unused.Provider = &provider{}

type provider struct {
	name  string
	disks unused.Disks
	meta  unused.Meta
}

func NewProvider(name string, meta unused.Meta, disks ...unused.Disk) *provider {
	if meta == nil {
		meta = make(unused.Meta)
	}
	return &provider{name, disks, meta}
}

func (p *provider) Name() string { return p.name }

func (p *provider) Meta() unused.Meta { return p.meta }

func (p *provider) ListUnusedDisks(ctx context.Context) (unused.Disks, error) {
	return p.disks, nil
}
