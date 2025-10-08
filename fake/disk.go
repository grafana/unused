//go:build fake

package fake

import (
	"strconv"
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
	m := unused.Meta{
		"name":     d.name,
		"provider": d.provider.Name(),
	}

	if r.Int()%6 == 0 {
		// no additional metadata
		return m
	}

	ns := namespaces[r.IntN(len(namespaces))]
	pvc := pvcs[r.IntN(len(pvcs))]
	m["kubernetes.io/created-for/pvc/namespace"] = ns
	m["kubernetes.io/created-for/pvc/name"] = "pvc-" + pvc + "-" + strconv.Itoa(r.IntN(20))
	m["kubernetes.io/created-for/pv/name"] = "pv-" + ns + "-" + pvc + "-" + strconv.Itoa(r.IntN(2048))

	return m
}

// Name implements unused.Disk.
func (d Disk) Name() string { return d.name }

// Provider implements unused.Disk.
func (d Disk) Provider() unused.Provider { return d.provider }

// SizeBytes implements unused.Disk.
func (d Disk) SizeBytes() float64 { return float64(d.bytes) }

// SizeGB implements unused.Disk.
func (d Disk) SizeGB() int { return int(d.bytes / 1024 / 1024 / 1024) }

var namespaces = []string{
	"frontend",
	"default",
	"app",
	"databases",
}

var pvcs = []string{
	"ingester",
	"web",
	"data",
}
