package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/grafana/unused"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

var ErrMissingProject = errors.New("missing project id")

var _ unused.Provider = &provider{}

type provider struct {
	project string
	svc     *compute.DisksService
	meta    unused.Meta
}

func (p *provider) Name() string { return "GCP" }

func (p *provider) Meta() unused.Meta { return p.meta }

func NewProvider(ctx context.Context, project string, meta unused.Meta, opts ...option.ClientOption) (unused.Provider, error) {
	if project == "" {
		return nil, ErrMissingProject
	}

	svc, err := compute.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot create compute service: %w", err)
	}

	if meta == nil {
		meta = make(unused.Meta)
	}

	return &provider{
		project: project,
		svc:     compute.NewDisksService(svc),
		meta:    meta,
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

					m, err := diskMetadata(d)
					if err != nil {
						// TODO do something with this error
					}
					disks = append(disks, &disk{d, p, m})
				}
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("listing unused disks: %w", err)
	}

	return disks, nil
}

func diskMetadata(d *compute.Disk) (unused.Meta, error) {
	m := make(unused.Meta)

	// GCP sends Kubernetes metadata as a JSON string in the
	// Description field.
	if d.Description != "" {
		if err := json.Unmarshal([]byte(d.Description), &m); err != nil {
			return nil, fmt.Errorf("cannot decode JSON description for disk %s: %w", d.Name, err)
		}
	}

	// Zone is returned as a URL, remove all but the zone name
	m["zone"] = d.Zone[strings.LastIndexByte(d.Zone, '/')+1:]

	return m, nil
}

func (p *provider) Delete(ctx context.Context, disk unused.Disk) error {
	_, err := p.svc.Delete(p.project, disk.Meta()["zone"], disk.Name()).Do()
	if err != nil {
		return fmt.Errorf("cannot delete GCP disk: %w", err)
	}
	return nil
}
