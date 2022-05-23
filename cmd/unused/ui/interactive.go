package ui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/grafana/unused"
	"github.com/grafana/unused/cmd/unused/ui/interactive"
)

func Interactive(ctx context.Context, options Options) error {
	m := interactive.New(options.Providers)

	if err := tea.NewProgram(m).Start(); err != nil {
		return fmt.Errorf("cannot start interactive UI: %w", err)
	}

	return nil
}

func oldInteractive(ctx context.Context, options Options) error {
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

	m := interactive.NewModel(options.Verbose, disks, options.ExtraColumns)

	if err := tea.NewProgram(m).Start(); err != nil {
		return fmt.Errorf("cannot start interactive UI: %w", err)
	}
	return nil
}
