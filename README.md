# CLI tool, Prometheus exporter, and Go module to list your unused disks in all cloud providers
This repository contains a Go library to list your unused persistent disks in different cloud providers, and binaries for displaying them at the CLI or exporting Prometheus metrics.

At Grafana Labs we host our workloads in multiple cloud providers.
Our workloads orchestration is managed by Kubernetes, and we've found that due to some misconfiguration in the backend storage, we used to have lots of unused resources, specially persistent disks.
These leaked resources cost money, and because these are resources that are not in use anymore, it translates into wasted money.
This library and its companion tools should help you out to identify these resources and clean them up.

## Go module `github.com/grafana/unused`
This module exports some interfaces and implementations to easily list all your unsed persistent disks in GCP, AWS, and Azure.
You can find the API in the [Go package documentation](https://pkg.go.dev/github.com/grafana/unused).

## Binaries
This repository also provides two binaries ready to use to interactively view unused disks (`cmd/unused`) or expose unused disk metrics (`cmd/unused-exporter`) to [Prometheus](https://prometheus.io).
Both programs can authenticate against the following providers using the listed CLI flags:

| Provider | Flag | Description |
|-|-|-|
| GCP | `-gcp.project` | ID of the GCP project |
| AWS | `-aws.profile` | AWS configuration profile name |
| Azure | `-azure.sub` | Azure subscription ID |

These flags can be specified more than once, allowing to have different configurations for each provider.

#### Notes on authentication
Both binaries are opinionated on how to authenticate against each Cloud Service Provider (CSP).

| Provider | Notes |
|-|-|
| GCP | Depends on [default credentials](https://cloud.google.com/docs/authentication/application-default-credentials) |
| AWS | Uses profile names from your [credentials file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) |
| Azure | Requires [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/) installed on the host and [signed in](https://learn.microsoft.com/en-us/cli/azure/authenticate-azure-cli) |

### `unused` binary
TUI tool to query all given providers and list them as a neat table.
It also supports an interactive mode which allows to select and delete disks in an easy way.

```
go install github.com/grafana/unused/cmd/unused@latest
```

### `unused-exporter` Prometheus exporter
Web server exposing Prometheus metrics about each providers count of unused disks.
It exposes the following metrics:

| Metric | Description |
|-|-|
| `unused_disks_count` | How many unused disks are in this provider |
| `unused_provider_duration_ms` | How long in milliseconds took to fetch this provider information |
| `unused_provider_info` | CSP information |
| `unused_provider_success` | Static metric indicating if collecting the metrics succeeded or not |

All metrics have the `provider` and `provider_id` labels to identify to which provider instance they belong.
The `unused_disks_count` metric has an additional `k8s_namespace` metric mapped to the `kubernetes.io/created-for/pvc/namespace` annotation assigned to persistent disks created by Kubernetes.

Information about each unused disk is currently logged to stdout given that it contains more changing information that could lead to cardinality explosion.

```
go install github.com/grafana/unused/cmd/unused-exporter@latest
```
