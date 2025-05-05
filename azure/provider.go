package azure

import (
	"context"
	"errors"
	"fmt"
	"strings"

	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/grafana/unused"
)

var ProviderName = "Azure"

var _ unused.Provider = &Provider{}

const ResourceGroupMetaKey = "resource-group"

// Provider implements [unused.Provider] for Azure.
type Provider struct {
	client *compute.DisksClient
	meta   unused.Meta
}

// Name returns Azure.
func (p *Provider) Name() string { return ProviderName }

// Meta returns the provider metadata.
func (p *Provider) Meta() unused.Meta { return p.meta }

// ID returns the subscription for this provider.
func (p *Provider) ID() string { return p.meta["SubscriptionID"] }

var ErrInvalidSubscriptionID = errors.New("invalid subscription ID in metadata")

// NewProvider creates a new Azure [unused.Provider].
//
// A valid Azure compute disks client must be supplied in order to
// list the unused resources.
func NewProvider(client *compute.DisksClient, meta unused.Meta) (*Provider, error) {
	if meta == nil {
		meta = make(unused.Meta)
	}
	if sid, ok := meta["SubscriptionID"]; !ok || sid == "" {
		return nil, ErrInvalidSubscriptionID
	}

	return &Provider{client: client, meta: meta}, nil
}

// ListUnusedDisks returns all the Azure compute disks that are not
// managed by other resources.
func (p *Provider) ListUnusedDisks(ctx context.Context) (unused.Disks, error) {
	var upds unused.Disks

	pages := p.client.NewListPager(&compute.DisksClientListOptions{})

	prefix := fmt.Sprintf("/subscriptions/%s/resourceGroups/", p.meta["SubscriptionID"])

	for pages.More() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing Azure disks: %w", err)
		}
		for _, d := range page.Value {
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
	}

	return upds, nil
}

// Delete deletes the given disk from Azure.
func (p *Provider) Delete(ctx context.Context, disk unused.Disk) error {
	poller, err := p.client.BeginDelete(ctx, disk.Meta()[ResourceGroupMetaKey], disk.Name(), nil)
	if err != nil {
		return fmt.Errorf("cannot delete Azure disk: failed to finish request: %w", err)
	}

	if _, err := poller.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("cannot delete Azure disk: %w", err)
	}

	return nil
}
