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

func TestDisk_LastUsedAt(t *testing.T) {
	createdAt := time.Date(2021, 7, 16, 5, 55, 0, 0, time.UTC)
	volumeID := "vol-123"

	d := &Disk{
		Volume: types.Volume{
			VolumeId:   &volumeID,
			CreateTime: &createdAt,
		},
	}

	// AWS doesn't track last used time, should return zero time
	if got := d.LastUsedAt(); !got.IsZero() {
		t.Errorf("LastUsedAt() = %v, want zero time", got)
	}
}

func TestDisk_Type(t *testing.T) {
	tests := []struct {
		name       string
		volumeType types.VolumeType
		expected   unused.DiskType
	}{
		{"gp2", types.VolumeTypeGp2, unused.SSD},
		{"gp3", types.VolumeTypeGp3, unused.SSD},
		{"io1", types.VolumeTypeIo1, unused.SSD},
		{"io2", types.VolumeTypeIo2, unused.SSD},
		{"st1", types.VolumeTypeSt1, unused.HDD},
		{"sc1", types.VolumeTypeSc1, unused.HDD},
		{"standard", types.VolumeTypeStandard, unused.HDD},
		{"unknown", types.VolumeType("unknown"), unused.Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volumeID := "vol-123"
			d := &Disk{
				Volume: types.Volume{
					VolumeId:   &volumeID,
					VolumeType: tt.volumeType,
				},
			}

			if got := d.DiskType(); got != tt.expected {
				t.Errorf("DiskType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDisk_Name(t *testing.T) {
	tests := []struct {
		name     string
		tags     []types.Tag
		expected string
	}{
		{
			name: "Name tag present",
			tags: []types.Tag{
				{Key: aws.String("Name"), Value: aws.String("my-disk")},
			},
			expected: "my-disk",
		},
		{
			name: "CSIVolumeName tag present",
			tags: []types.Tag{
				{Key: aws.String("CSIVolumeName"), Value: aws.String("csi-disk")},
			},
			expected: "csi-disk",
		},
		{
			name: "Name takes precedence over CSIVolumeName",
			tags: []types.Tag{
				{Key: aws.String("Name"), Value: aws.String("my-disk")},
				{Key: aws.String("CSIVolumeName"), Value: aws.String("csi-disk")},
			},
			expected: "my-disk",
		},
		{
			name:     "no Name tag",
			tags:     []types.Tag{{Key: aws.String("Other"), Value: aws.String("value")}},
			expected: "",
		},
		{
			name:     "empty tags",
			tags:     []types.Tag{},
			expected: "",
		},
		{
			name:     "nil tags",
			tags:     nil,
			expected: "",
		},
		{
			name: "empty Name tag value",
			tags: []types.Tag{
				{Key: aws.String("Name"), Value: aws.String("")},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volumeID := "vol-123"
			d := &Disk{
				Volume: types.Volume{
					VolumeId: &volumeID,
					Tags:     tt.tags,
				},
			}

			if got := d.Name(); got != tt.expected {
				t.Errorf("Name() = %q, want %q", got, tt.expected)
			}
		})
	}
}
