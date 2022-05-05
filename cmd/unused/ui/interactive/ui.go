package interactive

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/grafana/unused"
	"github.com/grafana/unused/cmd/unused/ui"
)

var _ ui.UI = &interactive{}

type interactive struct {
	verbose bool
}

func New(verbose bool) ui.UI {
	return &interactive{
		verbose: verbose,
	}
}

func (ui *interactive) Display(ctx context.Context, disks unused.Disks, extraColumns []string) error {
	m := NewModel(ui.verbose, disks, extraColumns)

	if err := tea.NewProgram(m).Start(); err != nil {
		return fmt.Errorf("cannot start interactive UI: %w", err)
	}
	return nil
}
