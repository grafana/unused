package aws_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	awsutil "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	endpoints "github.com/aws/smithy-go/endpoints"
	"github.com/grafana/unused"
	"github.com/grafana/unused/aws"
	"github.com/grafana/unused/unusedtest"
)

func TestNewProvider(t *testing.T) {
	cfg := awsutil.NewConfig()

	p, err := aws.NewProvider(nil, ec2.NewFromConfig(*cfg), map[string]string{"profile": "my-profile"})
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
		cfg := awsutil.NewConfig()

		return aws.NewProvider(nil, ec2.NewFromConfig(*cfg), meta)
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

func mockClient(t *testing.T, h http.HandlerFunc) (*ec2.Client, func()) {
	t.Helper()

	ts := httptest.NewServer(h)

	cfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("AKID", "SECRET", "SESSION")),
	)
	if err != nil {
		t.Fatalf("cannot load AWS config: %v", err)
	}

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("cannot parse test server URL: %v", err)
	}

	return ec2.NewFromConfig(cfg, func(o *ec2.Options) {
		o.BaseEndpoint = &u.Host
		o.EndpointResolverV2 = mockEndpointResolver(*u)
	}), ts.Close
}

func TestListUnusedDisks(t *testing.T) {
	client, cancel := mockClient(t, func(w http.ResponseWriter, req *http.Request) {
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
	})
	t.Cleanup(cancel)

	p, err := aws.NewProvider(nil, client, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	disks, err := p.ListUnusedDisks(t.Context())
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

func TestProviderDelete(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		client, cancel := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<DeleteVolumeResponse xmlns="http://ec2.amazonaws.com/doc/2016-11-15/">
    <requestId>test-request-id</requestId>
    <return>true</return>
</DeleteVolumeResponse>`))
		})
		t.Cleanup(cancel)

		provider, err := aws.NewProvider(nil, client, map[string]string{"profile": "test-profile"})
		if err != nil {
			t.Fatalf("cannot create provider: %v", err)
		}

		disk := &aws.Disk{
			Volume: types.Volume{
				VolumeId: awsutil.String("vol-12345"),
			},
		}

		err = provider.Delete(t.Context(), disk)
		if err != nil {
			t.Errorf("Delete() unexpected error: %v", err)
		}
	})

	t.Run("deletion error", func(t *testing.T) {
		client, cancel := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <Errors>
        <Error>
            <Code>InvalidVolume.NotFound</Code>
            <Message>The volume does not exist.</Message>
        </Error>
    </Errors>
</Response>`))
		})
		t.Cleanup(cancel)

		provider, err := aws.NewProvider(nil, client, map[string]string{"profile": "test-profile"})
		if err != nil {
			t.Fatalf("cannot create provider: %v", err)
		}

		disk := &aws.Disk{
			Volume: types.Volume{
				VolumeId: awsutil.String("vol-invalid"),
			},
		}

		err = provider.Delete(t.Context(), disk)
		if err == nil {
			t.Error("Delete() expected error, got nil")
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		client, cancel := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		})
		t.Cleanup(cancel)

		ctx, cancel := context.WithCancel(t.Context())
		cancel() // Cancel immediately

		provider, err := aws.NewProvider(nil, client, map[string]string{"profile": "test-profile"})
		if err != nil {
			t.Fatalf("cannot create provider: %v", err)
		}

		disk := &aws.Disk{
			Volume: types.Volume{
				VolumeId: awsutil.String("vol-12345"),
			},
		}

		err = provider.Delete(ctx, disk)
		if err == nil {
			t.Error("Delete() expected context cancellation error, got nil")
		}
	})
}
