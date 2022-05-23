package interactive

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/grafana/unused"
	"github.com/inkel/gotui/tabs"
)

var _ tea.Model = &model{}

type providerTab struct {
	provider unused.Provider
	disks    unused.Disks
}

func (t providerTab) Title() string     { return t.provider.Name() }
func (t providerTab) Data() interface{} { return t.disks }

type model struct {
	list    list.Model
	lbox    lipgloss.Style
	tabs    tabs.Model
	sidebar viewport.Model
	output  *output
	help    helpview

	extraCols []string

	selected map[string]unused.Disk
	verbose  bool
}

func NewModel(verbose bool, disks unused.Disks, extraColumns []string) *model {
	byProvider := make(map[string]*providerTab)
	var providerTabs []tabs.Tab

	for _, disk := range disks {
		p := disk.Provider()
		t, ok := byProvider[p.Name()]
		if !ok {
			t = &providerTab{provider: p}
			byProvider[p.Name()] = t
			providerTabs = append(providerTabs, t)
		}
		t.disks = append(t.disks, disk)
	}

	sort.Slice(providerTabs, func(i, j int) bool { return providerTabs[i].Title() < providerTabs[j].Title() })

	m := &model{
		verbose:  verbose,
		selected: make(map[string]unused.Disk, len(disks)),
		list:     list.New(nil, listDelegate(verbose), 0, 0),
		lbox:     activeSectionStyle,
		sidebar:  viewport.New(0, 15),
		tabs:     tabs.New(providerTabs...),

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

	m.output = NewOutput()

	return m
}

func (m *model) Init() tea.Cmd {
	cmds := []tea.Cmd{tea.EnterAltScreen, m.tabs.TabSelected(), refreshList(true)}

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
		for i, d := range disks {
			_, marked := m.selected[d.ID()]
			items[i] = item{d, m.verbose, marked, m.extraCols}
		}
		m.list.SetItems(items)

		if msg {
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

	case key.Matches(msg, listKeyMap.Mark):
		disk := m.list.SelectedItem().(item).disk
		idx := disk.ID()
		if _, marked := m.selected[idx]; marked {
			delete(m.selected, idx)
		} else {
			m.selected[idx] = disk
		}

		m.list.CursorDown()

		cmds := []tea.Cmd{
			displayDiskDetails(disk),
			refreshList(false),
		}
		return m, tea.Batch(cmds...)

	case key.Matches(msg, listKeyMap.Exec):
		disks := make(unused.Disks, 0, len(m.selected))
		for _, d := range m.selected {
			disks = append(disks, d)
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

type refreshListMsg bool

func refreshList(reset bool) tea.Cmd {
	return func() tea.Msg {
		return refreshListMsg(reset)
	}
}

func displayDiskDetails(disk unused.Disk) tea.Cmd {
	return func() tea.Msg {
		return disk
	}
}

// NEW MODEL
type state int

const (
	stateProviderList state = iota
	stateProviderView
	stateFetchingDisks
	stateDeletingDisks
)

var _ tea.Model = Model{}

type Model struct {
	providerList list.Model
	providerView table.Model
	provider     unused.Provider
	spinner      spinner.Model
	disks        map[unused.Provider]unused.Disks
	state        state
	loadingDone  chan struct{}
	extraCols    []string
	key, value   string
	output       viewport.Model
	deleteStatus map[string]*deleteStatus
}

type deleteStatus struct {
	cur  bool
	done bool
	err  error
}

func New(providers []unused.Provider, extraColumns []string, key, value string) Model {
	return Model{
		providerList: newProviderList(providers),
		providerView: newProviderView(extraColumns),
		disks:        make(map[unused.Provider]unused.Disks),
		state:        stateProviderList,
		spinner:      spinner.New(),
		loadingDone:  make(chan struct{}),
		extraCols:    extraColumns,
		key:          key,
		value:        value,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit

		case "esc":
			if m.state == stateProviderView {
				m.state = stateProviderList
			}

			return m, nil

		case "enter":
			if m.state == stateProviderList {
				m.provider = m.providerList.SelectedItem().(providerItem).Provider
				m.state = stateFetchingDisks

				go m.loadDisks()

				return m, spinner.Tick
			}

		case "x":
			if m.state == stateProviderView {
				if rows := m.providerView.SelectedRows(); len(rows) > 0 {
					m.state = stateDeletingDisks
					m.deleteStatus = make(map[string]*deleteStatus, len(rows))

					go m.deleteDisks()

					return m, spinner.Tick
				}
			}
		}

	case spinner.TickMsg:
		select {
		case <-m.loadingDone:
			m.state = stateProviderView
			m.providerView = m.providerView.WithRows(disksToRows(m.disks[m.provider], m.extraCols))
		default:
		}

	case tea.WindowSizeMsg:
		m.providerList.SetSize(msg.Width, msg.Height)
		m.providerView = m.providerView.WithTargetWidth(msg.Width).WithPageSize(msg.Height - 6)
		m.output.Width = msg.Width
		m.output.Height = msg.Height - 1
	}

	var cmd tea.Cmd

	switch m.state {
	case stateFetchingDisks:
		m.spinner, cmd = m.spinner.Update(msg)

	case stateProviderList:
		m.providerList, cmd = m.providerList.Update(msg)

	case stateProviderView:
		m.providerView, cmd = m.providerView.Update(msg)

	case stateDeletingDisks:
		m.spinner, cmd = m.spinner.Update(msg)
	}

	return m, cmd
}

var errorStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#cb4b16", Dark: "#d87979"})

func (m Model) View() string {
	switch m.state {
	case stateProviderList:
		return m.providerList.View()

	case stateProviderView:
		return m.providerView.View()

	case stateFetchingDisks:
		return fmt.Sprintf("Fetching disks for %s %s %s\n", m.provider.Name(), m.provider.Meta().String(), m.spinner.View())

	case stateDeletingDisks:
		rows := m.providerView.SelectedRows()
		title := fmt.Sprintf("Deleting %d disks for %s %s\n", len(rows), m.provider.Name(), m.provider.Meta().String())

		var sb strings.Builder

		for _, r := range rows {
			d := r.Data[columnDisk].(unused.Disk)
			s := m.deleteStatus[d.ID()]
			if s != nil {
				if s.cur {
					fmt.Fprintf(&sb, "➤ %s %s\n", d.Name(), m.spinner.View())
				} else if s.done {
					if s.err != nil {
						fmt.Fprintf(&sb, "𐄂 %s\n  %s\n", d.Name(), errorStyle.Render(s.err.Error()))
					} else {
						fmt.Fprintf(&sb, "✓ %s\n", d.Name())
					}
				}
			} else {
				fmt.Fprintf(&sb, "  %s\n", d.Name())
			}
		}

		return lipgloss.JoinVertical(lipgloss.Left, title, sb.String())

	default:
		return "WHAT"
	}
}

func (m Model) loadDisks() {
	m.state = stateProviderView

	disks, ok := m.disks[m.provider]
	if !ok {
		disks, _ = m.provider.ListUnusedDisks(context.TODO()) // TODO handle error

		if m.key != "" {
			filtered := make(unused.Disks, 0, len(disks))
			for _, d := range disks {
				if d.Meta().Matches(m.key, m.value) {
					filtered = append(filtered, d)
				}
			}
			disks = filtered
		}

		m.disks[m.provider] = disks
	}

	m.loadingDone <- struct{}{}
}

func (m Model) deleteDisks() {
	for _, r := range m.providerView.SelectedRows() {
		d := r.Data[columnDisk].(unused.Disk)
		s := &deleteStatus{cur: true}
		m.deleteStatus[d.ID()] = s

		s.err = d.Provider().Delete(context.TODO(), d)

		s.done = true
		s.cur = false
	}
}
