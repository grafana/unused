package ui

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/grafana/unused"
	"github.com/grafana/unused/cmd/internal"
)

var k8sHeaders = map[string]string{
	KubernetesNS:  "K8S_NS",
	KubernetesPVC: "K8S_PVC",
	KubernetesPV:  "K8S_PV",
}

func buildHeaders(ui UI) []string {
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
	return headers
}

func CSV(ctx context.Context, ui UI) error {
	disks, err := listUnusedDisks(ctx, ui.Providers)
	if err != nil {
		return err
	}

	disks = disks.Filter(ui.FilterFunc)

	if len(disks) == 0 {
		fmt.Println("No disks found")
		return nil
	}

	w := csv.NewWriter(os.Stdout)

	headers := buildHeaders(ui)

	if err := w.Write(headers); err != nil {
		return fmt.Errorf("writing headers: %w", err)
	}

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

			row = append(row, v)
		}

		if ui.Verbose {
			row = append(row, p.Meta().String(), d.Meta().String())
		}

		if err := w.Write(row); err != nil {
			return fmt.Errorf("writing row for disk %q: %w", d.Name(), err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("flushing CSV contents: %w", err)
	}
	return nil
}

func Table(ctx context.Context, ui UI) error {
	disks, err := listUnusedDisks(ctx, ui.Providers)
	if err != nil {
		return err
	}

	if ui.Filter.Key != "" {
		filtered := make(unused.Disks, 0, len(disks))
		for _, d := range disks {
			if d.Meta().Matches(ui.Filter.Key, ui.Filter.Value) {
				filtered = append(filtered, d)
			}
		}
		disks = filtered
	}

	if len(disks) == 0 {
		fmt.Println("No disks found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 8, 4, 2, ' ', 0)

	headers := buildHeaders(ui)

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

func listUnusedDisks(ctx context.Context, providers []unused.Provider) (unused.Disks, error) {
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		total unused.Disks
	)

	wg.Add(len(providers))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, len(providers))

	for _, p := range providers {
		go func(p unused.Provider) {
			defer wg.Done()

			disks, err := p.ListUnusedDisks(ctx)
			if err != nil {
				cancel()
				errCh <- fmt.Errorf("%s %s: %w", p.Name(), p.Meta(), err)
				return
			}

			mu.Lock()
			total = append(total, disks...)
			mu.Unlock()
		}(p)
	}

	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
	}

	return total, nil
}
