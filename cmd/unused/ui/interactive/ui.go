package interactive

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/grafana/unused"
	unusedui "github.com/grafana/unused/cmd/unused/ui"
)

var _ unusedui.UI = &ui{}

var listKeyMap = struct {
	Mark, Exec, Quit, Up, Down, PageUp, PageDown, Right, Left, Verbose key.Binding
}{
	Mark:     key.NewBinding(key.WithKeys("m", " "), key.WithHelp("m", "toggle mark")),
	Exec:     key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "delete")),
	Quit:     key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	Up:       key.NewBinding(key.WithKeys("up"), key.WithHelp("up", "move up one line")),
	Down:     key.NewBinding(key.WithKeys("down"), key.WithHelp("down", "move down one line")),
	Right:    key.NewBinding(key.WithKeys("right"), key.WithHelp("→", "next provider")),
	Left:     key.NewBinding(key.WithKeys("left"), key.WithHelp("←", "previous provider")),
	Verbose:  key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "toggle verbose mode")),
	PageUp:   key.NewBinding(key.WithKeys("pgup"), key.WithHelp("page up", "move up one page")),
	PageDown: key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("page down", "move down one page")),
}

type ui struct {
	list    list.Model
	lbox    lipgloss.Style
	tabs    *Tabs
	sidebar *Sidebar
	output  *output
	help    helpview

	selected map[string]map[int]struct{}
	disks    map[string]unused.Disks
	verbose  bool
}

func New(verbose bool) unusedui.UI {
	return &ui{
		verbose: verbose,
		help:    NewHelp(listKeyMap.Mark, listKeyMap.Exec, listKeyMap.Quit, listKeyMap.Up, listKeyMap.Down, listKeyMap.PageUp, listKeyMap.PageDown, listKeyMap.Left, listKeyMap.Right, listKeyMap.Verbose),
	}
}

func (ui *ui) Display(ctx context.Context, disks unused.Disks) error {
	ui.selected = make(map[string]map[int]struct{})
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
		ui.selected[p] = make(map[int]struct{})
		titles = append(titles, p)
	}
	sort.Strings(titles)
	ui.tabs = &Tabs{Titles: titles}

	ui.refresh(true)

	ui.output = NewOutput()

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
	selected := ui.selected[ui.tabs.Selected()]
	for i, d := range disks {
		_, marked := selected[i]
		items[i] = item{d, ui.verbose, marked}
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
		switch {
		case key.Matches(msg, listKeyMap.Quit):
			return ui, tea.Quit

		case key.Matches(msg, listKeyMap.Right):
			ui.tabs.Next()
			ui.refresh(true)
			return ui, nil

		case key.Matches(msg, listKeyMap.Left):
			ui.tabs.Prev()
			ui.refresh(true)
			return ui, nil

		case key.Matches(msg, listKeyMap.Mark):
			selected := ui.selected[ui.tabs.Selected()]
			idx := ui.list.Index()
			if _, marked := selected[idx]; marked {
				delete(selected, idx)
			} else {
				selected[idx] = struct{}{}
			}
			ui.refresh(false)
			ui.list.CursorDown()
			ui.refreshSidebar(ui.disks[ui.tabs.Selected()][idx])

		case key.Matches(msg, listKeyMap.Exec):
			var disks unused.Disks
			for p, sel := range ui.selected {
				for idx := range sel {
					disks = append(disks, ui.disks[p][idx])
				}
			}

			if len(disks) > 0 {
				disks.Sort(unused.ByName)
				ui.output.disks = disks
				return ui.output, nil
			}

			return ui, nil

		case key.Matches(msg, listKeyMap.Verbose):
			ui.verbose = !ui.verbose
			ui.list.SetDelegate(ui.listDelegate())
			ui.refresh(false)

		case key.Matches(msg, listKeyMap.Up, listKeyMap.Down, listKeyMap.PageUp, listKeyMap.PageDown):
			var cmd tea.Cmd
			ui.list, cmd = ui.list.Update(msg)
			return ui, cmd
		}

	case tea.WindowSizeMsg:
		h := msg.Height - lipgloss.Height(ui.tabs.View()) - lipgloss.Height(ui.help.View()) - 2
		w := (msg.Width / 2)
		ui.lbox.Width(w)
		ui.list.SetSize(w, h)
		ui.sidebar.SetSize(w, h)
		ui.output.SetSize(msg.Width, msg.Height)
	}

	return ui, nil
}

func (ui *ui) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		ui.tabs.View(),
		lipgloss.JoinHorizontal(lipgloss.Top,
			ui.lbox.Render(ui.list.View()),
			ui.sidebar.View(),
		),
		ui.help.View(),
	)
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
