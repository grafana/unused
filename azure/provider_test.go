package azure_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	corepolicy "github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v8"
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

func TestNewProviderErrors(t *testing.T) {
	tests := []struct {
		name    string
		meta    unused.Meta
		wantErr error
	}{
		{
			name:    "nil metadata",
			meta:    nil,
			wantErr: azure.ErrInvalidSubscriptionID,
		},
		{
			name:    "missing SubscriptionID",
			meta:    unused.Meta{},
			wantErr: azure.ErrInvalidSubscriptionID,
		},
		{
			name:    "empty SubscriptionID",
			meta:    unused.Meta{"SubscriptionID": ""},
			wantErr: azure.ErrInvalidSubscriptionID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := compute.NewDisksClient("test-sub", nil, nil)
			if err != nil {
				t.Fatalf("cannot create disks client: %v", err)
			}

			p, err := azure.NewProvider(c, tt.meta)
			if err != tt.wantErr {
				t.Errorf("NewProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
			if p != nil {
				t.Errorf("NewProvider() = %v, want nil", p)
			}
		})
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
	subID := uuid.New().String()

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
  },"id":"/subscriptions/` + subID + `/resourceGroups/RGNAME/providers/Microsoft.Compute/disks/disk-2"},
  {"name":"disk-3","managedBy":"grafana"}
]
}`))
		if err != nil {
			t.Fatalf("unexpected error writing response: %v", err)
		}
	}

	var (
		ctx = context.Background()
		ts  = httptest.NewServer(http.HandlerFunc(mock))
	)
	defer ts.Close()

	c, err := compute.NewDisksClient(subID, nil, &policy.ClientOptions{
		ClientOptions: corepolicy.ClientOptions{
			Cloud: cloud.Configuration{
				Services: map[cloud.ServiceName]cloud.ServiceConfiguration{
					cloud.ResourceManager: {
						Audience: "public",
						Endpoint: ts.URL,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("cannot create disks client: %v", err)
	}

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
		t.Log(disks[0].Meta())
		t.Fatalf("metadata doesn't match: %v", err)
	}
}

func TestProviderMethods(t *testing.T) {
	c, err := compute.NewDisksClient("test-subscription", nil, nil)
	if err != nil {
		t.Fatalf("cannot create disks client: %v", err)
	}

	meta := unused.Meta{
		"SubscriptionID": "test-subscription",
		"extra":          "metadata",
	}

	p, err := azure.NewProvider(c, meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("Name", func(t *testing.T) {
		if got := p.Name(); got != azure.ProviderName {
			t.Errorf("Name() = %v, want %v", got, azure.ProviderName)
		}
	})

	t.Run("Meta", func(t *testing.T) {
		if got := p.Meta(); !got.Equals(meta) {
			t.Errorf("Meta() = %v, want %v", got, meta)
		}
	})

	t.Run("ID", func(t *testing.T) {
		if got := p.ID(); got != "test-subscription" {
			t.Errorf("ID() = %v, want test-subscription", got)
		}
	})
}

func TestProviderDelete(t *testing.T) {
	subID := uuid.New().String()

	t.Run("successful deletion", func(t *testing.T) {
		mock := func(w http.ResponseWriter, req *http.Request) {
			if req.Method != "DELETE" {
				t.Errorf("expected DELETE request, got %s", req.Method)
			}
			// Return 200 OK to indicate synchronous completion (no polling needed)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}

		ctx := context.Background()
		ts := httptest.NewServer(http.HandlerFunc(mock))
		defer ts.Close()

		c, err := compute.NewDisksClient(subID, nil, &policy.ClientOptions{
			ClientOptions: corepolicy.ClientOptions{
				Cloud: cloud.Configuration{
					Services: map[cloud.ServiceName]cloud.ServiceConfiguration{
						cloud.ResourceManager: {
							Audience: "public",
							Endpoint: ts.URL,
						},
					},
				},
			},
		})
		if err != nil {
			t.Fatalf("cannot create disks client: %v", err)
		}

		p, err := azure.NewProvider(c, unused.Meta{"SubscriptionID": subID})
		if err != nil {
			t.Fatalf("unexpected error creating provider: %v", err)
		}

		disk := &azure.Disk{
			Disk: &compute.Disk{
				Name: new("test-disk"),
			},
		}
		disk.SetMeta(unused.Meta{azure.ResourceGroupMetaKey: "test-rg"})

		err = p.Delete(ctx, disk)
		if err != nil {
			t.Errorf("Delete() unexpected error: %v", err)
		}
	})

	t.Run("deletion error", func(t *testing.T) {
		mock := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"code":"ResourceNotFound","message":"The disk does not exist."}}`))
		}

		ctx := context.Background()
		ts := httptest.NewServer(http.HandlerFunc(mock))
		defer ts.Close()

		c, err := compute.NewDisksClient(subID, nil, &policy.ClientOptions{
			ClientOptions: corepolicy.ClientOptions{
				Cloud: cloud.Configuration{
					Services: map[cloud.ServiceName]cloud.ServiceConfiguration{
						cloud.ResourceManager: {
							Audience: "public",
							Endpoint: ts.URL,
						},
					},
				},
			},
		})
		if err != nil {
			t.Fatalf("cannot create disks client: %v", err)
		}

		p, err := azure.NewProvider(c, unused.Meta{"SubscriptionID": subID})
		if err != nil {
			t.Fatalf("unexpected error creating provider: %v", err)
		}

		disk := &azure.Disk{
			Disk: &compute.Disk{
				Name: new("nonexistent-disk"),
			},
		}
		disk.SetMeta(unused.Meta{azure.ResourceGroupMetaKey: "test-rg"})

		err = p.Delete(ctx, disk)
		if err == nil {
			t.Error("Delete() expected error, got nil")
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		mock := func(w http.ResponseWriter, req *http.Request) {
			// Simulate a slow operation
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusAccepted)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		ts := httptest.NewServer(http.HandlerFunc(mock))
		defer ts.Close()

		c, err := compute.NewDisksClient(subID, nil, &policy.ClientOptions{
			ClientOptions: corepolicy.ClientOptions{
				Cloud: cloud.Configuration{
					Services: map[cloud.ServiceName]cloud.ServiceConfiguration{
						cloud.ResourceManager: {
							Audience: "public",
							Endpoint: ts.URL,
						},
					},
				},
			},
		})
		if err != nil {
			t.Fatalf("cannot create disks client: %v", err)
		}

		p, err := azure.NewProvider(c, unused.Meta{"SubscriptionID": subID})
		if err != nil {
			t.Fatalf("unexpected error creating provider: %v", err)
		}

		disk := &azure.Disk{
			Disk: &compute.Disk{
				Name: new("test-disk"),
			},
		}
		disk.SetMeta(unused.Meta{azure.ResourceGroupMetaKey: "test-rg"})

		err = p.Delete(ctx, disk)
		if err == nil {
			t.Error("Delete() expected context cancellation error, got nil")
		}
	})
}
