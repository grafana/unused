package unusedtest

import (
	"context"
	"errors"
	"fmt"

	"github.com/grafana/unused"
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

func (p *provider) SetMeta(meta unused.Meta) { p.meta = meta }

func (p *provider) ListUnusedDisks(ctx context.Context) (unused.Disks, error) {
	return p.disks, nil
}

var ErrDiskNotFound = errors.New("disk not found")

func (p *provider) Delete(ctx context.Context, disk unused.Disk) error {
	for i := range p.disks {
		if disk.Name() == p.disks[i].Name() {
			p.disks = append(p.disks[:i], p.disks[i+1:]...)
			return nil
		}
	}

	return ErrDiskNotFound
}

func TestProviderMeta(newProvider func(meta unused.Meta) (unused.Provider, error)) error {
	tests := map[string]unused.Meta{
		"nil":   nil,
		"empty": map[string]string{},
		"respect values": map[string]string{
			"foo": "bar",
		},
	}

	for name, expMeta := range tests {
		p, err := newProvider(expMeta)
		if err != nil {
			return fmt.Errorf("%s: unexpected error: %v", name, err)
		}

		meta := p.Meta()
		if meta == nil {
			return fmt.Errorf("%s: expecting metadata, got nil", name)
		}

		if exp, got := len(expMeta), len(meta); exp != got {
			return fmt.Errorf("%s: expecting %d metadata value, got %d", name, exp, got)
		}
		for k, v := range expMeta {
			if exp, got := v, meta[k]; exp != got {
				return fmt.Errorf("%s: expecting metadata %q with value %q, got %q", name, k, exp, got)
			}
		}
	}

	return nil
}
