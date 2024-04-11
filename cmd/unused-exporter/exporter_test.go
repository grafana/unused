package main

import (
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/grafana/unused"
	"github.com/grafana/unused/aws"
	"github.com/grafana/unused/azure"
	"github.com/grafana/unused/gcp"
)

type MockProvider struct {
	unused.Provider

	name string
}

func (m MockProvider) Name() string { return m.name }

type MockDisk struct {
	unused.Disk
	name      string
	sizeGB    int
	createdAt time.Time
	meta      map[string]string
}

func (d *MockDisk) Name() string {
	return d.name
}

func (d *MockDisk) Meta() unused.Meta {
	return d.meta
}

func (d *MockDisk) CreatedAt() time.Time {
	return d.createdAt
}

func (d *MockDisk) SizeGB() int {
	return d.sizeGB
}

func TestGetRegionFromZone(t *testing.T) {
	type testCase struct {
		provider string
		zone     string
		expected string
	}

	testCases := map[string]testCase{
		"Azure": {azure.ProviderName, "eastus1", "eastus1"},
		"GCP":   {gcp.ProviderName, "us-central1-a", "us-central1"},
		"AWS":   {aws.ProviderName, "us-west-2a", "us-west-2"},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			p := &MockProvider{name: tc.provider}
			result := getRegionFromZone(p, tc.zone)
			if result != tc.expected {
				t.Errorf("getRegionFromZone(%s, %s) = %s, expected %s", tc.provider, tc.zone, result, tc.expected)
			}
		})
	}
}

func TestGetNamespace(t *testing.T) {
	type testCase struct {
		provider string
		diskMeta map[string]string
		expected string
	}

	testCases := map[string]testCase{
		"Azure": {
			provider: azure.ProviderName,
			diskMeta: map[string]string{
				"kubernetes.io-created-for-pvc-namespace": "azure-namespace",
			},
			expected: "azure-namespace",
		},
		"GCP": {
			provider: gcp.ProviderName,
			diskMeta: map[string]string{
				"kubernetes.io/created-for/pvc/namespace": "gcp-namespace",
			},
			expected: "gcp-namespace",
		},
		"AWS": {
			provider: aws.ProviderName,
			diskMeta: map[string]string{
				"kubernetes.io/created-for/pvc/namespace": "aws-namespace",
			},
			expected: "aws-namespace",
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			p := &MockProvider{name: tc.provider}
			d := &MockDisk{meta: tc.diskMeta}
			result := getNamespace(d, p)
			if result != tc.expected {
				t.Errorf("getNamespace(%v, %v) = %s, expected %s", d, p, result, tc.expected)
			}
		})
	}
}

func TestGetDiskLabels(t *testing.T) {
	type testCase struct {
		verbose  bool
		disk     *MockDisk
		expected []any
	}

	createdAt := time.Now()

	testCases := map[string]testCase{
		"Basic Disk Labels": {
			verbose: false,
			disk: &MockDisk{
				name:      "test-disk",
				sizeGB:    100,
				createdAt: createdAt,
				meta:      map[string]string{},
			},
			expected: []any{
				slog.String("name", "test-disk"),
				slog.Int("size_gb", 100),
				slog.Time("created", createdAt),
			},
		},
		"Verbose Disk Labels": {
			verbose: true,
			disk: &MockDisk{
				name:      "test-disk",
				sizeGB:    100,
				createdAt: createdAt,
				meta: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
			expected: []any{
				slog.String("name", "test-disk"),
				slog.Int("size_gb", 100),
				slog.Time("created", createdAt),
				slog.String("key1", "value1"),
				slog.String("key2", "value2"),
			},
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			actual := getDiskLabels(tc.disk, tc.verbose)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("getDiskLabels(%v, %v) = %v, expected %v", tc.disk, tc.verbose, actual, tc.expected)
			}
		})
	}
}
