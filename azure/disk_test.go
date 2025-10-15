package azure

import (
	"testing"
	"time"

	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/grafana/unused"
)

func TestDisk(t *testing.T) {
	createdAt := time.Date(2021, 7, 16, 5, 55, 00, 0, time.UTC)
	name := "my-disk"
	id := "my-disk-id"
	sizeGB := int32(10)
	sizeBytes := int64(10_737_418_240)
	sku := compute.DiskStorageAccountTypesStandardSSDLRS
	lastUsedAt := createdAt.Add(3 * 24 * time.Hour)

	var d unused.Disk = &Disk{
		&compute.Disk{
			ID:   &id,
			Name: &name,
			SKU: &compute.DiskSKU{
				Name: &sku,
			},
			Properties: &compute.DiskProperties{
				TimeCreated:   &createdAt,
				DiskSizeGB:    &sizeGB,
				DiskSizeBytes: &sizeBytes,

				LastOwnershipUpdateTime: &lastUsedAt,
			},
		},
		nil,
		nil,
	}

	if exp, got := "my-disk-id", d.ID(); exp != got {
		t.Errorf("expecting ID() %q, got %q", exp, got)
	}

	if exp, got := "Azure", d.Provider().Name(); exp != got {
		t.Errorf("expecting Provider() %q, got %q", exp, got)
	}

	if exp, got := "my-disk", d.Name(); exp != got {
		t.Errorf("expecting Name() %q, got %q", exp, got)
	}

	if exp, got := unused.SSD, d.DiskType(); exp != got {
		t.Errorf("expecting DiskType() %q, got %q", exp, got)
	}

	if !createdAt.Equal(d.CreatedAt()) {
		t.Errorf("expecting CreatedAt() %v, got %v", createdAt, d.CreatedAt())
	}

	if !lastUsedAt.Equal(d.LastUsedAt()) {
		t.Errorf("expecting LastUsedAt() %v, got %v", lastUsedAt, d.LastUsedAt())
	}

	if exp, got := int(sizeGB), d.SizeGB(); exp != got {
		t.Errorf("expecting SizeGB() %d, got %d", exp, got)
	}

	if exp, got := float64(sizeBytes), d.SizeBytes(); exp != got {
		t.Errorf("expecting SizeBytes() %f, got %f", exp, got)
	}

	t.Run("special case disk never used", func(t *testing.T) {
		dd := d.(*Disk)
		dd.Disk.Properties.LastOwnershipUpdateTime = nil

		if !d.CreatedAt().Equal(d.LastUsedAt()) {
			t.Errorf("expecting LastUsedAt() to be the same as CreatedAt() %v, got %v", d.CreatedAt(), d.LastUsedAt())
		}
	})
}
