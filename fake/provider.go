//go:build fake

// Package fake provides fake Provider and Disk implementations useful
// for E2E testing the binaries in cmd/ directory.
package fake

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/grafana/unused"
)

var r = rand.New(rand.NewPCG(1, 2))

var _ unused.Provider = &Provider{}

type Provider struct {
	name  string
	disks []Disk
}

func NewProvider(name string, n int) *Provider {
	p := Provider{
		name:  name,
		disks: make([]Disk, n),
	}

	now := time.Now()

	for i := range n {
		dur := -24 * time.Duration(r.IntN(365+365/2)) * time.Hour

		p.disks[i] = Disk{
			name:       fmt.Sprintf("pvc-%03d-%05d-%05d", i, r.Uint64(), r.Uint64()),
			bytes:      r.Uint64N(100 * unused.GiBbytes),
			provider:   &p,
			createdAt:  now.Add(dur),
			lastUsedAt: now.Add(dur / 2),
		}
	}

	return &p
}

// Delete implements unused.Provider.
func (p *Provider) Delete(ctx context.Context, disk unused.Disk) error {
	time.Sleep(time.Duration(r.Int64N(300)) * time.Millisecond)
	if c := r.Int(); c%7 == 1 {
		return fmt.Errorf("failed to delete: code %d", c)
	}

	for i, d := range p.disks {
		if d.name == disk.Name() {
			p.disks[i].deleted = true
		}
	}

	return nil
}

// ID implements unused.Provider.
func (p Provider) ID() string { return p.name }

// ListUnusedDisks implements unused.Provider.
func (p Provider) ListUnusedDisks(ctx context.Context) (unused.Disks, error) {
	ds := make([]unused.Disk, 0, len(p.disks))
	for _, d := range p.disks {
		time.Sleep(time.Duration(r.IntN(50)) * time.Millisecond)
		if !d.deleted {
			ds = append(ds, d)
		}
	}
	return ds, nil
}

// Meta implements unused.Provider.
func (p Provider) Meta() unused.Meta {
	return unused.Meta{
		"name": p.ID(),
	}
}

// Name implements unused.Provider.
func (p Provider) Name() string { return strings.Title(p.name) }
