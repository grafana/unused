package aws_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	awsutil "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/grafana/unused-pds/pkg/provider/aws"
	"github.com/grafana/unused-pds/pkg/unused"
	"github.com/grafana/unused-pds/pkg/unused/unusedtest"
)

func TestNewProvider(t *testing.T) {
	ctx := context.Background()

	p, err := aws.NewProvider(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p == nil {
		t.Fatal("expecting Provider, got nil")
	}
}

func TestProviderMeta(t *testing.T) {
	unusedtest.TestProviderMeta(t, func(meta unused.Meta) (unused.Provider, error) {
		return aws.NewProvider(context.Background(), meta)
	})
}

func TestListUnusedDisks(t *testing.T) {
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// How cannot love you, AWS
		w.Write([]byte(`<DescribeVolumesResponse xmlns="http://ec2.amazonaws.com/doc/2016-11-15/">
   <requestId>59dbff89-35bd-4eac-99ed-be587EXAMPLE</requestId>
   <volumeSet>
      <item>
         <volumeId>vol-1234567890abcdef0</volumeId>
         <size>80</size>
         <snapshotId/>
         <availabilityZone>us-east-1a</availabilityZone>
         <status>available</status>
         <createTime>2022-03-12T17:25:21.000Z</createTime>
         <volumeType>standard</volumeType>
         <encrypted>true</encrypted>
         <multiAttachEnabled>false</multiAttachEnabled>
      </item>
      <item>
         <volumeId>vol-abcdef01234567890</volumeId>
         <size>120</size>
         <snapshotId/>
         <availabilityZone>us-east-1a</availabilityZone>
         <status>available</status>
         <createTime>2022-02-12T17:25:21.000Z</createTime>
         <volumeType>standard</volumeType>
         <encrypted>true</encrypted>
         <multiAttachEnabled>false</multiAttachEnabled>
      </item>
   </volumeSet>
</DescribeVolumesResponse>`))
	}))
	defer ts.Close()

	p, err := aws.NewProvider(ctx, nil,
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: awsutil.Credentials{
				AccessKeyID:     "AKID",
				SecretAccessKey: "SECRET",
				SessionToken:    "SESSION",
				Source:          "example hard coded credentials",
			},
		}),
		config.WithEndpointResolver(awsutil.EndpointResolverFunc(
			func(svc, reg string) (awsutil.Endpoint, error) {
				return awsutil.Endpoint{URL: ts.URL}, nil
			})))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	disks, err := p.ListUnusedDisks(ctx)
	if err != nil {
		t.Fatal("unexpected error listing unused disks:", err)
	}

	if exp, got := 2, len(disks); exp != got {
		t.Errorf("expecting %d disks, got %d", exp, got)
	}
}
