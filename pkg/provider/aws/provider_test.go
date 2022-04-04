package aws_test

import (
	"context"
	"testing"

	awsutil "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/grafana/unused-pds/pkg/provider/aws"
)

func TestNewProvider(t *testing.T) {
	ctx := context.Background()

	p, err := aws.NewProvider(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p == nil {
		t.Fatal("expecting Provider, got nil")
	}
}

func newVolume(name string) types.Volume {
	return types.Volume{
		Tags: []types.Tag{
			{
				Key:   awsutil.String("Name"),
				Value: awsutil.String(name),
			},
		},
		State: types.VolumeStateAvailable,
	}
}

type mockClient struct {
	idx   int
	disks []types.Volume
}

func newMockClient(disks ...types.Volume) *mockClient {
	return &mockClient{0, disks}
}

func (m *mockClient) DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
	if m.idx == len(m.disks) {
		return &ec2.DescribeVolumesOutput{}, nil
	}

	res := &ec2.DescribeVolumesOutput{
		NextToken: awsutil.String("next-page"),
		Volumes:   []types.Volume{m.disks[m.idx]},
	}

	m.idx++

	return res, nil
}

func TestListUnusedDisks(t *testing.T) {
	ctx := context.Background()

	c := newMockClient(
		newVolume("disk-1"),
		newVolume("disk-3"),
	)

	disks, err := aws.ListUnusedDisks(ctx, c)
	if err != nil {
		t.Fatal("unexpected error listing unused disks:", err)
	}

	if exp, got := 2, len(disks); exp != got {
		t.Errorf("expecting %d disks, got %d", exp, got)
	}
}
