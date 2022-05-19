package gcp

import (
	"fmt"
	"time"

	"github.com/grafana/unused"
	"google.golang.org/api/compute/v1"
)

// ensure we are properly defining the interface
var _ unused.Disk = &Disk{}

type Disk struct {
	*compute.Disk
	provider *Provider
	meta     unused.Meta
}

func (d *Disk) ID() string { return fmt.Sprintf("gcp-disk-%d", d.Disk.Id) }

func (d *Disk) Provider() unused.Provider { return d.provider }

func (d *Disk) Name() string { return d.Disk.Name }

func (d *Disk) CreatedAt() time.Time {
	// it's safe to assume GCP will send a valid timestamp
	c, _ := time.Parse(time.RFC3339, d.Disk.CreationTimestamp)

	return c
}

func (d *Disk) Meta() unused.Meta { return d.meta }

func (d *Disk) LastUsedAt() time.Time {
	// it's safe to assume GCP will send a valid timestamp
	t, _ := time.Parse(time.RFC3339, d.Disk.LastDetachTimestamp)
	return t
}
