package table

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/grafana/unused"
	"github.com/grafana/unused/cli"
	"github.com/grafana/unused/cmd/unused/ui"
)

var _ ui.UI = table{}

type table struct {
	out     io.Writer
	verbose bool
}

func New(out io.Writer, verbose bool) table {
	return table{out, verbose}
}

func (t table) Display(ctx context.Context, options ui.Options) error {
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

	w := tabwriter.NewWriter(t.out, 8, 4, 2, ' ', 0)

	headers := []string{"PROVIDER", "DISK", "AGE", "UNUSED"}
	for _, c := range options.ExtraColumns {
		headers = append(headers, "META:"+c)
	}
	if t.verbose {
		headers = append(headers, "PROVIDER_META", "DISK_META")
	}

	fmt.Fprintln(w, strings.Join(headers, "\t"))

	for _, d := range disks {
		p := d.Provider()

		row := []string{p.Name(), d.Name(), cli.Age(d.CreatedAt()), cli.Age(d.LastUsedAt())}
		for _, c := range options.ExtraColumns {
			row = append(row, d.Meta()[c])
		}
		if t.verbose {
			row = append(row, p.Meta().String(), d.Meta().String())
		}

		fmt.Fprintln(w, strings.Join(row, "\t"))
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
