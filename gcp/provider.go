package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/grafana/unused"
	compute "google.golang.org/api/compute/v1"
)

var ProviderName = "GCP"

// ErrMissingProject is the error used when no project ID is provided
// when trying to create a provider.
var ErrMissingProject = errors.New("missing project id")

var _ unused.Provider = &Provider{}

// Provider implements [unused.Provider] for GCP.
type Provider struct {
	project string
	svc     *compute.Service
	meta    unused.Meta
	logger  *slog.Logger
}

// Name returns GCP.
func (p *Provider) Name() string { return ProviderName }

// Meta returns the provider metadata.
func (p *Provider) Meta() unused.Meta { return p.meta }

// ID returns the GCP project for this provider.
func (p *Provider) ID() string { return p.project }

// NewProvider creates a new GCP [unused.Provider].
//
// A valid GCP compute service must be supplied in order to listed the
// unused resources. It also requires a valid project ID which should
// be the project where the disks were created.
func NewProvider(logger *slog.Logger, svc *compute.Service, project string, meta unused.Meta) (*Provider, error) {
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
		logger:  logger,
	}, nil
}

// ListUnusedDisks returns all the GCP compute disks that aren't
// associated to any users, meaning that are not being in use.
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
						p.logger.Error("cannot parse disk metadata",
							slog.String("project", p.project),
							slog.String("disk", d.Name),
							slog.String("err", err.Error()),
						)
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

// Delete deletes the given disk from GCP.
func (p *Provider) Delete(ctx context.Context, disk unused.Disk) error {
	_, err := p.svc.Disks.Delete(p.project, disk.Meta()["zone"], disk.Name()).Do()
	if err != nil {
		return fmt.Errorf("cannot delete GCP disk: %w", err)
	}
	return nil
}
