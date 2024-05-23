package aws

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/grafana/unused"
)

func TestDisk(t *testing.T) {
	createdAt := time.Date(2021, 7, 16, 5, 55, 00, 0, time.UTC)
	size := int32(10)

	for _, keyName := range []string{"Name", "CSIVolumeName"} {
		t.Run(keyName, func(t *testing.T) {
			var d unused.Disk = &Disk{
				types.Volume{
					VolumeId:   aws.String("my-disk-id"),
					CreateTime: &createdAt,
					Tags: []types.Tag{
						{
							Key:   aws.String(keyName),
							Value: aws.String("my-disk"),
						},
					},
					Size: &size,
				},
				nil,
				nil,
			}

			if exp, got := "my-disk-id", d.ID(); exp != got {
				t.Errorf("expecting ID() %q, got %q", exp, got)
			}

			if exp, got := "AWS", d.Provider().Name(); exp != got {
				t.Errorf("expecting Provider() %q, got %q", exp, got)
			}

			if exp, got := "my-disk", d.Name(); exp != got {
				t.Errorf("expecting Name() %q, got %q", exp, got)
			}

			if !createdAt.Equal(d.CreatedAt()) {
				t.Errorf("expecting CreatedAt() %v, got %v", createdAt, d.CreatedAt())
			}

			if exp, got := int(size), d.SizeGB(); exp != got {
				t.Errorf("expecting SizeGB() %d, got %d", exp, got)
			}

			if exp, got := float64(size)*unused.GiBbytes, d.SizeBytes(); exp != got {
				t.Errorf("expecting SizeBytes() %f, got %f", exp, got)
			}
		})
	}
}
