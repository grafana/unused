package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/grafana/unused"
	"github.com/grafana/unused/aws"
	"github.com/grafana/unused/azure"
	"github.com/grafana/unused/gcp"
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "unused"

type metric struct {
	desc   *prometheus.Desc
	value  float64
	labels []string
}

type exporter struct {
	ctx     context.Context
	logger  *slog.Logger
	verbose bool

	timeout      time.Duration
	pollInterval time.Duration

	providers []unused.Provider

	info  *prometheus.Desc
	count *prometheus.Desc
	ds    *prometheus.Desc
	size  *prometheus.Desc
	dur   *prometheus.Desc
	suc   *prometheus.Desc
	dlu   *prometheus.Desc

	mu    sync.RWMutex
	cache map[unused.Provider][]metric
}

func registerExporter(ctx context.Context, providers []unused.Provider, cfg config) error {
	labels := []string{"provider", "provider_id"}

	e := &exporter{
		ctx:          ctx,
		logger:       cfg.Logger,
		verbose:      cfg.VerboseLogging,
		providers:    providers,
		timeout:      cfg.Collector.Timeout,
		pollInterval: cfg.Collector.PollInterval,

		info: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "provider", "info"),
			"CSP information",
			labels,
			nil),

		count: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disks", "count"),
			"How many unused disks are in this provider",
			append(labels, "k8s_namespace"),
			nil),
		ds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disk", "size_bytes"),
			"Disk size in bytes",
			append(labels, []string{"disk", "created_for_pv", "k8s_namespace", "type", "region", "zone"}...),
			nil),

		size: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disks", "total_size_bytes"),
			"Total size of unused disks in this provider in bytes",
			append(labels, "k8s_namespace", "type"),
			nil),

		dur: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "provider", "duration_seconds"),
			"How long in seconds took to fetch this provider information",
			labels,
			nil),

		suc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "provider", "success"),
			"Static metric indicating if collecting the metrics succeeded or not",
			labels,
			nil),

		dlu: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disks", "last_used_timestamp_seconds"),
			"Kubernetes metadata associated with each unused disk, with the value as the last time the disk was used (if available)",
			append(labels, []string{"disk", "created_for_pv", "created_for_pvc", "zone"}...),
			nil),

		cache: make(map[unused.Provider][]metric, len(providers)),
	}

	e.logger.Info("start background polling of providers",
		slog.Int("providers", len(e.providers)),
		slog.Duration("interval", e.pollInterval),
		slog.Duration("timeout", e.timeout),
	)

	for _, p := range providers {
		p := p
		go e.pollProvider(p)
	}

	return prometheus.Register(e)
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.info
	ch <- e.count
	ch <- e.size
	ch <- e.dur
	ch <- e.dlu
}

type namespaceInfo struct {
	Count      int
	SizeByType map[unused.DiskType]float64
}

func (e *exporter) pollProvider(p unused.Provider) {
	tick := time.NewTicker(e.pollInterval)
	defer tick.Stop()

	for {
		select {
		case <-e.ctx.Done(): // parent context was cancelled
			return

		default:
			// we don't wait for tick.C here as we want to start
			// polling immediately; we wait at the end.

			var (
				success int64 = 1

				logger = e.logger.With(
					slog.String("provider", strings.ToLower(p.Name())),
					slog.String("provider_id", p.ID()),
				)
			)

			logger.Info("collecting metrics")
			ctx, cancel := context.WithTimeout(e.ctx, e.timeout)
			start := time.Now()
			disks, err := p.ListUnusedDisks(ctx)
			cancel() // release resources early
			dur := time.Since(start)
			if err != nil {
				logger.Error("failed to collect metrics", slog.String("error", err.Error()))
				success = 0
			}

			diskInfoByNamespace := make(map[string]*namespaceInfo)
			var ms []metric

			for _, d := range disks {
				diskLabels := getDiskLabels(d, e.verbose)
				e.logger.Info("unused disk found", diskLabels...)

				ns := getNamespace(d, p)
				di := diskInfoByNamespace[ns]
				if di == nil {
					di = &namespaceInfo{
						SizeByType: make(map[unused.DiskType]float64),
					}
					diskInfoByNamespace[ns] = di
				}
				di.Count += 1
				di.SizeByType[d.DiskType()] += float64(d.SizeBytes())

				e.logger.Info(fmt.Sprintf("Disk %s last used at %v", d.Name(), d.LastUsedAt()))

				m := d.Meta()
				if m.CreatedForPV() == "" {
					continue
				}

				addMetric(&ms, p, e.dlu, lastUsedTS(d), d.ID(), m.CreatedForPV(), m.CreatedForPVC(), m.Zone())
				addMetric(&ms, p, e.ds, d.SizeBytes(), d.ID(), m.CreatedForPV(), ns, string(d.DiskType()), getRegionFromZone(p, m.Zone()), m.Zone())
			}

			addMetric(&ms, p, e.info, 1)
			addMetric(&ms, p, e.dur, float64(dur.Seconds()))
			addMetric(&ms, p, e.suc, float64(success))

			for ns, di := range diskInfoByNamespace {
				addMetric(&ms, p, e.count, float64(di.Count), ns)
				for diskType, diskSize := range di.SizeByType {
					addMetric(&ms, p, e.size, diskSize, ns, string(diskType))
				}
			}

			e.mu.Lock()
			e.cache[p] = ms
			e.mu.Unlock()

			logger.Info("metrics collected",
				slog.Int("metrics", len(ms)),
				slog.Bool("success", success == 1),
				slog.Duration("dur", dur),
			)

			<-tick.C
		}

	}
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for p, ms := range e.cache {
		labels := []any{
			slog.String("provider", p.Name()),
			slog.String("provider_id", p.ID()),
			slog.Int("metrics", len(ms)),
		}

		if e.verbose {
			providerMeta := p.Meta()
			providerMetaLabels := make([]any, 0, len(providerMeta))
			for _, k := range providerMeta.Keys() {
				providerMetaLabels = append(providerMetaLabels, slog.String(k, providerMeta[k]))
			}
			labels = append(labels, providerMetaLabels...)
		}

		e.logger.Info("reading provider cache", labels...)

		for _, m := range ms {
			ch <- prometheus.MustNewConstMetric(m.desc, prometheus.GaugeValue, m.value, m.labels...)
		}
	}
}

func getDiskLabels(d unused.Disk, v bool) []any {
	diskLabels := []any{
		slog.String("name", d.Name()),
		slog.Int("size_gb", d.SizeGB()),
		slog.Time("created", d.CreatedAt()),
	}

	if v {
		meta := d.Meta()
		diskMetaLabels := make([]any, 0, len(meta))
		for _, k := range meta.Keys() {
			diskMetaLabels = append(diskMetaLabels, slog.String(k, meta[k]))
		}
		diskLabels = append(diskLabels, diskMetaLabels...)
	}

	return diskLabels
}

func getNamespace(d unused.Disk, p unused.Provider) string {
	switch p.Name() {
	case gcp.ProviderName:
		return d.Meta()["kubernetes.io/created-for/pvc/namespace"]
	case aws.ProviderName:
		return d.Meta()["kubernetes.io/created-for/pvc/namespace"]
	case azure.ProviderName:
		return d.Meta()["kubernetes.io-created-for-pvc-namespace"]
	default:
		panic("getNamespace(): unrecognized provider name:" + p.Name())
	}
}

func addMetric(ms *[]metric, p unused.Provider, d *prometheus.Desc, v float64, lbls ...string) {
	*ms = append(*ms, metric{
		desc:   d,
		value:  v,
		labels: append([]string{strings.ToLower(p.Name()), p.ID()}, lbls...),
	})
}

func lastUsedTS(d unused.Disk) float64 {
	lastUsed := d.LastUsedAt()
	if lastUsed.IsZero() {
		return 0
	}

	return float64(lastUsed.Unix())
}

func getRegionFromZone(p unused.Provider, z string) string {
	switch p.Name() {
	case gcp.ProviderName:
		return z[:strings.LastIndex(z, "-")]
	case aws.ProviderName:
		return z[:len(z)-1]
	case azure.ProviderName:
		return z
	default:
		panic("getRegionFromZone(): unrecognized provider name:" + p.Name())
	}
}
