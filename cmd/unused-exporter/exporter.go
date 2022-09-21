package main

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/grafana/unused"
	"github.com/inkel/logfmt"
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "unused"

type exporter struct {
	ctx    context.Context
	logger *logfmt.Logger

	timeout time.Duration

	providers []unused.Provider

	info  *prometheus.Desc
	count *prometheus.Desc
	dur   *prometheus.Desc
	suc   *prometheus.Desc
}

func newExporter(ctx context.Context, logger *logfmt.Logger, ps []unused.Provider, timeout time.Duration) (*exporter, error) {
	labels := []string{"provider", "provider_id"}

	return &exporter{
		ctx:       ctx,
		logger:    logger,
		providers: ps,
		timeout:   timeout,

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
	}, nil
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.info
	ch <- e.count
	ch <- e.dur
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(e.ctx, e.timeout)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(e.providers))

	for _, p := range e.providers {
		go func(p unused.Provider) {
			defer wg.Done()

			meta := p.Meta()

			lbs := logfmt.Labels{
				"provider": p.Name(),
				"metadata": meta,
			}
			e.logger.Log("collecting metrics", lbs)

			start := time.Now()

			countByNamespace := make(map[string]int)

			disks, err := p.ListUnusedDisks(ctx)
			for _, d := range disks {
				meta := d.Meta()
				lbls := logfmt.Labels{
					"provider": d.Provider().Name(),
					"name":     d.Name(),
					"created":  d.CreatedAt(),
				}
				for _, k := range meta.Keys() {
					lbls[k] = meta[k]
				}
				e.logger.Log("unused disk found", lbls)
				countByNamespace[meta["kubernetes.io/created-for/pvc/namespace"]] += 1
			}

			dur := time.Since(start)

			name := strings.ToLower(p.Name())
			var pid string
			switch name {
			case "gcp":
				pid = meta["project"]
			case "aws":
				pid = meta["profile"]
			case "azure":
				pid = meta["subscription"]
			default:
				pid = meta.String()
			}

			emit := func(d *prometheus.Desc, v int) {
				ch <- prometheus.MustNewConstMetric(d, prometheus.GaugeValue, float64(v), name, pid)
			}

			var success int = 1

			if err != nil {
				lbs["error"] = err
				e.logger.Log("failed to collect metrics", lbs)
				success = 0
			}

			emit(e.info, 1)
			emit(e.dur, int(dur.Microseconds()))
			emit(e.suc, success)

			for ns, c := range countByNamespace {
				ch <- prometheus.MustNewConstMetric(e.count, prometheus.GaugeValue, float64(c), name, pid, ns)
			}
		}(p)
	}

	wg.Wait()
}
