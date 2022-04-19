package interactive

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/grafana/unused"
	unusedui "github.com/grafana/unused/cmd/unused/ui"
)

var _ unusedui.UI = &ui{}

type ui struct {
	list    list.Model
	lbox    lipgloss.Style
	tabs    *Tabs
	sidebar *Sidebar

	selected map[int]struct{}
	disks    map[string]unused.Disks
	verbose  bool
}

func New(verbose bool) unusedui.UI {
	return &ui{
		verbose: verbose,
	}
}

func (ui *ui) Display(ctx context.Context, disks unused.Disks) error {
	ui.selected = make(map[int]struct{})
	ui.list = list.New(nil, ui.listDelegate(), 0, 0)
	ui.list.SetShowTitle(false)
	ui.list.DisableQuitKeybindings()

	ui.lbox = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, true, true, false)

	ui.sidebar = NewSidebar()
	ui.sidebar.Style = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, true, true)

	ui.disks = make(map[string]unused.Disks)
	for _, d := range disks {
		p := d.Provider().Name()
		ui.disks[p] = append(ui.disks[p], d)
	}
	titles := make([]string, 0, len(ui.disks))
	for p := range ui.disks {
		titles = append(titles, p)
	}
	sort.Strings(titles)
	ui.tabs = &Tabs{Titles: titles}

	ui.refresh(true)

	if err := tea.NewProgram(ui).Start(); err != nil {
		return fmt.Errorf("cannot start interactive UI: %w", err)
	}
	return nil
}

func (ui *ui) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (ui *ui) refresh(reset bool) {
	disks := ui.disks[ui.tabs.Selected()]
	items := make([]list.Item, len(disks))
	for i, d := range disks {
		items[i] = item{d, ui.verbose}
	}
	ui.list.SetItems(items)

	if reset {
		ui.list.ResetSelected()
		ui.refreshSidebar(disks[0])
	}
}

func (ui *ui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return ui, tea.Quit

		case "right":
			ui.tabs.Next()
			ui.refresh(true)
			return ui, nil

		case "left":
			ui.tabs.Prev()
			ui.refresh(true)
			return ui, nil

		case "v":
			ui.verbose = !ui.verbose
			ui.list.SetDelegate(ui.listDelegate())
			ui.refresh(false)
		}

	case tea.WindowSizeMsg:
		h := msg.Height - lipgloss.Height(ui.tabs.View()) - 2
		w := (msg.Width / 2)
		ui.lbox.Width(w)
		ui.list.SetSize(w, h)
		ui.sidebar.SetSize(w, h)
	}

	// pass update msg to the rest of the components
	list, cmd := ui.list.Update(msg)
	ui.list = list

	return ui, cmd
}

func (ui *ui) View() string {
	var out strings.Builder

	out.WriteString(ui.tabs.View())
	out.WriteRune('\n')

	content := lipgloss.JoinHorizontal(lipgloss.Top,
		ui.lbox.Render(ui.list.View()),
		ui.sidebar.View(),
	)
	out.WriteString(content)

	return out.String()
}

func (ui *ui) refreshSidebar(disk unused.Disk) {
	printMeta := func(meta unused.Meta) {
		if len(meta) == 0 {
			return
		}
		ui.sidebar.WriteHeader("Metadata")
		ui.sidebar.Println()
		for _, k := range meta.Keys() {
			ui.sidebar.Printf("%s: %s\n", k, meta[k])
		}
	}

	ui.sidebar.Reset()

	ui.sidebar.WriteHeader("Created: ")
	ui.sidebar.WriteString(disk.CreatedAt().Format(time.RFC3339))
	ui.sidebar.Printf(" (%s)", age(disk.CreatedAt()))
	ui.sidebar.Println()

	printMeta(disk.Meta())
	ui.sidebar.Println()

	ui.sidebar.WriteHeader("Provider: ")
	ui.sidebar.Println(disk.Provider().Name())
	printMeta(disk.Provider().Meta())
}

func (ui *ui) listDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.ShowDescription = ui.verbose
	d.UpdateFunc = func(msg tea.Msg, list *list.Model) tea.Cmd {
		item, ok := list.SelectedItem().(item)
		if ok { // this should always happen
			ui.refreshSidebar(item.disk)
		}

		return nil
	}
	return d
}

func age(date time.Time) string {
	d := time.Since(date)

	if d <= time.Minute {
		return "1m"
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", d/time.Minute)
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh", d/time.Hour)
	} else if d < 365*24*time.Hour {
		return fmt.Sprintf("%dd", d/(24*time.Hour))
	} else {
		return fmt.Sprintf("%dy", d/(365*24*time.Hour))
	}
}
