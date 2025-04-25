package aws_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	awsutil "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	endpoints "github.com/aws/smithy-go/endpoints"
	"github.com/grafana/unused"
	"github.com/grafana/unused/aws"
	"github.com/grafana/unused/unusedtest"
)

func TestNewProvider(t *testing.T) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatalf("cannot load AWS config: %v", err)
	}

	p, err := aws.NewProvider(nil, ec2.NewFromConfig(cfg), map[string]string{"profile": "my-profile"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p == nil {
		t.Fatal("expecting Provider, got nil")
	}

	if exp, got := "my-profile", p.ID(); exp != got {
		t.Fatalf("provider id was incorrect, exp: %v, got: %v", exp, got)
	}
}

func TestProviderMeta(t *testing.T) {
	err := unusedtest.TestProviderMeta(func(meta unused.Meta) (unused.Provider, error) {
		ctx := context.Background()
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			t.Fatalf("cannot load AWS config: %v", err)
		}

		return aws.NewProvider(nil, ec2.NewFromConfig(cfg), meta)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type mockEndpointResolver url.URL

func (er mockEndpointResolver) ResolveEndpoint(ctx context.Context, params ec2.EndpointParameters) (endpoints.Endpoint, error) {
	return endpoints.Endpoint{
		URI: url.URL(er),
	}, nil
}

func TestListUnusedDisks(t *testing.T) {
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// How cannot love you, AWS
		_, err := w.Write([]byte(`<DescribeVolumesResponse xmlns="http://ec2.amazonaws.com/doc/2016-11-15/">
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
         <availabilityZone>us-west-2b</availabilityZone>
         <status>available</status>
         <createTime>2022-02-12T17:25:21.000Z</createTime>
         <volumeType>standard</volumeType>
         <encrypted>true</encrypted>
         <multiAttachEnabled>false</multiAttachEnabled>
         <tagSet>
            <item>
               <key>CSIVolumeName</key>
               <value>prometheus-1</value>
            </item>
            <item>
               <key>ebs.csi.aws.com/cluster</key>
               <value>true</value>
            </item>
            <item>
               <key>kubernetes.io-created-for-pv-name</key>
               <value>pvc-prometheus-1</value>
            </item>
            <item>
               <key>kubernetes.io-created-for-pvc-name</key>
               <value>prometheus-1</value>
            </item>
            <item>
               <key>kubernetes.io-created-for-pvc-namespace</key>
               <value>monitoring</value>
            </item>
         </tagSet>
      </item>
   </volumeSet>
</DescribeVolumesResponse>`))
		if err != nil {
			t.Fatalf("unexpected error writing response: %v", err)
		}
	}))
	defer ts.Close()

	tsURL, _ := url.Parse(ts.URL)
	er := mockEndpointResolver(*tsURL)

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: awsutil.Credentials{
				AccessKeyID:     "AKID",
				SecretAccessKey: "SECRET",
				SessionToken:    "SESSION",
				Source:          "example hard coded credentials",
			},
		}))
	if err != nil {
		t.Fatalf("cannot load AWS config: %v", err)
	}

	p, err := aws.NewProvider(nil, ec2.NewFromConfig(cfg, ec2.WithEndpointResolverV2(ec2.EndpointResolverV2(er))), nil)
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

	err = unusedtest.AssertEqualMeta(unused.Meta{"zone": "us-east-1a"}, disks[0].Meta())
	if err != nil {
		t.Fatalf("metadata doesn't match: %v", err)
	}

	err = unusedtest.AssertEqualMeta(unused.Meta{
		"zone":                                    "us-west-2b",
		"ebs.csi.aws.com/cluster":                 "true",
		"kubernetes.io-created-for-pv-name":       "pvc-prometheus-1",
		"kubernetes.io-created-for-pvc-name":      "prometheus-1",
		"kubernetes.io-created-for-pvc-namespace": "monitoring",
	}, disks[1].Meta())
	if err != nil {
		t.Fatalf("metadata doesn't match: %v", err)
	}
}
