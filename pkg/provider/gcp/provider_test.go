package gcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grafana/unused-pds/pkg/provider/gcp"
	"github.com/grafana/unused-pds/pkg/unused"
	"github.com/grafana/unused-pds/pkg/unused/unusedtest"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func TestNewProvider(t *testing.T) {
	t.Run("project is required", func(t *testing.T) {
		p, err := gcp.NewProvider(context.Background(), "", nil)
		if !errors.Is(err, gcp.ErrMissingProject) {
			t.Fatalf("expecting error %v, got %v", gcp.ErrMissingProject, err)
		}
		if p != nil {
			t.Fatalf("expecting nil provider, got %v", p)
		}
	})

	t.Run("metadata", func(t *testing.T) {
		unusedtest.TestProviderMeta(t, func(meta unused.Meta) (unused.Provider, error) {
			return gcp.NewProvider(context.Background(), "my-provider", meta)
		})
	})
}

func TestProviderListUnusedDisks(t *testing.T) {
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// are we requesting the right API endpoint?
		if got, exp := req.URL.Path, "/projects/my-project/aggregated/disks"; exp != got {
			t.Fatalf("expecting request to %s, got %s", exp, got)
		}

		res := &compute.DiskAggregatedList{
			Items: map[string]compute.DisksScopedList{
				"foo": {
					Disks: []*compute.Disk{
						{Name: "disk-1"},
						{Name: "with-users", Users: []string{"inkel"}},
						{Name: "disk-2"},
					},
				},
			},
		}

		b, _ := json.Marshal(res)
		w.Write(b)
	}))
	defer ts.Close()

	p, err := gcp.NewProvider(ctx, "my-project", nil, option.WithEndpoint(ts.URL))
	if err != nil {
		t.Fatal("unexpected error creating provider:", err)
	}

	disks, err := p.ListUnusedDisks(ctx)
	if err != nil {
		t.Fatal("unexpected error listing unused disks:", err)
	}

	if exp, got := 2, len(disks); exp != got {
		t.Errorf("expecting %d disks, got %d", exp, got)
	}
}
