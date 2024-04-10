package aws

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/grafana/unused"
)

var _ unused.Provider = &Provider{}

var ProviderName = "AWS"

// Provider implements [unused.Provider] for AWS.
type Provider struct {
	client *ec2.Client
	meta   unused.Meta
	logger *slog.Logger
}

// Name returns AWS.
func (p *Provider) Name() string { return ProviderName }

// Meta returns the provider metadata.
func (p *Provider) Meta() unused.Meta { return p.meta }

// ID returns the profile of this provider.
func (p *Provider) ID() string { return p.meta["profile"] }

// NewProvider creates a new AWS [unused.Provider].
//
// A valid EC2 client must be supplied in order to list the unused
// resources. The metadata passed will be used to identify the
// provider.
func NewProvider(logger *slog.Logger, client *ec2.Client, meta unused.Meta) (*Provider, error) {
	if meta == nil {
		meta = make(unused.Meta)
	}

	return &Provider{
		client: client,
		meta:   meta,
		logger: logger,
	}, nil
}

// ListUnusedDisks returns all the AWS EC2 volumes that are available,
// ie. not used by any other resource.
func (p *Provider) ListUnusedDisks(ctx context.Context) (unused.Disks, error) {
	params := &ec2.DescribeVolumesInput{
		Filters: []types.Filter{
			// only show available (i.e. not "in-use") volumes
			{
				Name:   aws.String("status"),
				Values: []string{string(types.VolumeStateAvailable)},
			},
			// exclude snapshots
			{
				Name:   aws.String("snapshot-id"),
				Values: []string{""},
			},
		},
	}

	pager := ec2.NewDescribeVolumesPaginator(p.client, params)

	var upds unused.Disks

	for pager.HasMorePages() {
		res, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot list AWS unused disks: %w", err)
		}

		for _, v := range res.Volumes {
			m := unused.Meta{
				"zone": *v.AvailabilityZone,
			}
			for _, t := range v.Tags {
				k := *t.Key
				if k == "Name" || k == "CSIVolumeName" {
					// already returned in Name()
					continue
				}
				m[k] = *t.Value
			}

			upds = append(upds, &Disk{v, p, m})
		}
	}

	return upds, nil
}

// Delete deletes the given disk from AWS.
func (p *Provider) Delete(ctx context.Context, disk unused.Disk) error {
	_, err := p.client.DeleteVolume(ctx, &ec2.DeleteVolumeInput{
		VolumeId: aws.String(disk.ID()),
	})
	if err != nil {
		return fmt.Errorf("cannot delete AWS disk: %w", err)
	}
	return nil
}
