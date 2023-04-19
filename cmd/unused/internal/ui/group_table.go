package ui

import (
	"context"
	"fmt"
	"os"
	"sort"
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

	if options.Verbose {
		fmt.Printf("Grouping by '%s' tag.\n", options.Group)
	}
	headers := []string{"PROVIDER", strings.ToUpper(options.Group), "TYPE", "DISKS_COUNT", "TOTAL_SIZE_GB"}
	// headers := []string{"PROVIDER", "GROUP_BY_TAG", "TYPE", "DISKS_COUNT", "TOTAL_SIZE_GB"}
	totalSize := make(map[[3]string]int)
	totalCount := make(map[[3]string]int)

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
		totalSize[aggrKey] += d.SizeGB()
		totalCount[aggrKey] += 1
	}

	keys := make([][3]string, 0, len(totalSize))
	for k := range totalSize {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		for k := 0; k < len(keys[i]); k += 1 {
			if keys[i][k] != keys[j][k] {
				return keys[i][k] < keys[j][k]
			}
		}
		return true
	})
	for _, info := range keys {
		row := info[:]
		row = append(row, fmt.Sprintf("%d", totalCount[info]))
		row = append(row, fmt.Sprintf("%d", totalSize[info]))
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flushing table contents: %w", err)
	}

	return nil
}
