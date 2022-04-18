package azure_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grafana/unused/azure"
	"github.com/grafana/unused"
	"github.com/grafana/unused/unusedtest"
)

func TestNewProvider(t *testing.T) {
	subID := "my-subscription"
	p, err := azure.NewProvider(subID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p == nil {
		t.Fatal("expecting provider")
	}
}

func TestProviderMeta(t *testing.T) {
	unusedtest.TestProviderMeta(t, func(meta unused.Meta) (unused.Provider, error) {
		return azure.NewProvider("my-subscription", meta)
	})
}

func TestListUnusedDisks(t *testing.T) {
	// Azure is really strange when it comes to marhsaling JSON, so,
	// yeah, this is an awful hack.
	mock := func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(`{
"value": [
  {"name":"disk-1","managedBy":"grafana"},
  {"name":"disk-2","location":"germanywestcentral","tags": {
      "created-by": "kubernetes-azure-dd",
      "kubernetes.io-created-for-pv-name": "pvc-prometheus-1",
      "kubernetes.io-created-for-pvc-name": "prometheus-1",
      "kubernetes.io-created-for-pvc-namespace": "monitoring"
  }},
  {"name":"disk-3","managedBy":"grafana"}
]
}`))
	}

	var (
		ctx   = context.Background()
		subID = "my-subscription"
		ts    = httptest.NewServer(http.HandlerFunc(mock))
	)
	defer ts.Close()

	p, err := azure.NewProvider(subID, nil, azure.WithBaseURI(ts.URL))
	if err != nil {
		t.Fatalf("unexpected error creating provider: %v", err)
	}

	disks, err := p.ListUnusedDisks(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if exp, got := 1, len(disks); exp != got {
		t.Errorf("expecting %d disks, got %d", exp, got)
	}

	unusedtest.AssertEqualMeta(t, unused.Meta{
		"location":                                "germanywestcentral",
		"created-by":                              "kubernetes-azure-dd",
		"kubernetes.io-created-for-pv-name":       "pvc-prometheus-1",
		"kubernetes.io-created-for-pvc-name":      "prometheus-1",
		"kubernetes.io-created-for-pvc-namespace": "monitoring",
	}, disks[0].Meta())
}
