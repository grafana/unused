package unusedtest

import (
	"time"

	"github.com/grafana/unused"
)

var _ unused.Disk = disk{}

type disk struct {
	id, name  string
	provider  unused.Provider
	createdAt time.Time
	meta      unused.Meta
}

func NewDisk(name string, provider unused.Provider, createdAt time.Time) disk {
	return disk{name, name, provider, createdAt, nil}
}

func (d disk) ID() string                { return d.name }
func (d disk) Provider() unused.Provider { return d.provider }
func (d disk) Name() string              { return d.name }
func (d disk) CreatedAt() time.Time      { return d.createdAt }
func (d disk) Meta() unused.Meta         { return d.meta }
func (d disk) LastUsedAt() time.Time     { return time.Time{} }
