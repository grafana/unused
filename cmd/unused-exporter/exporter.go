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
	logger *logfmt.Logger

	providers []unused.Provider

	info  *prometheus.Desc
	count *prometheus.Desc
	dur   *prometheus.Desc
	suc   *prometheus.Desc
}

func newExporter(logger *logfmt.Logger, ps []unused.Provider) (*exporter, error) {
	labels := []string{"provider", "provider_id"}

	return &exporter{
		logger:    logger,
		providers: ps,

		info: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "provider", "info"),
			"CSP information",
			labels,
			nil),

		count: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disks", "count"),
			"How many unused disks are in this provider",
			labels,
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

			ctx := context.TODO()

			start := time.Now()
			c, err := e.collect(ctx, p)
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
			emit(e.count, c)
			emit(e.suc, success)
		}(p)
	}

	wg.Wait()
}

func (e *exporter) collect(ctx context.Context, p unused.Provider) (int, error) {
	disks, err := p.ListUnusedDisks(ctx)
	if err != nil {
		return 0, err
	}

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
	}

	return len(disks), nil
}
