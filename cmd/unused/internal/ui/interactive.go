package ui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/grafana/unused/cmd/unused/internal/ui/interactive"
)

func Interactive(ctx context.Context, ui UI) error {
	m := interactive.New(ui.Providers, ui.ExtraColumns, ui.Filter, ui.DryRun)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		return fmt.Errorf("cannot start interactive UI: %w", err)
	}

	return nil
}
