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
}

func NewProvider(ctx context.Context) (unused.Provider, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot load AWS config: %w", err)
	}

	return &provider{
		client: ec2.NewFromConfig(cfg),
	}, nil
}

func (p *provider) ListUnusedDisks(ctx context.Context) (unused.Disks, error) {
	return ListUnusedDisks(ctx, p.client)
}

func ListUnusedDisks(ctx context.Context, c ec2.DescribeVolumesAPIClient) (unused.Disks, error) {
	params := &ec2.DescribeVolumesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("status"),
				Values: []string{string(types.VolumeStateAvailable)},
			},
		},
	}

	p := ec2.NewDescribeVolumesPaginator(c, params)

	var upds unused.Disks

	for p.HasMorePages() {
		res, err := p.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot list AWS unused disks: %w", err)
		}

		for _, v := range res.Volumes {
			upds = append(upds, &disk{v})
		}
	}

	return upds, nil
}
