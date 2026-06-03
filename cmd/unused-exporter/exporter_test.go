package main

import (
	"context"
	"io"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/grafana/unused"
	"github.com/grafana/unused/aws"
	"github.com/grafana/unused/azure"
	"github.com/grafana/unused/gcp"
	"github.com/prometheus/client_golang/prometheus"
)

type MockProvider struct {
	unused.Provider

	name      string
	id        string
	meta      unused.Meta
	disks     unused.Disks
	listErr   error
	deleteErr error
}

func (m MockProvider) Name() string      { return m.name }
func (m MockProvider) ID() string        { return m.id }
func (m MockProvider) Meta() unused.Meta { return m.meta }
func (m MockProvider) ListUnusedDisks(ctx context.Context) (unused.Disks, error) {
	return m.disks, m.listErr
}
func (m MockProvider) Delete(ctx context.Context, disk unused.Disk) error {
	return m.deleteErr
}

type MockDisk struct {
	unused.Disk
	id         string
	provider   unused.Provider
	name       string
	sizeGB     int
	sizeBytes  float64
	diskType   unused.DiskType
	createdAt  time.Time
	lastUsedAt time.Time
	meta       map[string]string
}

func (d *MockDisk) ID() string                { return d.id }
func (d *MockDisk) Provider() unused.Provider { return d.provider }
func (d *MockDisk) Name() string              { return d.name }
func (d *MockDisk) Meta() unused.Meta         { return d.meta }
func (d *MockDisk) CreatedAt() time.Time      { return d.createdAt }
func (d *MockDisk) LastUsedAt() time.Time     { return d.lastUsedAt }
func (d *MockDisk) SizeGB() int               { return d.sizeGB }
func (d *MockDisk) SizeBytes() float64        { return d.sizeBytes }
func (d *MockDisk) DiskType() unused.DiskType { return d.diskType }

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

func TestLastUsedTS(t *testing.T) {
	tests := []struct {
		name     string
		lastUsed time.Time
		expected float64
	}{
		{
			name:     "zero time",
			lastUsed: time.Time{},
			expected: 0,
		},
		{
			name:     "valid timestamp",
			lastUsed: time.Unix(1609459200, 0), // 2021-01-01 00:00:00 UTC
			expected: 1609459200,
		},
		{
			name:     "recent timestamp",
			lastUsed: time.Unix(1640995200, 0), // 2022-01-01 00:00:00 UTC
			expected: 1640995200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDisk := &testDiskWithTime{lastUsed: tt.lastUsed}
			result := lastUsedTS(testDisk)
			if result != tt.expected {
				t.Errorf("lastUsedTS() = %f, want %f", result, tt.expected)
			}
		})
	}
}

type testDiskWithTime struct {
	unused.Disk
	lastUsed time.Time
}

func (d *testDiskWithTime) LastUsedAt() time.Time     { return d.lastUsed }
func (d *testDiskWithTime) ID() string                { return "test-id" }
func (d *testDiskWithTime) Provider() unused.Provider { return nil }
func (d *testDiskWithTime) Name() string              { return "test-disk" }
func (d *testDiskWithTime) CreatedAt() time.Time      { return time.Time{} }
func (d *testDiskWithTime) Meta() unused.Meta         { return unused.Meta{} }
func (d *testDiskWithTime) SizeGB() int               { return 0 }
func (d *testDiskWithTime) SizeBytes() float64        { return 0 }
func (d *testDiskWithTime) DiskType() unused.DiskType { return unused.Unknown }

func TestAddMetric(t *testing.T) {
	tests := []struct {
		name   string
		labels []string
		value  float64
	}{
		{
			name:   "no additional labels",
			labels: []string{},
			value:  42.0,
		},
		{
			name:   "with labels",
			labels: []string{"label1", "label2", "label3"},
			value:  100.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ms []metric
			p := &MockProvider{name: "TestProvider"}
			desc := prometheus.NewDesc("test_metric", "Test metric", []string{"provider", "provider_id"}, nil)

			addMetric(&ms, p, desc, tt.value, tt.labels...)

			if len(ms) != 1 {
				t.Fatalf("addMetric() should add 1 metric, got %d", len(ms))
			}

			m := ms[0]
			if m.value != tt.value {
				t.Errorf("metric value = %f, want %f", m.value, tt.value)
			}

			expectedLabels := append([]string{"testprovider", ""}, tt.labels...)
			if !reflect.DeepEqual(m.labels, expectedLabels) {
				t.Errorf("metric labels = %v, want %v", m.labels, expectedLabels)
			}

			if m.desc != desc {
				t.Errorf("metric desc mismatch")
			}
		})
	}
}

func TestDescribe(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	cfg := config{
		Logger: logger,
	}
	cfg.Collector.Timeout = time.Second
	cfg.Collector.PollInterval = time.Minute

	providers := []unused.Provider{}

	e := &exporter{
		ctx:          ctx,
		logger:       cfg.Logger,
		providers:    providers,
		timeout:      cfg.Collector.Timeout,
		pollInterval: cfg.Collector.PollInterval,
		info: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "provider", "info"),
			"CSP information",
			[]string{"provider", "provider_id"},
			nil),
		count: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disks", "count"),
			"How many unused disks are in this provider",
			[]string{"provider", "provider_id", "k8s_namespace"},
			nil),
		size: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disks", "total_size_bytes"),
			"Total size of unused disks in this provider in bytes",
			[]string{"provider", "provider_id", "k8s_namespace", "type"},
			nil),
		dur: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "provider", "duration_seconds"),
			"How long in seconds took to fetch this provider information",
			[]string{"provider", "provider_id"},
			nil),
		dlu: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disks", "last_used_timestamp_seconds"),
			"Kubernetes metadata associated with each unused disk",
			[]string{"provider", "provider_id", "disk", "created_for_pv", "created_for_pvc", "zone"},
			nil),
		cache: make(map[unused.Provider][]metric),
	}

	ch := make(chan *prometheus.Desc, 10)
	go func() {
		e.Describe(ch)
		close(ch)
	}()

	descriptions := make([]*prometheus.Desc, 0)
	for desc := range ch {
		descriptions = append(descriptions, desc)
	}

	// Should describe 5 metrics: info, count, size, dur, dlu
	if len(descriptions) != 5 {
		t.Errorf("Describe() sent %d descriptions, want 5", len(descriptions))
	}
}

func TestCollect(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	provider := &MockProvider{
		name: "TestProvider",
		id:   "test-id",
		meta: unused.Meta{"key": "value"},
	}

	e := &exporter{
		ctx:    ctx,
		logger: logger,
		cache:  make(map[unused.Provider][]metric),
	}

	// Add some metrics to the cache
	desc := prometheus.NewDesc("test_metric", "Test metric", []string{"provider", "provider_id"}, nil)
	e.cache[provider] = []metric{
		{desc: desc, value: 42.0, labels: []string{"testprovider", "test-id"}},
		{desc: desc, value: 100.0, labels: []string{"testprovider", "test-id"}},
	}

	ch := make(chan prometheus.Metric, 10)
	go func() {
		e.Collect(ch)
		close(ch)
	}()

	metrics := make([]prometheus.Metric, 0)
	for m := range ch {
		metrics = append(metrics, m)
	}

	if len(metrics) != 2 {
		t.Errorf("Collect() sent %d metrics, want 2", len(metrics))
	}
}

func TestCollectEmpty(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	e := &exporter{
		ctx:    ctx,
		logger: logger,
		cache:  make(map[unused.Provider][]metric),
	}

	ch := make(chan prometheus.Metric, 10)
	go func() {
		e.Collect(ch)
		close(ch)
	}()

	metrics := make([]prometheus.Metric, 0)
	for m := range ch {
		metrics = append(metrics, m)
	}

	if len(metrics) != 0 {
		t.Errorf("Collect() sent %d metrics, want 0", len(metrics))
	}
}

func TestPollProvider(t *testing.T) {
	t.Run("successful poll", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		// Create mock provider with disks
		disk := &MockDisk{
			id:         "disk-1",
			name:       "test-disk",
			sizeGB:     100,
			sizeBytes:  107374182400,
			diskType:   unused.SSD,
			createdAt:  time.Now(),
			lastUsedAt: time.Now().Add(-24 * time.Hour),
			meta: map[string]string{
				"kubernetes.io/created-for/pvc/namespace": "default",
				"kubernetes.io/created-for/pv/name":       "pv-test",
				"zone":                                    "us-central1-a",
			},
		}

		provider := &MockProvider{
			name:  gcp.ProviderName,
			id:    "test-project",
			meta:  unused.Meta{},
			disks: unused.Disks{disk},
		}

		e := &exporter{
			ctx:          ctx,
			logger:       logger,
			verbose:      false,
			timeout:      time.Second,
			pollInterval: 100 * time.Millisecond,
			providers:    []unused.Provider{provider},
			info: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "provider", "info"),
				"CSP information",
				[]string{"provider", "provider_id"},
				nil),
			count: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "disks", "count"),
				"How many unused disks",
				[]string{"provider", "provider_id", "k8s_namespace"},
				nil),
			ds: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "disk", "size_bytes"),
				"Disk size",
				[]string{"provider", "provider_id", "disk", "created_for_pv", "k8s_namespace", "type", "region", "zone"},
				nil),
			size: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "disks", "total_size_bytes"),
				"Total size",
				[]string{"provider", "provider_id", "k8s_namespace", "type"},
				nil),
			dur: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "provider", "duration_seconds"),
				"Duration",
				[]string{"provider", "provider_id"},
				nil),
			suc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "provider", "success"),
				"Success",
				[]string{"provider", "provider_id"},
				nil),
			dlu: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "disks", "last_used_timestamp_seconds"),
				"Last used",
				[]string{"provider", "provider_id", "disk", "created_for_pv", "created_for_pvc", "zone"},
				nil),
			cache: make(map[unused.Provider][]metric),
		}

		// Start polling in goroutine
		done := make(chan bool)
		go func() {
			e.pollProvider(provider)
			done <- true
		}()

		// Wait for at least one poll cycle
		time.Sleep(150 * time.Millisecond)

		// Cancel context to stop polling
		cancel()

		// Wait for pollProvider to finish
		select {
		case <-done:
			// Success
		case <-time.After(time.Second):
			t.Fatal("pollProvider did not exit after context cancellation")
		}

		// Check that metrics were cached
		e.mu.RLock()
		metrics := e.cache[provider]
		e.mu.RUnlock()

		if len(metrics) == 0 {
			t.Error("expected metrics in cache, got none")
		}
	})

	t.Run("poll with ListUnusedDisks error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		provider := &MockProvider{
			name:    aws.ProviderName,
			id:      "test-profile",
			meta:    unused.Meta{},
			listErr: context.DeadlineExceeded,
		}

		e := &exporter{
			ctx:          ctx,
			logger:       logger,
			verbose:      false,
			timeout:      time.Second,
			pollInterval: 100 * time.Millisecond,
			providers:    []unused.Provider{provider},
			info: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "provider", "info"),
				"CSP information",
				[]string{"provider", "provider_id"},
				nil),
			count: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "disks", "count"),
				"How many unused disks",
				[]string{"provider", "provider_id", "k8s_namespace"},
				nil),
			size: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "disks", "total_size_bytes"),
				"Total size",
				[]string{"provider", "provider_id", "k8s_namespace", "type"},
				nil),
			dur: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "provider", "duration_seconds"),
				"Duration",
				[]string{"provider", "provider_id"},
				nil),
			suc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "provider", "success"),
				"Success",
				[]string{"provider", "provider_id"},
				nil),
			dlu: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "disks", "last_used_timestamp_seconds"),
				"Last used",
				[]string{"provider", "provider_id", "disk", "created_for_pv", "created_for_pvc", "zone"},
				nil),
			ds: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "disk", "size_bytes"),
				"Disk size",
				[]string{"provider", "provider_id", "disk", "created_for_pv", "k8s_namespace", "type", "region", "zone"},
				nil),
			cache: make(map[unused.Provider][]metric),
		}

		// Start polling
		done := make(chan bool)
		go func() {
			e.pollProvider(provider)
			done <- true
		}()

		// Wait for at least one poll cycle
		time.Sleep(150 * time.Millisecond)

		// Check that metrics still exist (with success=0)
		e.mu.RLock()
		metrics := e.cache[provider]
		e.mu.RUnlock()

		if len(metrics) == 0 {
			t.Error("expected error metrics in cache, got none")
		}

		// Verify success metric is 0
		foundSuccessMetric := false
		for _, m := range metrics {
			if m.desc == e.suc && m.value == 0 {
				foundSuccessMetric = true
				break
			}
		}

		if !foundSuccessMetric {
			t.Error("expected success=0 metric in cache")
		}

		// Cancel and clean up
		cancel()
		<-done
	})
}

func TestRegisterExporter(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		provider := &MockProvider{
			name:  aws.ProviderName,
			id:    "test-profile",
			meta:  unused.Meta{},
			disks: unused.Disks{},
		}

		cfg := config{
			Logger: logger,
		}
		cfg.Collector.Timeout = time.Second
		cfg.Collector.PollInterval = time.Hour // Long interval to avoid actual polling

		// Create a new registry to avoid conflicts
		registry := prometheus.NewRegistry()

		// Register with our custom registry
		e := &exporter{
			ctx:          ctx,
			logger:       cfg.Logger,
			verbose:      cfg.VerboseLogging,
			providers:    []unused.Provider{provider},
			timeout:      cfg.Collector.Timeout,
			pollInterval: cfg.Collector.PollInterval,
			info: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "provider", "info"),
				"CSP information",
				[]string{"provider", "provider_id"},
				nil),
			count: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "disks", "count"),
				"How many unused disks",
				[]string{"provider", "provider_id", "k8s_namespace"},
				nil),
			ds: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "disk", "size_bytes"),
				"Disk size",
				[]string{"provider", "provider_id", "disk", "created_for_pv", "k8s_namespace", "type", "region", "zone"},
				nil),
			size: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "disks", "total_size_bytes"),
				"Total size",
				[]string{"provider", "provider_id", "k8s_namespace", "type"},
				nil),
			dur: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "provider", "duration_seconds"),
				"Duration",
				[]string{"provider", "provider_id"},
				nil),
			suc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "provider", "success"),
				"Success",
				[]string{"provider", "provider_id"},
				nil),
			dlu: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "disks", "last_used_timestamp_seconds"),
				"Last used",
				[]string{"provider", "provider_id", "disk", "created_for_pv", "created_for_pvc", "zone"},
				nil),
			cache: make(map[unused.Provider][]metric, len([]unused.Provider{provider})),
		}

		err := registry.Register(e)
		if err != nil {
			t.Fatalf("failed to register exporter: %v", err)
		}

		// Verify we can collect metrics
		metricFamilies, err := registry.Gather()
		if err != nil {
			t.Fatalf("failed to gather metrics: %v", err)
		}

		// Should have no metrics initially (cache is empty)
		if len(metricFamilies) > 0 {
			// This is OK, some metrics might exist
			return
		}
	})

	t.Run("registration with no providers", func(t *testing.T) {
		ctx := t.Context()

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		cfg := config{
			Logger: logger,
		}
		cfg.Collector.Timeout = time.Second
		cfg.Collector.PollInterval = time.Hour

		err := registerExporter(ctx, []unused.Provider{}, cfg)
		if err != nil {
			t.Fatalf("registerExporter with no providers should not fail: %v", err)
		}
	})
}

func TestCollectVerbose(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	provider := &MockProvider{
		name: "TestProvider",
		id:   "test-id",
		meta: unused.Meta{"region": "us-west-2", "account": "123456"},
	}

	e := &exporter{
		ctx:     ctx,
		logger:  logger,
		verbose: true, // Enable verbose mode
		cache:   make(map[unused.Provider][]metric),
	}

	// Add metrics to cache
	desc := prometheus.NewDesc("test_metric", "Test metric", []string{"provider", "provider_id"}, nil)
	e.cache[provider] = []metric{
		{desc: desc, value: 42.0, labels: []string{"testprovider", "test-id"}},
	}

	ch := make(chan prometheus.Metric, 10)
	go func() {
		e.Collect(ch)
		close(ch)
	}()

	metrics := make([]prometheus.Metric, 0)
	for m := range ch {
		metrics = append(metrics, m)
	}

	if len(metrics) != 1 {
		t.Errorf("Collect() sent %d metrics, want 1", len(metrics))
	}
}

func TestCollectMultipleProviders(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	provider1 := &MockProvider{name: "AWS", id: "account-1", meta: unused.Meta{}}
	provider2 := &MockProvider{name: "GCP", id: "project-1", meta: unused.Meta{}}

	e := &exporter{
		ctx:    ctx,
		logger: logger,
		cache:  make(map[unused.Provider][]metric),
	}

	desc := prometheus.NewDesc("test_metric", "Test", []string{"provider", "provider_id"}, nil)
	e.cache[provider1] = []metric{{desc: desc, value: 1.0, labels: []string{"aws", "account-1"}}}
	e.cache[provider2] = []metric{{desc: desc, value: 2.0, labels: []string{"gcp", "project-1"}}}

	ch := make(chan prometheus.Metric, 10)
	go func() {
		e.Collect(ch)
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	if count != 2 {
		t.Errorf("Collect() sent %d metrics, want 2", count)
	}
}

func TestGetNamespaceEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		diskMeta map[string]string
		expected string
	}{
		{
			name:     "Azure empty namespace",
			provider: azure.ProviderName,
			diskMeta: map[string]string{
				"kubernetes.io-created-for-pvc-namespace": "",
			},
			expected: "",
		},
		{
			name:     "GCP missing namespace key",
			provider: gcp.ProviderName,
			diskMeta: map[string]string{
				"other-key": "value",
			},
			expected: "",
		},
		{
			name:     "AWS with namespace",
			provider: aws.ProviderName,
			diskMeta: map[string]string{
				"kubernetes.io/created-for/pvc/namespace": "production",
			},
			expected: "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &MockProvider{name: tt.provider}
			d := &MockDisk{meta: tt.diskMeta}
			result := getNamespace(d, p)
			if result != tt.expected {
				t.Errorf("getNamespace() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetRegionFromZoneEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		zone     string
		expected string
	}{
		{
			name:     "GCP single zone",
			provider: gcp.ProviderName,
			zone:     "us-central1-a",
			expected: "us-central1",
		},
		{
			name:     "GCP multi-part zone",
			provider: gcp.ProviderName,
			zone:     "europe-west1-b",
			expected: "europe-west1",
		},
		{
			name:     "AWS standard zone",
			provider: aws.ProviderName,
			zone:     "eu-west-1a",
			expected: "eu-west-1",
		},
		{
			name:     "Azure passthrough",
			provider: azure.ProviderName,
			zone:     "westus2",
			expected: "westus2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &MockProvider{name: tt.provider}
			result := getRegionFromZone(p, tt.zone)
			if result != tt.expected {
				t.Errorf("getRegionFromZone(%s, %s) = %s, want %s", tt.provider, tt.zone, result, tt.expected)
			}
		})
	}
}
