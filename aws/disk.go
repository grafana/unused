package aws

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/grafana/unused"
)

var _ unused.Disk = &disk{}

type disk struct {
	types.Volume
	provider *provider
	meta     unused.Meta
}

func (d *disk) ID() string { return *d.Volume.VolumeId }

func (d *disk) Provider() unused.Provider { return d.provider }

func (d *disk) Name() string {
	for _, t := range d.Volume.Tags {
		if *t.Key == "Name" || *t.Key == "CSIVolumeName" {
			return *t.Value
		}
	}
	return ""
}

func (d *disk) CreatedAt() time.Time { return *d.Volume.CreateTime }

func (d *disk) Meta() unused.Meta { return d.meta }

func (d *disk) LastUsedAt() time.Time { return time.Time{} }
