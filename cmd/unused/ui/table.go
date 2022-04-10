package ui

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/grafana/unused"
)

var _ Displayer = Table{}

type Table struct {
	out io.Writer
}

func NewTable(out io.Writer) Table {
	return Table{out}
}

func (t Table) Display(ctx context.Context, disks unused.Disks) error {
	w := tabwriter.NewWriter(t.out, 8, 2, 0, ' ', 0)

	fmt.Fprintln(w, "PROVIDER\tNAME\tMETADATA")
	for _, d := range disks {
		fmt.Fprintf(w, "%s\t%s\t%s\n", d.Provider().Name(), d.Name(), d.Meta())
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flushing table contents: %w", err)
	}

	return nil
}
