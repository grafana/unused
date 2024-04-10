package azure

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/grafana/unused"
)

var ProviderName = "Azure"

var _ unused.Provider = &Provider{}

const (
	ResourceGroupMetaKey     = "resource-group"
	DefaultAzureProviderName = "Azure"
)

// Provider implements [unused.Provider] for Azure.
type Provider struct {
	client compute.DisksClient
	meta   unused.Meta
}

// Name returns Azure.
func (p *Provider) Name() string { return ProviderName }

// Meta returns the provider metadata.
func (p *Provider) Meta() unused.Meta { return p.meta }

// ID returns the subscription for this provider.
func (p *Provider) ID() string { return p.client.SubscriptionID }

// NewProvider creates a new Azure [unused.Provider].
//
// A valid Azure compute disks client must be supplied in order to
// list the unused resources.
func NewProvider(client compute.DisksClient, meta unused.Meta) (*Provider, error) {
	if meta == nil {
		meta = make(unused.Meta)
	}

	return &Provider{client: client, meta: meta}, nil
}

// ListUnusedDisks returns all the Azure compute disks that are not
// managed by other resources.
func (p *Provider) ListUnusedDisks(ctx context.Context) (unused.Disks, error) {
	var upds unused.Disks

	res, err := p.client.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing Azure disks: %w", err)
	}

	prefix := fmt.Sprintf("/subscriptions/%s/resourceGroups/", p.client.SubscriptionID)

	for res.NotDone() {
		for _, d := range res.Values() {
			if d.ManagedBy != nil {
				continue
			}

			m := make(unused.Meta, len(d.Tags)+1)
			m["location"] = *d.Location
			for k, v := range d.Tags {
				m[k] = *v
			}

			// Azure doesn't return the resource group directly
			// "/subscriptions/$subscription-id/resourceGroups/$resource-group-name/providers/Microsoft.Compute/disks/$disk-name"
			rg := strings.TrimPrefix(*d.ID, prefix)
			m[ResourceGroupMetaKey] = rg[:strings.IndexRune(rg, '/')]

			upds = append(upds, &Disk{d, p, m})
		}

		err := res.NextWithContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot advance page: %w", err)
		}
	}

	return upds, nil
}

// Delete deletes the given disk from Azure.
func (p *Provider) Delete(ctx context.Context, disk unused.Disk) error {
	_, err := p.client.Delete(ctx, disk.Meta()[ResourceGroupMetaKey], disk.Name())
	if err != nil {
		return fmt.Errorf("cannot delete Azure disk: %w", err)
	}
	return nil
}
