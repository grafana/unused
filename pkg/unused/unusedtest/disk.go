package unusedtest

import (
	"time"

	"github.com/grafana/unused-pds/pkg/unused"
)

var _ unused.Disk = disk{}

type disk struct {
	name      string
	provider  unused.Provider
	createdAt time.Time
}

func NewDisk(name string, provider unused.Provider, createdAt time.Time) disk {
	return disk{name, provider, createdAt}
}

func (d disk) Provider() unused.Provider { return d.provider }
func (d disk) Name() string              { return d.name }
func (d disk) CreatedAt() time.Time      { return d.createdAt }
