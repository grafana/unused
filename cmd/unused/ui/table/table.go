package table

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/grafana/unused"
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

func (t table) Display(ctx context.Context, disks unused.Disks) error {
	w := tabwriter.NewWriter(t.out, 8, 4, 2, ' ', 0)

	fmt.Fprint(w, "PROVIDER\tNAME")
	if t.verbose {
		fmt.Fprint(w, "\tMETADATA")
	}
	fmt.Fprintln(w)

	for _, d := range disks {
		p := d.Provider()
		if t.verbose {
			fmt.Fprintf(w, "%s{%s}\t%s\t%s\n", p.Name(), p.Meta(), d.Name(), d.Meta())
		} else {
			fmt.Fprintf(w, "%s\t%s\n", p.Name(), d.Name())
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flushing table contents: %w", err)
	}

	return nil
}
