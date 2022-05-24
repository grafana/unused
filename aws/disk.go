package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	ctypes "github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
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

func (d *Disk) RefreshLastUsedAt(ctx context.Context) error {
	c := cloudtrail.NewFromConfig(d.provider.cfg)

	res, err := c.LookupEvents(ctx, &cloudtrail.LookupEventsInput{
		StartTime: d.Volume.CreateTime,
		LookupAttributes: []ctypes.LookupAttribute{
			{
				AttributeKey:   ctypes.LookupAttributeKeyResourceName,
				AttributeValue: d.Volume.VolumeId,
			},
			{
				AttributeKey:   ctypes.LookupAttributeKeyEventName,
				AttributeValue: aws.String("DetachVolume"),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("cannot lookup DetachVolume events for Volume %s: %v", *d.VolumeId, err)
	}

	for _, e := range res.Events {
		if *e.EventName == "DetachVolume" {
			d.lastUsed = *e.EventTime
			break
		}
	}

	return nil
}
