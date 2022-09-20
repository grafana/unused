package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/grafana/unused"
	"github.com/inkel/logfmt"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	logger *logfmt.Logger

	providers  *prometheus.GaugeVec
	disksCount *prometheus.GaugeVec
	duration   *prometheus.GaugeVec
}

func newMetrics(logger *logfmt.Logger) (metrics, error) {
	const (
		namespace = "unusedpds"
		subsystem = "provider"
	)

	// for providers metadata is small (and it should stay that way)
	// so we can use it as a label
	labels := []string{"name", "metadata"}

	m := metrics{
		logger: logger,

		providers: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "info",
			Help:      "Information about each cloud provider",
		}, labels),

		duration: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "fetch_duration_ms",
			Help:      "How long in milliseconds took to list the unused disks for this provider",
		}, labels),

		disksCount: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "unused_disks_count",
			Help:      "How many unused disks are currently in this provider",
		}, labels),
	}

	for _, metric := range []prometheus.Collector{
		m.providers,
		m.disksCount,
		m.duration,
	} {
		if err := prometheus.Register(metric); err != nil {
			return m, fmt.Errorf("registering metric %v: %w", metric, err)
		}
	}

	return m, nil
}

func (m metrics) Collect(ctx context.Context, providers []unused.Provider) {
	var wg sync.WaitGroup

	l := len(providers)

	m.logger.Log("collecting metrics", logfmt.Labels{"providers": l})
	wg.Add(l)

	for _, p := range providers {
		go func(p unused.Provider) {
			defer wg.Done()
			m.collect(ctx, p)
		}(p)
	}

	wg.Wait()
}

func (m metrics) collect(ctx context.Context, p unused.Provider) {
	labels := []string{p.Name(), p.Meta().String()}

	m.providers.WithLabelValues(labels...).Set(1)

	start := time.Now()

	disks, err := p.ListUnusedDisks(ctx)
	if err != nil {
		m.logger.Log("listing unused disks", logfmt.Labels{"provider": p.Name(), "meta": p.Meta(), "err": err})
		return
	}

	dur := time.Since(start)
	m.duration.WithLabelValues(labels...).Set(float64(dur.Milliseconds()))

	count := len(disks)
	m.disksCount.WithLabelValues(labels...).Set(float64(count))

	m.logger.Log("listing unused disks", logfmt.Labels{"provider": p.Name(), "meta": p.Meta(), "duration": dur, "count": count})

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
}
