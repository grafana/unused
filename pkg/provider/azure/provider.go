package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/grafana/unused-pds/pkg/unused"
)

var _ unused.Provider = &provider{}

type provider struct {
	client compute.DisksClient
	meta   unused.Meta
}

func (p *provider) Name() string { return "Azure" }

func (p *provider) Meta() unused.Meta { return p.meta }

type OptionFunc func(c *compute.DisksClient)

func WithBaseURI(uri string) OptionFunc {
	return func(c *compute.DisksClient) {
		c.BaseURI = uri
	}
}

func WithAuthorizer(authorizer autorest.Authorizer) OptionFunc {
	return func(c *compute.DisksClient) {
		c.Authorizer = authorizer
	}
}

func NewProvider(subID string, meta unused.Meta, opts ...OptionFunc) (unused.Provider, error) {
	c := compute.NewDisksClient(subID)
	for _, o := range opts {
		o(&c)
	}

	if meta == nil {
		meta = make(unused.Meta)
	}

	return &provider{client: c, meta: meta}, nil
}

func (p *provider) ListUnusedDisks(ctx context.Context) (unused.Disks, error) {
	var upds unused.Disks

	res, err := p.client.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing Azure disks: %w", err)
	}

	for res.NotDone() {
		for _, d := range res.Values() {
			if d.ManagedBy != nil {
				continue
			}
			m := unused.Meta{
				"location": *d.Location,
			}
			upds = append(upds, &disk{d, p, m})
		}

		err := res.NextWithContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot advance page: %w", err)
		}
	}

	return upds, nil
}
