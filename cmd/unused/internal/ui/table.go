package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/grafana/unused/cmd/internal"
)

var k8sHeaders = map[string]string{
	KubernetesNS:  "K8S:NS",
	KubernetesPVC: "K8S:PVC",
	KubernetesPV:  "K8S:PV",
}

func Table(ctx context.Context, ui UI) error {
	disks, err := ui.listUnusedDisks(ctx)
	if err != nil {
		return err
	}

	if len(disks) == 0 {
		fmt.Println("No disks found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 8, 4, 2, ' ', 0)

	headers := []string{"PROVIDER", "DISK", "AGE", "UNUSED", "TYPE", "SIZE_GB"}
	for _, c := range ui.ExtraColumns {
		h, ok := k8sHeaders[c]
		if !ok {
			h = "META:" + c
		}
		headers = append(headers, h)
	}
	if ui.Verbose {
		headers = append(headers, "PROVIDER_META", "DISK_META")
	}

	fmt.Fprintln(w, strings.Join(headers, "\t")) // nolint:errcheck

	for _, d := range disks {
		p := d.Provider()

		row := []string{p.Name(), d.Name(), internal.Age(d.CreatedAt()), internal.Age(d.LastUsedAt()), string(d.DiskType()), fmt.Sprintf("%d", d.SizeGB())}
		meta := d.Meta()
		for _, c := range ui.ExtraColumns {
			var v string
			switch c {
			case KubernetesNS:
				v = meta.CreatedForNamespace()
			case KubernetesPV:
				v = meta.CreatedForPV()
			case KubernetesPVC:
				v = meta.CreatedForPVC()
			default:
				v = meta[c]
			}
			if v == "" {
				v = "-"
			}
			row = append(row, v)
		}
		if ui.Verbose {
			row = append(row, p.Meta().String(), d.Meta().String())
		}

		fmt.Fprintln(w, strings.Join(row, "\t")) // nolint:errcheck
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flushing table contents: %w", err)
	}

	return nil
}
