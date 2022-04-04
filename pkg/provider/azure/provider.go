package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/grafana/unused-pds/pkg/unused"
)

var _ unused.Provider = &provider{}

type provider struct {
	client compute.DisksClient
}

func NewProvider(subID string) (unused.Provider, error) {
	return &provider{
		client: compute.NewDisksClient(subID),
	}, nil
}

func (p *provider) ListUnusedDisks(ctx context.Context) (unused.Disks, error) {
	return ListUnusedDisks(ctx, p.client)
}

func ListUnusedDisks(ctx context.Context, c compute.DisksClient) (unused.Disks, error) {
	var upds unused.Disks

	res, err := c.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing Azure disks: %w", err)
	}

	for res.NotDone() {
		for _, d := range res.Values() {
			if d.ManagedBy != nil {
				continue
			}
			upds = append(upds, &disk{d})
		}

		err := res.NextWithContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot advance page: %w", err)
		}
	}

	return upds, nil
}
