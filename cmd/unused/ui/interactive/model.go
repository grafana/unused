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
	"github.com/inkel/gotui/tabs"
)

var _ tea.Model = &model{}

type model struct {
	list    list.Model
	lbox    lipgloss.Style
	tabs    tabs.Model
	sidebar viewport.Model
	output  *output
	help    helpview

	extraCols []string

	selected map[string]map[int]struct{}
	disks    map[string]unused.Disks
	verbose  bool
}

func NewModel(verbose bool, disks unused.Disks, extraColumns []string) *model {
	m := &model{
		verbose:  verbose,
		selected: make(map[string]map[int]struct{}),
		list:     list.New(nil, listDelegate(verbose), 0, 0),
		lbox:     activeSectionStyle,
		disks:    make(map[string]unused.Disks),
		sidebar:  viewport.New(0, 15),

		extraCols: extraColumns,

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

	providerTabs := make([]tabs.Tab, 0, len(m.disks))
	for p, ds := range m.disks {
		m.selected[p] = make(map[int]struct{})
		providerTabs = append(providerTabs, disksTab{p, ds})
	}
	sort.Slice(providerTabs, func(i, j int) bool { return providerTabs[i].Title() < providerTabs[j].Title() })
	m.tabs = tabs.New(providerTabs...)

	m.output = NewOutput()

	return m
}

func (m *model) Init() tea.Cmd {
	cmds := []tea.Cmd{tea.EnterAltScreen, refreshList(true)}

	var disk unused.Disk
	if disks := m.tabs.Selected().Data().(unused.Disks); len(disks) > 0 {
		disk = disks[0]
		cmds = append(cmds, displayDiskDetails(disk))
	}

	return tea.Batch(cmds...)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		return m.updateKeyMsg(msg)

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
		disks := m.tabs.Selected().Data().(unused.Disks)
		items := make([]list.Item, len(disks))
		selected := m.selected[m.tabs.Selected().Title()]
		for i, d := range disks {
			_, marked := selected[i]
			items[i] = item{d, m.verbose, marked, m.extraCols}
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

	case tabs.TabSelectedMsg:
		return m, refreshList(true)
	}

	return m, nil
}

func (m *model) updateKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		selected := m.selected[m.tabs.Selected().Title()]
		idx := m.list.Index()
		if _, marked := selected[idx]; marked {
			delete(selected, idx)
		} else {
			selected[idx] = struct{}{}
		}

		m.list.CursorDown()

		cmds := []tea.Cmd{
			displayDiskDetails(m.disks[m.tabs.Selected().Title()][idx]),
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
			m.output.SetDisks(disks)
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

	default:
		var cmd tea.Cmd
		m.tabs, cmd = m.tabs.Update(msg)
		return m, cmd
	}
}

func (m *model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.tabs.View(),
		m.lbox.Render(m.list.View()),
		m.sidebar.View(),
		centerStyle.Copy().Width(m.lbox.GetWidth()).Render(m.help.View()),
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

type disksTab struct {
	title string
	disks unused.Disks
}

func (t disksTab) Title() string     { return t.title }
func (t disksTab) Data() interface{} { return t.disks }
