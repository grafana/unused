package azure

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/grafana/unused"
)

var _ unused.Disk = &disk{}

type disk struct {
	compute.Disk
	provider *provider
	meta     unused.Meta
}

func (d *disk) Provider() unused.Provider { return d.provider }

func (d *disk) Name() string { return *d.Disk.Name }

func (d *disk) CreatedAt() time.Time { return d.Disk.TimeCreated.ToTime() }

func (d *disk) Meta() unused.Meta { return d.meta }
