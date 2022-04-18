package unusedtest

import (
	"context"
	"testing"

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

func (p *provider) ListUnusedDisks(ctx context.Context) (unused.Disks, error) {
	return p.disks, nil
}

func TestProviderMeta(t *testing.T, newProvider func(meta unused.Meta) (unused.Provider, error)) {
	t.Helper()

	tests := map[string]unused.Meta{
		"empty": nil,
		"respect values": map[string]string{
			"foo": "bar",
		},
	}

	for name, expMeta := range tests {
		t.Run(name, func(t *testing.T) {
			p, err := newProvider(expMeta)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			meta := p.Meta()
			if meta == nil {
				t.Error("expecting metadata, got nil")
			}

			if exp, got := len(expMeta), len(meta); exp != got {
				t.Errorf("expecting %d metadata value, got %d", exp, got)
			}
			for k, v := range expMeta {
				if exp, got := v, meta[k]; exp != got {
					t.Errorf("expecting metadata %q with value %q, got %q", k, exp, got)
				}
			}
		})
	}
}
