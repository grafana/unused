package gcp

import (
	"time"

	"github.com/grafana/unused"
	"google.golang.org/api/compute/v1"
)

// ensure we are properly defining the interface
var _ unused.Disk = &disk{}

type disk struct {
	*compute.Disk
	provider *provider
	meta     unused.Meta
}

func (d *disk) Provider() unused.Provider { return d.provider }

func (d *disk) Name() string { return d.Disk.Name }

func (d *disk) CreatedAt() time.Time {
	// it's safe to assume GCP will send a valid timestamp
	c, _ := time.Parse(time.RFC3339, d.Disk.CreationTimestamp)

	return c
}

func (d *disk) Meta() unused.Meta { return d.meta }