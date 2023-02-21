package gcp

import (
	"fmt"
	"strings"
	"time"

	"github.com/grafana/unused"
	"google.golang.org/api/compute/v1"
)

// ensure we are properly defining the interface
var _ unused.Disk = &Disk{}

// Disk holds information about a GCP compute disk.
type Disk struct {
	*compute.Disk
	provider *Provider
	meta     unused.Meta
}

// ID returns the GCP compute disk ID, prefixed by gcp-disk.
func (d *Disk) ID() string { return fmt.Sprintf("gcp-disk-%d", d.Disk.Id) } // TODO remove prefix

// Provider returns a reference to the provider used to instantiate
// this disk.
func (d *Disk) Provider() unused.Provider { return d.provider }

// Name returns the name of the GCP compute disk.
func (d *Disk) Name() string { return d.Disk.Name }

// CreatedAt returns the time when the GCP compute disk was created.
func (d *Disk) CreatedAt() time.Time {
	// it's safe to assume GCP will send a valid timestamp
	c, _ := time.Parse(time.RFC3339, d.Disk.CreationTimestamp)

	return c
}

// Meta returns the disk metadata.
func (d *Disk) Meta() unused.Meta { return d.meta }

// LastUsedAt returns the time when the GCP compute disk was last
// detached.
func (d *Disk) LastUsedAt() time.Time {
	// it's safe to assume GCP will send a valid timestamp
	t, _ := time.Parse(time.RFC3339, d.Disk.LastDetachTimestamp)
	return t
}

// SizeGB returns the size of the GCP compute disk in GB.
func (d *Disk) SizeGB() int { return int(d.Disk.SizeGb) }

// DiskType Type returns the type of the GCP compute disk.
func (d *Disk) DiskType() string {
	splitDiskType := strings.Split(d.Disk.Type, "/")
	return splitDiskType[len(splitDiskType)-1]
}
