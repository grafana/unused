package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/grafana/unused"
	"google.golang.org/api/compute/v1"
)

var ErrMissingProject = errors.New("missing project id")

var _ unused.Provider = &Provider{}

type Provider struct {
	project string
	svc     *compute.Service
	meta    unused.Meta
}

func (p *Provider) Name() string { return "GCP" }

func (p *Provider) Meta() unused.Meta { return p.meta }

func NewProvider(svc *compute.Service, project string, meta unused.Meta) (*Provider, error) {
	if project == "" {
		return nil, ErrMissingProject
	}

	if meta == nil {
		meta = make(unused.Meta)
	}

	return &Provider{
		project: project,
		svc:     svc,
		meta:    meta,
	}, nil
}

func (p *Provider) ListUnusedDisks(ctx context.Context) (unused.Disks, error) {
	var disks unused.Disks

	err := p.svc.Disks.AggregatedList(p.project).Filter("").Pages(ctx,
		func(res *compute.DiskAggregatedList) error {
			for _, item := range res.Items {
				for _, d := range item.Disks {
					if len(d.Users) > 0 {
						continue
					}

					m, err := diskMetadata(d)
					if err != nil {
						return fmt.Errorf("reading disk metadata: %w", err)
					}
					disks = append(disks, &Disk{d, p, m})
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

func (p *Provider) Delete(ctx context.Context, disk unused.Disk) error {
	_, err := p.svc.Disks.Delete(p.project, disk.Meta()["zone"], disk.Name()).Do()
	if err != nil {
		return fmt.Errorf("cannot delete GCP disk: %w", err)
	}
	return nil
}
