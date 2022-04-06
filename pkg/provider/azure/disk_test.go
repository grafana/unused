package azure

import (
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/grafana/unused-pds/pkg/unused"
)

func TestDisk(t *testing.T) {
	createdAt := time.Date(2021, 7, 16, 5, 55, 00, 0, time.UTC)
	name := "my-disk"

	var d unused.Disk = &disk{
		compute.Disk{
			Name: &name,
			DiskProperties: &compute.DiskProperties{
				TimeCreated: &date.Time{Time: createdAt},
			},
		},
	}

	if exp, got := "Azure", d.Provider(); exp != got {
		t.Errorf("expecting Provider() %q, got %q", exp, got)
	}

	if exp, got := "my-disk", d.Name(); exp != got {
		t.Errorf("expecting Name() %q, got %q", exp, got)
	}

	if !createdAt.Equal(d.CreatedAt()) {
		t.Errorf("expecting CreatedAt() %v, got %v", createdAt, d.CreatedAt())
	}
}
