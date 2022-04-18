# CLI tool, Prometheus exporter, and Go module to list your unused disks in all cloud providers
This repository contains a Go library to list your unused persistent disks in different cloud providers, and binaries for displaying them at the CLI or exporting Prometheus metrics.

At Grafana Labs we host our workloads in multiple cloud providers.
Our workloads orchestration is managed by Kubernetes, and we've found that due to some misconfiguration in the backend storage, we usually had lots of unused resources, specially persistent disks.
These leaked resources cost money, and because these are resources that are not in use anymore, it translates into wasted money.
This library and its companion tools should help you out to identify these resources and clean them up.

This repository provides the following:

- Go module `github.com/grafana/unused`
  This module exports some interfaces and implementations to easily list all your unsed persistent disks in GCP, AWS, and Azure.

- `unused` tool
  This CLI tool will query all given providers and list them as a neat table.
```
go install github.com/grafana/unused/cmd/unused@latest
```

- `unused-exporter` Prometheus exporter
  This tool will run a web server exposing Prometheus metrics about each providers count of unused disks.
  The exposed metrics currently are:

  * `unusedpds_provider_fetch_duration_ms`: How long in milliseconds took to list the unused disks for this provider (gauge).
  * `unusedpds_provider_info`: Information about each cloud provider (gauge).
  * `unusedpds_provider_unused_disks_count`: How many unused disks are currently in this provider (gauge).

  All metrics have the labels `provider` with the provider's name and `metadata` with some simple metadata for each provider.
  Provider's metadata **should** always be small enough that it won't transform into a label cardinality explosion.

  Information about each unused disk is currently logged to stdout given that it contains more changing information that could lead to cardinality explosion.
  We might revise this in the future, but having the information as a log stream is useful if these are forwarded to Loki.

```
go install github.com/grafana/unused/cmd/unused-exporter@latest
```
