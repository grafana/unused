package unusedtest

import (
	"time"

	"github.com/grafana/unused"
)

var _ unused.Disk = Disk{}

// Disk implements [unused.Disk] for testing purposes.
type Disk struct {
	id, name  string
	provider  unused.Provider
	createdAt time.Time
	meta      unused.Meta
	size      int
	diskType  unused.DiskType
}

// NewDisk returns a new test disk.
func NewDisk(name string, provider unused.Provider, createdAt time.Time) Disk {
	return Disk{name, name, provider, createdAt, nil, 0, unused.Unknown}
}

func (d Disk) ID() string                { return d.name }
func (d Disk) Provider() unused.Provider { return d.provider }
func (d Disk) Name() string              { return d.name }
func (d Disk) CreatedAt() time.Time      { return d.createdAt }
func (d Disk) Meta() unused.Meta         { return d.meta }
func (d Disk) LastUsedAt() time.Time     { return d.createdAt.Add(1 * time.Minute) }
func (d Disk) SizeGB() int               { return d.size }
func (d Disk) SizeBytes() float64        { return float64(d.size) * unused.GiBbytes }
func (d Disk) DiskType() unused.DiskType { return d.diskType }

func (d *Disk) SetMeta(m unused.Meta) { d.meta = m }
