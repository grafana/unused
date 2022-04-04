package azure_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/grafana/unused-pds/pkg/provider/azure"
)

func TestNewProvider(t *testing.T) {
	subID := "my-subscription"
	p, err := azure.NewProvider(subID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p == nil {
		t.Fatal("expecting provider")
	}
}

func TestListUnusedDisks(t *testing.T) {
	// Azure is really strange when it comes to marhsaling JSON, so,
	// yeah, this is an awful hack.
	mock := func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(`{
"value": [
  {"name":"disk-1","managedBy":"grafana"},
  {"name":"disk-2"},
  {"name":"disk-3","managedBy":"grafana"}
]
}`))
	}

	var (
		ctx   = context.Background()
		subID = "my-subscription"
		ts    = httptest.NewServer(http.HandlerFunc(mock))
		c     = compute.NewDisksClientWithBaseURI(ts.URL, subID)
	)
	defer ts.Close()

	disks, err := azure.ListUnusedDisks(ctx, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if exp, got := 1, len(disks); exp != got {
		t.Errorf("expecting %d disks, got %d", exp, got)
	}
}
