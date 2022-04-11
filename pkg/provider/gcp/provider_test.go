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
		tests := map[string]unused.Meta{
			"empty": nil,
			"respect values": map[string]string{
				"foo": "bar",
			},
		}

		for name, expMeta := range tests {
			t.Run(name, func(t *testing.T) {
				p, err := gcp.NewProvider(context.Background(), "my-provider", expMeta)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				meta := p.Meta()
				if meta == nil {
					t.Error("expecting metadata, got nil")
				}

				if exp, got := len(expMeta), len(meta); exp != got {
					t.Errorf("expecting %d metadata value, got %d", exp, got)
				}
				for k, v := range expMeta {
					if exp, got := v, meta[k]; exp != got {
						t.Errorf("expecting metadata %q with value %q, got %q", k, exp, got)
					}
				}
			})
		}
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
