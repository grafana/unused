package unusedtest

import (
	"context"
	"errors"
	"fmt"

	"github.com/grafana/unused"
)

var _ unused.Provider = &Provider{}

// Provider implements [unused.Provider] for testing purposes.
type Provider struct {
	name  string
	disks unused.Disks
	meta  unused.Meta
}

// NewProvider returns a new test provider that return the given disks
// as unused.
func NewProvider(name string, meta unused.Meta, disks ...unused.Disk) *Provider {
	if meta == nil {
		meta = make(unused.Meta)
	}
	return &Provider{name, disks, meta}
}

func (p *Provider) Name() string { return p.name }

func (p *Provider) ID() string { return "my-id" }

func (p *Provider) Meta() unused.Meta { return p.meta }

func (p *Provider) SetMeta(meta unused.Meta) { p.meta = meta }

func (p *Provider) ListUnusedDisks(ctx context.Context) (unused.Disks, error) {
	return p.disks, nil
}

var ErrDiskNotFound = errors.New("disk not found")

func (p *Provider) Delete(ctx context.Context, disk unused.Disk) error {
	for i := range p.disks {
		if disk.Name() == p.disks[i].Name() {
			p.disks = append(p.disks[:i], p.disks[i+1:]...)
			return nil
		}
	}

	return ErrDiskNotFound
}

// TestProviderMeta returns nil if the provider properly implements
// storing metadata.
//
// It accepts a constructor function that should return a valid
// [unused.Provider] or an error when it isn't compliant with the
// semantics of creating a provider.
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
