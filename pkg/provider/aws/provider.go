package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/grafana/unused-pds/pkg/unused"
)

var _ unused.Provider = &provider{}

type provider struct {
	client *ec2.Client
	meta   unused.Meta
}

func (p *provider) Name() string { return "AWS" }

func (p *provider) Meta() unused.Meta { return p.meta }

func NewProvider(ctx context.Context, meta unused.Meta, optFns ...func(*config.LoadOptions) error) (unused.Provider, error) {
	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, fmt.Errorf("cannot load AWS config: %w", err)
	}

	if meta == nil {
		meta = make(unused.Meta)
	}

	return &provider{
		client: ec2.NewFromConfig(cfg),
		meta:   meta,
	}, nil
}

func (p *provider) ListUnusedDisks(ctx context.Context) (unused.Disks, error) {
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

			upds = append(upds, &disk{v, p, m})
		}
	}

	return upds, nil
}
