package ui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/grafana/unused/cmd/unused/internal/ui/interactive"
)

func Interactive(ctx context.Context, options Options) error {
	m := interactive.New(options.Providers, options.ExtraColumns, options.Filter.Key, options.Filter.Value, options.DryRun)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		return fmt.Errorf("cannot start interactive UI: %w", err)
	}

	return nil
}
