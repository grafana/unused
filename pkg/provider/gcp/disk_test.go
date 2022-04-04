package gcp

import (
	"testing"
	"time"

	"github.com/grafana/unused-pds/pkg/unused"
	"google.golang.org/api/compute/v1"
)

func TestDisk(t *testing.T) {
	createdAt := time.Date(2021, 7, 16, 5, 55, 00, 0, time.UTC)

	var d unused.Disk = &disk{
		&compute.Disk{
			Name:              "my-disk",
			CreationTimestamp: createdAt.Format(time.RFC3339),
		},
	}

	if exp, got := "GCP", d.Provider(); exp != got {
		t.Errorf("expecting Provider() %q, got %q", exp, got)
	}

	if exp, got := "my-disk", d.Name(); exp != got {
		t.Errorf("expecting Name() %q, got %q", exp, got)
	}

	if !createdAt.Equal(d.CreatedAt()) {
		t.Errorf("expecting CreatedAt() %v, got %v", createdAt, d.CreatedAt())
	}
}
