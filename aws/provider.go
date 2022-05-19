package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/grafana/unused"
)

var _ unused.Provider = &Provider{}

type Provider struct {
	client *ec2.Client
	meta   unused.Meta
}

func (p *Provider) Name() string { return "AWS" }

func (p *Provider) Meta() unused.Meta { return p.meta }

func NewProvider(ctx context.Context, meta unused.Meta, optFns ...func(*config.LoadOptions) error) (*Provider, error) {
	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, fmt.Errorf("cannot load AWS config: %w", err)
	}

	if meta == nil {
		meta = make(unused.Meta)
	}

	return &Provider{
		client: ec2.NewFromConfig(cfg),
		meta:   meta,
	}, nil
}

func (p *Provider) ListUnusedDisks(ctx context.Context) (unused.Disks, error) {
	params := &ec2.DescribeVolumesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("status"),
				Values: []string{string(types.VolumeStateAvailable)},
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

func (p *Provider) Delete(ctx context.Context, disk unused.Disk) error {
	_, err := p.client.DeleteVolume(ctx, &ec2.DeleteVolumeInput{
		VolumeId: aws.String(disk.ID()),
	})
	if err != nil {
		return fmt.Errorf("cannot delete AWS disk: %w", err)
	}
	return nil
}
