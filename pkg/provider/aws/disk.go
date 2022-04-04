package aws

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/grafana/unused-pds/pkg/unused"
)

var _ unused.Disk = &disk{}

type disk struct {
	types.Volume
}

func (d *disk) Provider() string { return "AWS" }

func (d *disk) Name() string {
	for _, t := range d.Volume.Tags {
		if *t.Key == "Name" {
			return *t.Value
		}
	}
	return ""
}

func (d *disk) CreatedAt() time.Time { return *d.Volume.CreateTime }
