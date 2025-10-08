//go:build fake

package fake

import (
	"strings"
	"time"

	"github.com/grafana/unused"
)

var _ unused.Disk = Disk{}

type Disk struct {
	name       string
	createdAt  time.Time
	lastUsedAt time.Time
	provider   unused.Provider
	bytes      uint64
	deleted    bool
}

// CreatedAt implements unused.Disk.
func (d Disk) CreatedAt() time.Time { return d.createdAt }

// DiskType implements unused.Disk.
func (d Disk) DiskType() unused.DiskType {
	if d.SizeGB()%5 == 0 {
		return unused.HDD
	}
	return unused.SSD
}

// ID implements unused.Disk.
func (d Disk) ID() string { return strings.ToLower(d.name) }

// LastUsedAt implements unused.Disk.
func (d Disk) LastUsedAt() time.Time { return d.lastUsedAt }

// Meta implements unused.Disk.
func (d Disk) Meta() unused.Meta {
	return unused.Meta{
		"name":     d.name,
		"provider": d.provider.Name(),
	}
}

// Name implements unused.Disk.
func (d Disk) Name() string { return d.name }

// Provider implements unused.Disk.
func (d Disk) Provider() unused.Provider { return d.provider }

// SizeBytes implements unused.Disk.
func (d Disk) SizeBytes() float64 { return float64(d.bytes) }

// SizeGB implements unused.Disk.
func (d Disk) SizeGB() int { return int(d.bytes / 1024 / 1024 / 1024) }
