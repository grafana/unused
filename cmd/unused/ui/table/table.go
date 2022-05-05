package table

import (
	"context"
	"fmt"
	"io"
	"strings"
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

func (t table) Display(ctx context.Context, disks unused.Disks, extraColumns []string) error {
	w := tabwriter.NewWriter(t.out, 8, 4, 2, ' ', 0)

	headers := []string{"PROVIDER", "DISK", "AGE"}
	for _, c := range extraColumns {
		headers = append(headers, "META:"+c)
	}
	if t.verbose {
		headers = append(headers, "PROVIDER_META", "DISK_META")
	}

	fmt.Fprintln(w, strings.Join(headers, "\t"))

	for _, d := range disks {
		p := d.Provider()

		row := []string{p.Name(), d.Name(), cli.Age(d.CreatedAt())}
		for _, c := range extraColumns {
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
