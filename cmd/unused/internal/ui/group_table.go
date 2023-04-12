package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/grafana/unused"
)

func GroupTable(ctx context.Context, options Options) error {
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

	headers := []string{"PROVIDER", "GROUP_BY", "TYPE", "SIZE_GB"}
	aggrData := make(map[[3]string]int)

	fmt.Fprintln(w, strings.Join(headers, "\t"))

	var aggrValue string
	for _, d := range disks {
		p := d.Provider()
		if value, ok := d.Meta()[options.Group]; ok {
			aggrValue = value
		} else {
			aggrValue = "NONE"
		}
		aggrKey := [3]string{p.Name(), aggrValue, string(d.DiskType())}
		aggrData[aggrKey] += d.SizeGB()
	}

	for info, totalSize := range aggrData {
		row := info[:]
		row = append(row, fmt.Sprintf("%d", totalSize))
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flushing table contents: %w", err)
	}

	return nil
}
