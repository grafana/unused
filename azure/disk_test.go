package azure

import (
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/grafana/unused"
)

func TestDisk(t *testing.T) {
	createdAt := time.Date(2021, 7, 16, 5, 55, 00, 0, time.UTC)
	name := "my-disk"
	id := "my-disk-id"

	var d unused.Disk = &Disk{
		compute.Disk{
			ID:   &id,
			Name: &name,
			Sku: &compute.DiskSku{
				Name: compute.StandardSSDLRS,
			},
			DiskProperties: &compute.DiskProperties{
				TimeCreated: &date.Time{Time: createdAt},
			},
		},
		&Provider{name: DefaultProviderName},
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

	if !d.LastUsedAt().IsZero() {
		t.Errorf("Azure doesn't provide a last usage timestamp for disks, got %v", d.LastUsedAt())
	}
}
