package interactive

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/grafana/unused"
	unusedui "github.com/grafana/unused/cmd/unused/ui"
)

var _ unusedui.UI = &ui{}

type ui struct {
	list list.Model

	selected map[int]struct{}
	disks    unused.Disks
	verbose  bool
}

func New(verbose bool) unusedui.UI {
	return &ui{
		verbose: verbose,
	}
}

func (ui *ui) Display(ctx context.Context, disks unused.Disks) error {
	ui.disks = disks

	ui.list = list.New(nil, list.NewDefaultDelegate(), 0, 0)
	ui.list.Title = "Cloud Providers Unused Disks"

	ui.refresh()

	if err := tea.NewProgram(ui).Start(); err != nil {
		return fmt.Errorf("cannot start interactive UI: %w", err)
	}
	return nil
}

func (ui *ui) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (ui *ui) refresh() {
	items := make([]list.Item, len(ui.disks))
	for i, d := range ui.disks {
		items[i] = item{d, ui.verbose}
	}

	ui.list.SetItems(items)
}

func (ui *ui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return ui, tea.Quit

		case "v":
			ui.verbose = !ui.verbose
			ui.refresh()
		}

	case tea.WindowSizeMsg:
		w, h := 0, 0
		ui.list.SetSize(msg.Width-w, msg.Height-h)
	}

	// pass update msg to the rest of the components
	list, cmd := ui.list.Update(msg)
	ui.list = list

	return ui, cmd
}

func (ui *ui) View() string {
	return ui.list.View()
}

func (ui *ui) itemUpdateFunc(msg tea.Msg, list *list.Model) tea.Cmd {
	idx := list.Index()
	_, unmark := ui.selected[idx]
	if unmark {
		delete(ui.selected, idx)
	} else {
		ui.selected[idx] = struct{}{}
	}

	fmt.Println(ui.selected)

	return nil
}
