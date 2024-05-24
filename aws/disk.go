package aws

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/grafana/unused"
)

var _ unused.Disk = &Disk{}

// Disk holds information about an AWS EC2 volume.
type Disk struct {
	types.Volume
	provider *Provider
	meta     unused.Meta
}

// ID returns the volume ID of this AWS EC2 volume.
func (d *Disk) ID() string { return *d.Volume.VolumeId }

// Provider returns a reference to the provider used to instantiate
// this disk.
func (d *Disk) Provider() unused.Provider { return d.provider }

// Name returns the name of this AWS EC2 volume.
//
// AWS EC2 volumes do not have a name property, instead they store the
// name in tags. This method will try to find the Name or
// CSIVolumeName, otherwise it will return empty.
func (d *Disk) Name() string {
	for _, t := range d.Volume.Tags {
		if *t.Key == "Name" || *t.Key == "CSIVolumeName" {
			return *t.Value
		}
	}
	return ""
}

// CreatedAt returns the time when the AWS EC2 volume was created.
func (d *Disk) CreatedAt() time.Time { return *d.Volume.CreateTime }

// Meta returns the disk metadata.
func (d *Disk) Meta() unused.Meta { return d.meta }

// SizeGB returns the size of this AWS EC2 volume in GiB.
func (d *Disk) SizeGB() int { return int(*d.Volume.Size) }

// SizeBytes returns the size of this AWS EC2 volume in bytes.
func (d *Disk) SizeBytes() float64 { return float64(*d.Volume.Size) * unused.GiBbytes }

// LastUsedAt returns a zero [time.Time] value, as AWS does not
// provide this information.
func (d *Disk) LastUsedAt() time.Time { return time.Time{} }

// DiskType Type returns the type of this AWS EC2 volume.
func (d *Disk) DiskType() unused.DiskType {
	switch d.Volume.VolumeType {
	case types.VolumeTypeGp2, types.VolumeTypeGp3, types.VolumeTypeIo1, types.VolumeTypeIo2:
		return unused.SSD
	case types.VolumeTypeSt1, types.VolumeTypeSc1, types.VolumeTypeStandard:
		return unused.HDD
	default:
		return unused.Unknown
	}
}
