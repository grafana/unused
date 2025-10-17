# CLI Tool, Prometheus Exporter, and Go Module to List Your Unused Disks in All Cloud Providers
This repository contains a Go library to list your unused persistent disks in different cloud providers, and binaries for displaying them at the CLI or exporting Prometheus metrics.

At Grafana Labs we host our workloads in multiple cloud providers.
Our workloads orchestration is managed by Kubernetes, and we've found that due to some misconfiguration in the backend storage, we used to have lots of unused resources, specially persistent disks.
These leaked resources cost money, and because these are resources that are not in use anymore, it translates into wasted money.
This library and its companion tools should help you out to identify these resources and clean them up.

## Go Module `github.com/grafana/unused`
This module exports some interfaces and implementations to easily list all your unused persistent disks in GCP, AWS, and Azure.
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

#### Notes on Authentication
Both binaries are opinionated on how to authenticate against each Cloud Service Provider (CSP).

| Provider | Notes |
|-|-|
| GCP | Depends on [default credentials](https://cloud.google.com/docs/authentication/application-default-credentials) |
| AWS | Uses profile names from your [credentials file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) or `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_REGION` env variables |
| Azure | Either specify an `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, and `AZURE_TENANT_ID`, or requires [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/) installed on the host and [signed in](https://learn.microsoft.com/en-us/cli/azure/authenticate-azure-cli) |

### `unused` Binary
TUI tool to query all given providers and list them as a neat table.
It also supports an interactive mode which allows to select and delete disks in an easy way.

```
go install github.com/grafana/unused/cmd/unused@latest
```

#### Example Usage of `unused`

Below are examples of using the tool for Azure, GCP, and AWS.

```shell
unused -azure.sub=AZURE_SUBSCRIPTION -add-k8s-column=ns -add-k8s-column=pvc -add-k8s-column=pv -v -csv

unused -gcp.project=GCP_PROJECT_NAME -add-k8s-column=ns -add-k8s-column=pvc -add-k8s-column=pv -v -csv

unused -aws.profile=AWS_PROFILE -add-k8s-column=ns -add-k8s-column=pvc -add-k8s-column=pv -v -csv

```

##### Output (Table or CSV)

Saving the results as a CSV can ease import into other tools but viewing the data as a human readable table can also be nice. The CLI interface provides both options as shown below.

```shell
# output as table to STDOUT
./unused -gcp.project=GCP_PROJECT_NAME -add-k8s-column=ns -add-k8s-column=pvc -add-k8s-column=pv -v

# output as CSV to STDOUT
./unused -gcp.project=GCP_PROJECT_NAME -add-k8s-column=ns -add-k8s-column=pvc -add-k8s-column=pv -v -csv
```

### `unused-exporter` Prometheus Exporter
Web server exposing Prometheus metrics about each providers count of unused disks.
It exposes the following metrics:

| Metric | Description |
|-|-|
| `unused_disks_count` | How many unused disks are in this provider |
| `unused_disks_total_size_bytes` | Total size of unused disks in this provider in bytes |
| `unused_disk_size_bytes` | Size of each disk in bytes |
| `unused_disks_last_used_timestamp_seconds` | Last timestamp (unix seconds) when this disk was used. GCP only! |
| `unused_provider_duration_seconds` | How long in seconds took to fetch this provider information |
| `unused_provider_info` | CSP information |
| `unused_provider_success` | Static metric indicating if collecting the metrics succeeded or not |

All metrics have the `provider` and `provider_id` labels to identify to which provider instance they belong.
The `unused_disks_count`, `unused_disk_size_bytes`, and `unused_disks_total_size_bytes` metrics have an additional `k8s_namespace` metric mapped to the `kubernetes.io/created-for/pvc/namespace` annotation assigned to persistent disks created by Kubernetes.

Information about each unused disk is currently logged to stdout given that it contains more changing information that could lead to cardinality explosion.

```
go install github.com/grafana/unused/cmd/unused-exporter@latest
```

## Testing Against Fake Providers
In order to make E2E and UI testing easier, we implemented a fake provider that is only available when running `go` with the `-tags=fake` flag.
Usage of this flag should produce a deterministic output of fake unused disks for different providers.
For example, the following runs the `unused` binary using the fake provider:

```
➜ go run -tags=fake ./cmd/unused
time=2025-10-08T15:06:22.926-03:00 level=WARN msg="Using fake provider"
PROVIDER  DISK                                               AGE     UNUSED  TYPE    SIZE_GB
Medium    pvc-000-11371241257079532652-14470142590855381128  1y      210d    ssd     79
Medium    pvc-001-760102831717374652-9221744211007427193     128d    64d     ssd     44
Medium    pvc-002-3389241988064777392-12210202232702069999   72d     36d     ssd     44
Medium    pvc-003-2240328155279531677-7311121042813227358    1y      257d    ssd     78
Medium    pvc-004-9381769212557126946-1350674201389090105    1y      199d    ssd     33
Medium    pvc-005-11814882063598695543-3824056318896229933   26d     13d     ssd     2
Medium    pvc-006-13037077871211764336-13617661536098221011  182d    91d     ssd     82
Medium    pvc-007-8499734460532927466-10917103977888096435   1y      262d    ssd     67
Medium    pvc-008-137166597355566241-11349393104106307790    274d    137d    ssd     94
Medium    pvc-009-1873294050595685917-6999251097555031736    1y      191d    hdd     90
Medium    pvc-010-7129530298895469460-12892292024388140256   1y      230d    hdd     5
Medium    pvc-011-7832170631747924588-17404438405895559565   345d    172d    hdd     55
Medium    pvc-012-15199846812963376007-15282558343396517750  1y      243d    ssd     89
Medium    pvc-013-5425635553360250513-2558341386968588437    356d    178d    ssd     58
```

This flag is also available for running tests: `go test -tags=fake ./...`
