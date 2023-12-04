package azure_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/grafana/unused"
	"github.com/grafana/unused/azure"
	"github.com/grafana/unused/unusedtest"
)

func TestNewProvider(t *testing.T) {
	c := compute.NewDisksClient("my-subscription")
	p, err := azure.NewProvider(c, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p == nil {
		t.Fatal("expecting provider")
	}
}

func TestProviderMeta(t *testing.T) {
	err := unusedtest.TestProviderMeta(func(meta unused.Meta) (unused.Provider, error) {
		c := compute.NewDisksClient("my-subscription")
		return azure.NewProvider(c, meta)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListUnusedDisks(t *testing.T) {
	// Azure is really strange when it comes to marhsaling JSON, so,
	// yeah, this is an awful hack.
	mock := func(w http.ResponseWriter, req *http.Request) {
		_, err := w.Write([]byte(`{
"value": [
  {"name":"disk-1","managedBy":"grafana"},
  {"name":"disk-2","location":"germanywestcentral","tags": {
      "created-by": "kubernetes-azure-dd",
      "kubernetes.io-created-for-pv-name": "pvc-prometheus-1",
      "kubernetes.io-created-for-pvc-name": "prometheus-1",
      "kubernetes.io-created-for-pvc-namespace": "monitoring"
  },"id":"/subscriptions/my-subscription/resourceGroups/RGNAME/providers/Microsoft.Compute/disks/disk-2"},
  {"name":"disk-3","managedBy":"grafana"}
]
}`))
		if err != nil {
			t.Fatalf("unexpected error writing response: %v", err)
		}
	}

	var (
		ctx   = context.Background()
		subID = "my-subscription"
		ts    = httptest.NewServer(http.HandlerFunc(mock))
	)
	defer ts.Close()

	c := compute.NewDisksClient(subID)
	c.BaseURI = ts.URL

	p, err := azure.NewProvider(c, nil)
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

	err = unusedtest.AssertEqualMeta(unused.Meta{
		"location":                                "germanywestcentral",
		"created-by":                              "kubernetes-azure-dd",
		"kubernetes.io-created-for-pv-name":       "pvc-prometheus-1",
		"kubernetes.io-created-for-pvc-name":      "prometheus-1",
		"kubernetes.io-created-for-pvc-namespace": "monitoring",
		azure.ResourceGroupMetaKey:                "RGNAME",
	}, disks[0].Meta())
	if err != nil {
		t.Fatalf("metadata doesn't match: %v", err)
	}
}
