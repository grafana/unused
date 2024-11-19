package ui

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/grafana/unused"
)

// Disks are aggregated by the key composed by these 3 strings:
//  1. Provider
//  2. Value of the tag from disk metadata
//     Name of key passed in the options.Group
//     If requested key is absent in disk metadata, use value "NONE"
//  3. Disk type: "hdd", "ssd" or "unknown"
type groupKey [3]string

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

	headers := []string{"PROVIDER", options.Group, "TYPE", "DISKS_COUNT", "TOTAL_SIZE_GB"}
	totalSize := make(map[groupKey]int)
	totalCount := make(map[groupKey]int)

	fmt.Fprintln(w, strings.Join(headers, "\t")) // nolint:errcheck

	var aggrValue string
	for _, d := range disks {
		p := d.Provider()
		if value, ok := d.Meta()[options.Group]; ok {
			aggrValue = value
		} else {
			aggrValue = "NONE"
		}
		aggrKey := groupKey{p.Name(), aggrValue, string(d.DiskType())}
		totalSize[aggrKey] += d.SizeGB()
		totalCount[aggrKey] += 1
	}

	keys := make([]groupKey, 0, len(totalSize))
	for k := range totalSize {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		for k := 0; k < len(keys[i]); k++ {
			if keys[i][k] != keys[j][k] {
				return keys[i][k] < keys[j][k]
			}
		}
		return true
	})

	for _, aggrKey := range keys {
		row := aggrKey[:]
		row = append(row, strconv.Itoa(totalCount[aggrKey]), strconv.Itoa(totalSize[aggrKey]))
		fmt.Fprintln(w, strings.Join(row, "\t")) // nolint:errcheck
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flushing table contents: %w", err)
	}

	return nil
}
