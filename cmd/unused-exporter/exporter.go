package main

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/grafana/unused"
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "unused"

type metric struct {
	desc   *prometheus.Desc
	value  int64
	labels []string
}

type exporter struct {
	ctx    context.Context
	logger *slog.Logger

	timeout      time.Duration
	pollInterval time.Duration

	providers []unused.Provider

	info  *prometheus.Desc
	count *prometheus.Desc
	size  *prometheus.Desc
	dur   *prometheus.Desc
	suc   *prometheus.Desc

	mu    sync.RWMutex
	cache map[unused.Provider][]metric
}

func registerExporter(ctx context.Context, providers []unused.Provider, cfg config) error {
	labels := []string{"provider", "provider_id"}

	e := &exporter{
		ctx:          ctx,
		logger:       cfg.Logger,
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

		size: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disks", "size_gb"),
			"Total size of unused disks in this provider in GB",
			append(labels, "k8s_namespace", "type"),
			nil),

		dur: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "provider", "duration_ms"),
			"How long in milliseconds took to fetch this provider information",
			labels,
			nil),

		suc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "provider", "success"),
			"Static metric indicating if collecting the metrics succeeded or not",
			labels,
			nil),

		cache: make(map[unused.Provider][]metric, len(providers)),
	}

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

			ctx, cancel := context.WithTimeout(e.ctx, e.timeout)
			defer cancel()

			logger := e.logger.With(
				slog.String("provider", p.Name()),
				slog.String("provider_id", p.Id()),
			)

			logger.Info("collecting metrics")

			start := time.Now()
			disks, err := p.ListUnusedDisks(ctx)
			dur := time.Since(start)

			name := strings.ToLower(p.Name())

			var ms []metric // TODO we can optimize this creation here and allocate memory only once

			emit := func(d *prometheus.Desc, v int64, lbls ...string) {
				ms = append(ms, metric{
					desc:   d,
					value:  v,
					labels: append([]string{name, p.Id()}, lbls...),
				})
			}

			var success int64 = 1

			if err != nil {
				logger.Error("failed to collect metrics", slog.String("error", err.Error()))
				success = 0
			}

			emit(e.info, 1)
			emit(e.dur, int64(dur.Microseconds()))
			emit(e.suc, success)

			type namespaceInfo struct {
				Count      int
				SizeByType map[unused.DiskType]int64
			}

			diskInfoByNamespace := make(map[string]*namespaceInfo)
			for _, d := range disks {
				diskLabels := []any{
					slog.String("name", d.Name()),
					slog.Int("size_gb", d.SizeGB()),
					slog.Time("created", d.CreatedAt()),
				}
				logger.Info("unused disk found", diskLabels...)
				meta := d.Meta()
				ns := meta["kubernetes.io/created-for/pvc/namespace"]
				di := diskInfoByNamespace[ns]
				if di == nil {
					di = &namespaceInfo{
						SizeByType: make(map[unused.DiskType]int64),
					}
					diskInfoByNamespace[ns] = di
				}

				di.Count += 1
				di.SizeByType[d.DiskType()] += int64(d.SizeGB())
			}
			for ns, di := range diskInfoByNamespace {
				emit(e.count, int64(di.Count), ns)
				for diskType, diskSize := range di.SizeByType {
					emit(e.size, diskSize, ns, string(diskType))
				}
			}

			e.mu.Lock()
			e.cache[p] = ms
			e.mu.Unlock()

			<-tick.C
		}
	}
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for p, ms := range e.cache {
		e.logger.Info("reading provider cache", slog.String("provider", p.Name()), slog.String("provider_id", p.Id()), slog.Int("metrics", len(ms)))

		for _, m := range ms {
			ch <- prometheus.MustNewConstMetric(m.desc, prometheus.GaugeValue, float64(m.value), m.labels...)
		}
	}
}
