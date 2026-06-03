package azure

import (
	"testing"
	"time"

	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v8"
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
		dd.Properties.LastOwnershipUpdateTime = nil

		if !d.CreatedAt().Equal(d.LastUsedAt()) {
			t.Errorf("expecting LastUsedAt() to be the same as CreatedAt() %v, got %v", d.CreatedAt(), d.LastUsedAt())
		}
	})
}

func TestDiskType(t *testing.T) {
	tests := []struct {
		name     string
		diskSKU  compute.DiskStorageAccountTypes
		expected unused.DiskType
	}{
		{"Standard HDD", compute.DiskStorageAccountTypesStandardLRS, unused.HDD},
		{"Standard SSD", compute.DiskStorageAccountTypesStandardSSDLRS, unused.SSD},
		{"Premium SSD", compute.DiskStorageAccountTypesPremiumLRS, unused.SSD},
		{"Ultra SSD", compute.DiskStorageAccountTypesUltraSSDLRS, unused.SSD},
		{"Premium V2 SSD", compute.DiskStorageAccountTypesPremiumV2LRS, unused.Unknown},
		{"Unknown type", compute.DiskStorageAccountTypes("UnknownType_LRS"), unused.Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sku := tt.diskSKU
			d := &Disk{
				&compute.Disk{
					SKU: &compute.DiskSKU{Name: &sku},
				},
				nil,
				nil,
			}

			if got := d.DiskType(); got != tt.expected {
				t.Errorf("DiskType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDiskMeta(t *testing.T) {
	meta := unused.Meta{"key": "value"}
	d := &Disk{
		&compute.Disk{},
		nil,
		meta,
	}

	if got := d.Meta(); !got.Equals(meta) {
		t.Errorf("Meta() = %v, want %v", got, meta)
	}
}
