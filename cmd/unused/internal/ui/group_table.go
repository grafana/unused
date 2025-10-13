package ui

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"
)

// Disks are aggregated by the key composed by these 3 strings:
//  1. Provider
//  2. Value of the tag from disk metadata
//     Name of key passed in the options.Group
//     If requested key is absent in disk metadata, use value "NONE"
//  3. Disk type: "hdd", "ssd" or "unknown"
type groupKey [3]string

func GroupTable(ctx context.Context, ui UI) error {
	disks, err := ui.listUnusedDisks(ctx)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(ui.Out, 8, 4, 2, ' ', 0)

	var groupByHeader string
	switch ui.Group {
	case "k8s:ns", "k8s:pvc", "k8s:pv":
		groupByHeader = strings.ToUpper(ui.Group)
	default:
		groupByHeader = ui.Group
	}

	headers := []string{"PROVIDER", groupByHeader, "TYPE", "DISKS_COUNT", "TOTAL_SIZE_GB"}
	totalSize := make(map[groupKey]int)
	totalCount := make(map[groupKey]int)

	fmt.Fprintln(w, strings.Join(headers, "\t")) // nolint:errcheck

	var aggrValue string
	for _, d := range disks {
		var (
			value string
			ok    bool
			meta  = d.Meta()
		)

		switch ui.Group {
		case "k8s:ns":
			value = meta.CreatedForNamespace()
			ok = value != ""
		case "k8s:pvc":
			value = meta.CreatedForPVC()
			ok = value != ""
		case "k8s:pv":
			value = meta.CreatedForPV()
			ok = value != ""
		default:
			value, ok = meta[ui.Group]
		}

		if ok {
			aggrValue = value
		} else {
			aggrValue = "NONE"
		}

		aggrKey := groupKey{d.Provider().Name(), aggrValue, string(d.DiskType())}
		totalSize[aggrKey] += d.SizeGB()
		totalCount[aggrKey] += 1
	}

	keys := slices.SortedFunc(maps.Keys(totalSize), func(a, b groupKey) int {
		for k := range b {
			if c := strings.Compare(a[k], b[k]); c != 0 {
				return c
			}
		}
		return 0
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
