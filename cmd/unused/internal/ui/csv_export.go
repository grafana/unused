package ui

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/grafana/unused"
	"github.com/grafana/unused/cmd/internal"
)

func CsvExport(ctx context.Context, options Options) error {
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
		return nil
	}

	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	headers := []string{"PROVIDER", "ACCOUNT", "DISK", "AGE", "UNUSED", "TYPE", "SIZE_GB"}
	if options.RawDate {
		headers[3] = "CREATED"
	}
	for _, c := range options.ExtraColumns {
		headers = append(headers, "META:"+c)
	}
	if options.Verbose {
		headers = append(headers, "PROVIDER_META", "DISK_META")
	}

	w.Write(headers)

	for _, d := range disks {
		p := d.Provider()

		row := []string{p.Name(), p.Account(), d.Name(), internal.Age(d.CreatedAt()), internal.Age(d.LastUsedAt()), string(d.DiskType()), fmt.Sprintf("%d", d.SizeGB())}
		if options.RawDate {
			row[3] = d.CreatedAt().Format(time.RFC3339)
			row[4] = d.LastUsedAt().Format(time.RFC3339)
		}
		meta := d.Meta()
		for _, c := range options.ExtraColumns {
			row = append(row, meta[c])
		}
		if options.Verbose {
			row = append(row, p.Meta().String(), d.Meta().String())
		}

		w.Write(row)
	}

	return nil
}