package gcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/grafana/unused"
	"github.com/grafana/unused/gcp"
	"github.com/grafana/unused/unusedtest"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func TestNewProvider(t *testing.T) {
	l := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("project is required", func(t *testing.T) {
		p, err := gcp.NewProvider(l, nil, "", nil)
		if !errors.Is(err, gcp.ErrMissingProject) {
			t.Fatalf("expecting error %v, got %v", gcp.ErrMissingProject, err)
		}
		if p != nil {
			t.Fatalf("expecting nil provider, got %v", p)
		}
	})

	t.Run("provider information is correct", func(t *testing.T) {
		p, err := gcp.NewProvider(l, nil, "my-project", unused.Meta{})
		if err != nil {
			t.Fatalf("error creating provider: %v", err)
		}
		if p == nil {
			t.Fatalf("error creating provider, provider is nil")
		}

		if exp, got := "my-project", p.ID(); exp != got {
			t.Fatalf("provider id was incorrect, exp: %v, got: %v", exp, got)
		}
	})

	t.Run("metadata", func(t *testing.T) {
		err := unusedtest.TestProviderMeta(func(meta unused.Meta) (unused.Provider, error) {
			svc, err := compute.NewService(context.Background(), option.WithAPIKey("123abc"))
			if err != nil {
				t.Fatalf("unexpected error creating GCP compute service: %v", err)
			}
			return gcp.NewProvider(l, svc, "my-provider", meta)
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestProviderListUnusedDisks(t *testing.T) {
	ctx := context.Background()
	l := slog.New(slog.NewTextHandler(io.Discard, nil))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// are we requesting the right API endpoint?
		if got, exp := req.URL.Path, "/projects/my-project/aggregated/disks"; exp != got {
			t.Fatalf("expecting request to %s, got %s", exp, got)
		}

		res := &compute.DiskAggregatedList{
			Items: map[string]compute.DisksScopedList{
				"foo": {
					Disks: []*compute.Disk{
						{Name: "disk-1", Zone: "https://www.googleapis.com/compute/v1/projects/ops-tools-1203/zones/us-central1-a"},
						{Name: "with-users", Users: []string{"inkel"}},
						{Name: "disk-2", Zone: "eu-west2-b", Description: `{"kubernetes.io-created-for-pv-name":"pvc-prometheus-1","kubernetes.io-created-for-pvc-name":"prometheus-1","kubernetes.io-created-for-pvc-namespace":"monitoring"}`},
					},
				},
			},
		}

		b, _ := json.Marshal(res)
		_, err := w.Write(b)
		if err != nil {
			t.Fatalf("unexpected error writing response: %v", err)
		}
	}))
	defer ts.Close()

	svc, err := compute.NewService(context.Background(), option.WithAPIKey("123abc"), option.WithEndpoint(ts.URL))
	if err != nil {
		t.Fatalf("unexpected error creating GCP compute service: %v", err)
	}

	p, err := gcp.NewProvider(l, svc, "my-project", nil)
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

	err = unusedtest.AssertEqualMeta(unused.Meta{"zone": "us-central1-a"}, disks[0].Meta())
	if err != nil {
		t.Fatalf("metadata doesn't match: %v", err)
	}
	err = unusedtest.AssertEqualMeta(unused.Meta{
		"zone":                                    "eu-west2-b",
		"kubernetes.io-created-for-pv-name":       "pvc-prometheus-1",
		"kubernetes.io-created-for-pvc-name":      "prometheus-1",
		"kubernetes.io-created-for-pvc-namespace": "monitoring",
	}, disks[1].Meta())
	if err != nil {
		t.Fatalf("metadata doesn't match: %v", err)
	}

	t.Run("disk without JSON in description", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// are we requesting the right API endpoint?
			if got, exp := req.URL.Path, "/projects/my-project/aggregated/disks"; exp != got {
				t.Fatalf("expecting request to %s, got %s", exp, got)
			}

			res := &compute.DiskAggregatedList{
				Items: map[string]compute.DisksScopedList{
					"foo": {
						Disks: []*compute.Disk{
							{Name: "disk-2", Zone: "eu-west2-b", Description: "some string that isn't JSON"},
						},
					},
				},
			}

			b, _ := json.Marshal(res)
			_, err := w.Write(b)
			if err != nil {
				t.Fatalf("unexpected error writing response %v", err)
			}
		}))
		defer ts.Close()

		svc, err := compute.NewService(context.Background(), option.WithAPIKey("123abc"), option.WithEndpoint(ts.URL))
		if err != nil {
			t.Fatalf("unexpected error creating GCP compute service: %v", err)
		}

		var buf bytes.Buffer
		l := slog.New(slog.NewTextHandler(&buf, nil))

		p, err := gcp.NewProvider(l, svc, "my-project", nil)
		if err != nil {
			t.Fatal("unexpected error creating provider:", err)
		}

		disks, err := p.ListUnusedDisks(ctx)
		if err != nil {
			t.Fatal("unexpected error listing unused disks:", err)
		}

		if len(disks) != 1 {
			t.Fatalf("expecting 1 unused disk, got %d", len(disks))
		}

		// check that we logged about it
		m, _ := regexp.MatchString(`msg="cannot parse disk metadata".+disk=disk-2`, buf.String())
		if !m {
			t.Fatal("expecting a log line to be emitted")
		}
	})
}
