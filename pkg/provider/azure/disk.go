package azure

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/grafana/unused-pds/pkg/unused"
)

var _ unused.Disk = &disk{}

type disk struct {
	compute.Disk
}

func (d *disk) Provider() string { return "Azure" }

func (d *disk) Name() string { return *d.Disk.Name }

func (d *disk) CreatedAt() time.Time { return d.Disk.TimeCreated.ToTime() }
