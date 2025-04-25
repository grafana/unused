package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/grafana/unused"
	"github.com/grafana/unused/cmd/internal"
)

var k8sHeaders = map[string]string{
	KubernetesNS:  "K8S:NS",
	KubernetesPVC: "K8S:PVC",
	KubernetesPV:  "K8S:PV",
}

func Table(ctx context.Context, options Options) error {
	disks, err := listUnusedDisks(ctx, options.Providers)
	if err != nil {
		return err
	}

	if options.Filter.Key != "" {
		filtered := make(unused.Disks, 0, len(disks))
		for _, d := range disks {
			if d.Meta().Matches(options.Filter.Key, options.Filter.Value) {
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

	headers := []string{"PROVIDER", "DISK", "AGE", "UNUSED", "TYPE", "SIZE_GB"}
	for _, c := range options.ExtraColumns {
		h, ok := k8sHeaders[c]
		if !ok {
			h = "META:" + c
		}
		headers = append(headers, h)
	}
	if options.Verbose {
		headers = append(headers, "PROVIDER_META", "DISK_META")
	}

	fmt.Fprintln(w, strings.Join(headers, "\t")) // nolint:errcheck

	for _, d := range disks {
		p := d.Provider()

		row := []string{p.Name(), d.Name(), internal.Age(d.CreatedAt()), internal.Age(d.LastUsedAt()), string(d.DiskType()), fmt.Sprintf("%d", d.SizeGB())}
		meta := d.Meta()
		for _, c := range options.ExtraColumns {
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
		if options.Verbose {
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
