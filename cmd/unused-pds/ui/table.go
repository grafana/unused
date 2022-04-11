package ui

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/grafana/unused-pds/pkg/unused"
)

func DumpAsTable(out io.Writer, disks unused.Disks) error {
	w := tabwriter.NewWriter(out, 8, 2, 0, ' ', 0)

	fmt.Fprintln(w, "PROVIDER\tNAME\tMETADATA")
	for _, d := range disks {
		fmt.Fprintf(w, "%s\t%s\t%s\n", d.Provider().Name(), d.Name(), d.Meta())
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flushing table contents: %w", err)
	}

	return nil
}
