package interactive

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/grafana/unused"
)

var _ tea.Model = &model{}

type model struct {
	list    list.Model
	lbox    lipgloss.Style
	tabs    *Tabs
	sidebar viewport.Model
	output  *output
	help    helpview

	selected map[string]map[int]struct{}
	disks    map[string]unused.Disks
	verbose  bool
}

func NewModel(verbose bool, disks unused.Disks) *model {
	m := &model{
		verbose:  verbose,
		selected: make(map[string]map[int]struct{}),
		list:     list.New(nil, listDelegate(verbose), 0, 0),
		lbox:     activeSectionStyle,
		disks:    make(map[string]unused.Disks),
		sidebar:  viewport.New(0, 15),

		help: NewHelp(listKeyMap.Mark, listKeyMap.Exec, listKeyMap.Quit,
			listKeyMap.Up, listKeyMap.Down, listKeyMap.PageUp, listKeyMap.PageDown,
			listKeyMap.Left, listKeyMap.Right, listKeyMap.Verbose),
	}

	m.list.SetShowTitle(false)
	m.list.SetShowHelp(false)
	m.list.SetShowStatusBar(true)
	m.list.SetShowFilter(false)
	m.list.DisableQuitKeybindings()

	m.sidebar.Style = sectionStyle

	for _, d := range disks {
		p := d.Provider().Name()
		m.disks[p] = append(m.disks[p], d)
	}
	titles := make([]string, 0, len(m.disks))
	for p := range m.disks {
		m.selected[p] = make(map[int]struct{})
		titles = append(titles, p)
	}
	sort.Strings(titles)
	m.tabs = &Tabs{Titles: titles}

	m.output = NewOutput()

	return m
}

func (m *model) Init() tea.Cmd {
	cmds := []tea.Cmd{tea.EnterAltScreen, refreshList(true)}

	var disk unused.Disk
	if disks := m.disks[m.tabs.Selected()]; len(disks) > 0 {
		disk = disks[0]
		cmds = append(cmds, displayDiskDetails(disk))
	}

	return tea.Batch(cmds...)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, listKeyMap.Quit):
			return m, tea.Quit

		case key.Matches(msg, listKeyMap.Right):
			m.tabs.Next()
			return m, refreshList(true)

		case key.Matches(msg, listKeyMap.Left):
			m.tabs.Prev()
			return m, refreshList(true)

		case key.Matches(msg, listKeyMap.Mark):
			selected := m.selected[m.tabs.Selected()]
			idx := m.list.Index()
			if _, marked := selected[idx]; marked {
				delete(selected, idx)
			} else {
				selected[idx] = struct{}{}
			}

			m.list.CursorDown()

			cmds := []tea.Cmd{
				displayDiskDetails(m.disks[m.tabs.Selected()][idx]),
				refreshList(false),
			}
			return m, tea.Batch(cmds...)

		case key.Matches(msg, listKeyMap.Exec):
			var disks unused.Disks
			for p, sel := range m.selected {
				for idx := range sel {
					disks = append(disks, m.disks[p][idx])
				}
			}

			if len(disks) > 0 {
				disks.Sort(unused.ByName)
				m.output.disks = disks
				return m.output, nil
			}

			return m, nil

		case key.Matches(msg, listKeyMap.Verbose):
			m.verbose = !m.verbose
			return m, refreshList(false)

		case key.Matches(msg, listKeyMap.Up, listKeyMap.Down, listKeyMap.PageUp, listKeyMap.PageDown):
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

	case unused.Disk:
		var s strings.Builder

		diskDetails.ExecuteTemplate(&s, "disk", msg) // TODO don't ignore error

		m.sidebar.SetContent(lipgloss.JoinVertical(
			lipgloss.Left,
			s.String(),
			strings.Repeat(" ", m.sidebar.Width),
		))

	case refreshListMsg:
		m.list.SetDelegate(listDelegate(m.verbose))
		disks := m.disks[m.tabs.Selected()]
		items := make([]list.Item, len(disks))
		selected := m.selected[m.tabs.Selected()]
		for i, d := range disks {
			_, marked := selected[i]
			items[i] = item{d, m.verbose, marked}
		}
		m.list.SetItems(items)

		if msg.resetSelected {
			m.list.ResetSelected()
		}

	case tea.WindowSizeMsg:
		m.sidebar.Width = msg.Width - 2
		m.sidebar.SetContent(strings.Repeat(" ", m.sidebar.Width))

		h := msg.Height - lipgloss.Height(m.tabs.View()) - lipgloss.Height(m.help.View()) - lipgloss.Height(m.sidebar.View()) - 2
		m.lbox.Width(msg.Width - 2)
		m.list.SetSize(msg.Width-2, h)

		m.output.SetSize(msg.Width, msg.Height)
	}

	return m, nil
}

func (m *model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.tabs.View(),
		m.lbox.Render(m.list.View()),
		m.sidebar.View(),
		m.help.View(),
	)
}

type refreshListMsg struct {
	resetSelected bool
}

func refreshList(reset bool) tea.Cmd {
	return func() tea.Msg {
		return refreshListMsg{
			resetSelected: reset,
		}
	}
}

func displayDiskDetails(disk unused.Disk) tea.Cmd {
	return func() tea.Msg {
		return disk
	}
}
