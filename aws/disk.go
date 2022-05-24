package aws

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/grafana/unused"
)

var _ unused.Disk = &Disk{}

type Disk struct {
	types.Volume
	provider *Provider
	meta     unused.Meta
	lastUsed time.Time
}

func (d *Disk) ID() string { return *d.Volume.VolumeId }

func (d *Disk) Provider() unused.Provider { return d.provider }

func (d *Disk) Name() string {
	for _, t := range d.Volume.Tags {
		if *t.Key == "Name" || *t.Key == "CSIVolumeName" {
			return *t.Value
		}
	}
	return ""
}

func (d *Disk) CreatedAt() time.Time { return *d.Volume.CreateTime }

func (d *Disk) Meta() unused.Meta { return d.meta }

func (d *Disk) LastUsedAt() time.Time { return d.lastUsed }
