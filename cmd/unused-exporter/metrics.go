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

type metrics struct {
	logger *logfmt.Logger

	providers []unused.Provider

	info  *prometheus.Desc
	count *prometheus.Desc
	dur   *prometheus.Desc
	suc   *prometheus.Desc
}

func newMetrics(logger *logfmt.Logger, ps []unused.Provider) (*metrics, error) {
	labels := []string{"provider", "provider_id"}

	return &metrics{
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

func (c *metrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.info
	ch <- c.count
	ch <- c.dur
}

func (m *metrics) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	wg.Add(len(m.providers))

	for _, p := range m.providers {
		go func(p unused.Provider) {
			defer wg.Done()

			meta := p.Meta()

			lbs := logfmt.Labels{
				"provider": p.Name(),
				"metadata": meta,
			}
			m.logger.Log("collecting metrics", lbs)

			ctx := context.TODO()

			start := time.Now()
			c, err := m.collect(ctx, p)
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
				m.logger.Log("failed to collect metrics", lbs)
				success = 0
			}

			emit(m.info, 1)
			emit(m.dur, int(dur.Microseconds()))
			emit(m.count, c)
			emit(m.suc, success)
		}(p)
	}

	wg.Wait()
}

func (m *metrics) collect(ctx context.Context, p unused.Provider) (int, error) {
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
		m.logger.Log("unused disk found", lbls)
	}

	return len(disks), nil
}
