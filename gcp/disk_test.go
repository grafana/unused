package gcp

import (
	"testing"
	"time"

	"github.com/grafana/unused"
	"github.com/grafana/unused/unusedtest"
	compute "google.golang.org/api/compute/v1"
)

func TestDisk(t *testing.T) {
	createdAt := time.Date(2021, 7, 16, 5, 55, 00, 0, time.UTC)
	detachedAt := createdAt.Add(1 * time.Hour)
	size := 10

	var d unused.Disk = &Disk{
		&compute.Disk{
			Id:                  1234,
			Name:                "my-disk",
			CreationTimestamp:   createdAt.Format(time.RFC3339),
			LastDetachTimestamp: detachedAt.Format(time.RFC3339),
			SizeGb:              int64(size),
		},
		nil,
		unused.Meta{"foo": "bar"},
	}

	if exp, got := "gcp-disk-1234", d.ID(); exp != got {
		t.Errorf("expecting ID() %q, got %q", exp, got)
	}

	if exp, got := "GCP", d.Provider().Name(); exp != got {
		t.Errorf("expecting Provider() %q, got %q", exp, got)
	}

	if exp, got := "my-disk", d.Name(); exp != got {
		t.Errorf("expecting Name() %q, got %q", exp, got)
	}

	if !createdAt.Equal(d.CreatedAt()) {
		t.Errorf("expecting CreatedAt() %v, got %v", createdAt, d.CreatedAt())
	}

	if !detachedAt.Equal(d.LastUsedAt()) {
		t.Errorf("expecting LastUsedAt() %v, got %v", detachedAt, d.LastUsedAt())
	}

	if exp, got := size, d.SizeGB(); exp != got {
		t.Errorf("expecting SizeGB() %d, got %d", exp, got)
	}

	if exp, got := float64(size)*unused.GiBbytes, d.SizeBytes(); exp != got {
		t.Errorf("expecting SizeBytes() %f, got %f", exp, got)
	}

	err := unusedtest.AssertEqualMeta(unused.Meta{"foo": "bar"}, d.Meta())
	if err != nil {
		t.Fatalf("metadata doesn't match: %v", err)
	}

	t.Run("special case disk never used", func(t *testing.T) {
		dd := d.(*Disk)
		dd.LastDetachTimestamp = ""

		if !d.CreatedAt().Equal(d.LastUsedAt()) {
			t.Errorf("expecting LastUsedAt() to be the same as CreatedAt() %v, got %v", d.CreatedAt(), d.LastUsedAt())
		}
	})
}

func TestDisk_Type(t *testing.T) {
	tests := []struct {
		name     string
		diskType string
		expected unused.DiskType
	}{
		{"pd-ssd", "https://www.googleapis.com/compute/v1/projects/my-project/zones/us-central1-a/diskTypes/pd-ssd", unused.SSD},
		{"pd-standard", "https://www.googleapis.com/compute/v1/projects/my-project/zones/us-central1-a/diskTypes/pd-standard", unused.HDD},
		{"pd-balanced", "https://www.googleapis.com/compute/v1/projects/my-project/zones/us-central1-a/diskTypes/pd-balanced", unused.Unknown},
		{"unknown type", "https://www.googleapis.com/compute/v1/projects/my-project/zones/us-central1-a/diskTypes/pd-extreme", unused.Unknown},
		{"simple pd-ssd", "pd-ssd", unused.SSD},
		{"simple pd-standard", "pd-standard", unused.HDD},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Disk{
				Disk: &compute.Disk{
					Type: tt.diskType,
				},
			}

			if got := d.DiskType(); got != tt.expected {
				t.Errorf("DiskType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// SetMeta is defined as a Disk method ONLY for tests.
func (d *Disk) SetMeta(m unused.Meta) { d.meta = m }
