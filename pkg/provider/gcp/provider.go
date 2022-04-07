package gcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/grafana/unused-pds/pkg/unused"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

var ErrMissingProject = errors.New("missing project id")

var _ unused.Provider = &provider{}

type provider struct {
	project string
	svc     *compute.DisksService
}

func (p *provider) Name() string { return "GCP" }

func NewProvider(ctx context.Context, project string, opts ...option.ClientOption) (unused.Provider, error) {
	if project == "" {
		return nil, ErrMissingProject
	}

	svc, err := compute.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot create compute service: %w", err)
	}

	return &provider{
		project: project,
		svc:     compute.NewDisksService(svc),
	}, nil
}

func (p *provider) ListUnusedDisks(ctx context.Context) (unused.Disks, error) {
	var disks unused.Disks

	err := p.svc.AggregatedList(p.project).Filter("").Pages(ctx,
		func(res *compute.DiskAggregatedList) error {
			for _, item := range res.Items {
				for _, d := range item.Disks {
					if len(d.Users) > 0 {
						continue
					}

					disks = append(disks, &disk{d})
				}
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("listing unused disks: %w", err)
	}

	return disks, nil
}
