package ui

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/grafana/unused/cmd/internal"
)

var k8sHeaders = map[string]string{
	KubernetesNS:  "K8S_NS",
	KubernetesPVC: "K8S_PVC",
	KubernetesPV:  "K8S_PV",
}

type textWriter interface {
	Headers(hdrs []string)
	AddRow(cols []string)
	Flush() error
}

func text(ctx context.Context, ui UI, w textWriter) error {
	disks, err := ui.listUnusedDisks(ctx)
	if err != nil {
		return err
	}

	if len(disks) == 0 {
		fmt.Println("No disks found")
		return nil
	}

	headers := []string{"PROVIDER", "DISK", "AGE", "UNUSED", "TYPE", "SIZE_GB"}
	for _, c := range ui.ExtraColumns {
		h, ok := k8sHeaders[c]
		if !ok {
			h = "META_" + c
		}
		headers = append(headers, h)
	}
	if ui.Verbose {
		headers = append(headers, "PROVIDER_META", "DISK_META")
	}

	w.Headers(headers)

	for _, d := range disks {
		p := d.Provider()

		row := []string{
			p.Name(),
			d.Name(),
			internal.Age(d.CreatedAt()),
			internal.Age(d.LastUsedAt()),
			string(d.DiskType()),
			fmt.Sprintf("%d", d.SizeGB()),
		}

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

		w.AddRow(row)
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flushing contents: %w", err)
	}

	return nil
}

type csvWriter struct {
	w *csv.Writer
}

func (w csvWriter) Headers(hdrs []string) {
	w.w.Write(hdrs) // nolint:errcheck
}

func (w csvWriter) AddRow(cols []string) {
	w.w.Write(cols) // nolint:errcheck
}

func (w csvWriter) Flush() error {
	w.w.Flush()
	return w.w.Error()
}

func CSV(ctx context.Context, ui UI) error {
	w := csvWriter{
		w: csv.NewWriter(ui.Out),
	}

	return text(ctx, ui, w)
}

type tableWriter struct {
	w *tabwriter.Writer
}

func (w tableWriter) Headers(hdrs []string) {
	fmt.Fprintln(w.w, strings.Join(hdrs, "\t")) // nolint:errcheck
}

func (w tableWriter) AddRow(cols []string) {
	fmt.Fprintln(w.w, strings.Join(cols, "\t")) // nolint:errcheck
}

func (w tableWriter) Flush() error { return w.w.Flush() }

func Table(ctx context.Context, ui UI) error {
	w := tableWriter{
		w: tabwriter.NewWriter(ui.Out, 8, 4, 2, ' ', 0),
	}

	return text(ctx, ui, w)
}
