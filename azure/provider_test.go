package azure_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/google/uuid"
	"github.com/grafana/unused"
	"github.com/grafana/unused/azure"
	"github.com/grafana/unused/unusedtest"
)

func TestNewProvider(t *testing.T) {
	c, err := compute.NewDisksClient("my-subscription", nil, nil)
	if err != nil {
		t.Fatalf("cannot create disks client: %v", err)
	}
	p, err := azure.NewProvider(c, unused.Meta{"SubscriptionID": "my-subscription"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p == nil {
		t.Fatal("expecting provider")
	}

	if exp, got := "my-subscription", p.ID(); exp != got {
		t.Fatalf("provider id was incorrect, exp: %v, got: %v", exp, got)
	}
}

func TestProviderMeta(t *testing.T) {
	t.Skip("skip this test while we figure out the right way to test this provider for metadata")
	err := unusedtest.TestProviderMeta(func(meta unused.Meta) (unused.Provider, error) {
		c, err := compute.NewDisksClient("my-subscription", nil, nil)
		if err != nil {
			t.Fatalf("cannot create disks client: %v", err)
		}
		return azure.NewProvider(c, meta)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListUnusedDisks(t *testing.T) {
	t.Skip("Azure now checks if the subscription ID exists so it fails to authenticate")
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
		subID = uuid.New().String()
		ts    = httptest.NewServer(http.HandlerFunc(mock))
	)
	defer ts.Close()

	c, err := compute.NewDisksClient(subID, nil, nil)
	if err != nil {
		t.Fatalf("cannot create disks client: %v", err)
	}
	//c.BaseURI = ts.URL

	p, err := azure.NewProvider(c, unused.Meta{"SubscriptionID": subID})
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
