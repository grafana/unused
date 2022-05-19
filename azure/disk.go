package azure

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/grafana/unused"
)

var _ unused.Disk = &Disk{}

type Disk struct {
	compute.Disk
	provider *Provider
	meta     unused.Meta
}

func (d *Disk) ID() string { return *d.Disk.ID }

func (d *Disk) Provider() unused.Provider { return d.provider }

func (d *Disk) Name() string { return *d.Disk.Name }

func (d *Disk) CreatedAt() time.Time { return d.Disk.TimeCreated.ToTime() }

func (d *Disk) Meta() unused.Meta { return d.meta }

func (d *Disk) LastUsedAt() time.Time { return time.Time{} }
