package interactive

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/grafana/unused"
)

type state int

const (
	stateProviderList state = iota
	stateProviderView
	stateFetchingDisks
	stateDeletingDisks
)

var _ tea.Model = Model{}

type Model struct {
	providerList providerListModel
	providerView table.Model
	provider     unused.Provider
	spinner      spinner.Model
	disks        map[unused.Provider]unused.Disks
	state        state
	extraCols    []string
	key, value   string
	output       viewport.Model
	help         help.Model
}

func New(providers []unused.Provider, extraColumns []string, key, value string) Model {
	return Model{
		providerList: newProviderListModel(providers),
		providerView: newProviderView(extraColumns),
		disks:        make(map[unused.Provider]unused.Disks),
		state:        stateProviderList,
		spinner:      spinner.New(),
		extraCols:    extraColumns,
		key:          key,
		value:        value,
		help:         newHelp(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keyMap.Quit):
			return m, tea.Quit

		case key.Matches(msg, keyMap.Back):
			switch m.state {
			case stateProviderView:
				m.state = stateProviderList
				return m, nil

			case stateDeletingDisks:
				delete(m.disks, m.provider)
				m.state = stateFetchingDisks
				return m, tea.Batch(spinner.Tick, loadDisks(m.provider, m.disks, m.key, m.value))
			}

			return m, nil

		case key.Matches(msg, keyMap.Delete):
			if m.state == stateProviderView {
				if rows := m.providerView.SelectedRows(); len(rows) > 0 {
					s := deleteProgress{
						disks:  make(unused.Disks, 0, len(rows)),
						status: make([]*deleteStatus, len(rows)),
					}
					for _, r := range rows {
						s.disks = append(s.disks, r.Data[columnDisk].(unused.Disk))
					}

					m.state = stateDeletingDisks
					return m, tea.Batch(spinner.Tick, deleteCurrent(s))
				}
			}
		}

	case unused.Provider:
		if m.state == stateProviderList {
			m.provider = msg
			m.providerView = m.providerView.WithRows(nil)
			m.state = stateFetchingDisks

			return m, tea.Batch(spinner.Tick, loadDisks(m.provider, m.disks, m.key, m.value))
		}

	case loadedDisks:
		// TODO handle error
		m.providerView = m.providerView.WithRows(disksToRows(msg.disks, m.extraCols))
		m.state = stateProviderView

	case deleteProgress:
		sb := &strings.Builder{}

		fmt.Fprintf(sb, "Deleting %d disks from %s %s\n\n", len(msg.disks), m.provider.Name(), m.provider.Meta().String())

		for i, d := range msg.disks {
			s := msg.status[i]

			switch {
			case s == nil:
				fmt.Fprintf(sb, "  %s\n", d.Name())

			case msg.cur == i:
				fmt.Fprintf(sb, "➤ %s %s\n", d.Name(), m.spinner.View())

			case !s.done:

			case s.err != nil:
				fmt.Fprintf(sb, "𐄂 %s\n  %s\n", d.Name(), errorStyle.Render(s.err.Error()))

			default:
				fmt.Fprintf(sb, "✓ %s\n", d.Name())
			}
		}

		m.output.SetContent(sb.String())

		return m, deleteCurrent(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		helpHeight := lipgloss.Height(m.getHelp())
		m.providerList.SetSize(msg.Width, msg.Height)
		m.providerView = m.providerView.WithTargetWidth(msg.Width).WithPageSize(msg.Height - 4 - helpHeight)
		m.output.Width = msg.Width
		m.output.Height = msg.Height - 1 - helpHeight
	}

	var cmd tea.Cmd

	switch m.state {
	case stateProviderList:
		m.providerList, cmd = m.providerList.Update(msg)

	case stateProviderView:
		m.providerView, cmd = m.providerView.Update(msg)
	}

	return m, cmd
}

var errorStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#cb4b16", Dark: "#d87979"})

func (m Model) getHelp() string {
	return m.help.View(keyMap)
}

func (m Model) View() string {
	var view string
	switch m.state {
	case stateProviderList:
		return m.providerList.View()

	case stateProviderView:
		view = m.providerView.View()

	case stateFetchingDisks:
		view = fmt.Sprintf("Fetching disks for %s %s %s\n", m.provider.Name(), m.provider.Meta().String(), m.spinner.View())

	case stateDeletingDisks:
		view = m.output.View()

	default:
		return "WHAT"
	}

	return lipgloss.JoinVertical(lipgloss.Left, view, m.getHelp())
}

type loadedDisks struct {
	disks unused.Disks
	err   error
}

func loadDisks(provider unused.Provider, cache map[unused.Provider]unused.Disks, key, value string) tea.Cmd {
	return func() tea.Msg {
		if disks, ok := cache[provider]; ok {
			return loadedDisks{disks, nil}
		}

		disks, err := provider.ListUnusedDisks(context.TODO())
		if err != nil {
			return loadedDisks{nil, err}
		}

		if key != "" {
			filtered := make(unused.Disks, 0, len(disks))
			for _, d := range disks {
				if d.Meta().Matches(key, value) {
					filtered = append(filtered, d)
				}
			}
			disks = filtered
		}

		cache[provider] = disks

		return loadedDisks{disks, nil}
	}
}

type deleteStatus struct {
	done bool
	err  error
}

type deleteProgress struct {
	cur    int
	disks  unused.Disks
	status []*deleteStatus
}

func deleteCurrent(p deleteProgress) tea.Cmd {
	if p.cur == len(p.disks) {
		return nil
	}

	if p.status[p.cur] == nil {
		ds := &deleteStatus{}
		p.status[p.cur] = ds

		go func() {
			d := p.disks[p.cur]
			ds.err = d.Provider().Delete(context.TODO(), d)
			ds.done = true
		}()
	} else if p.status[p.cur].done {
		p.cur++
	}

	return func() tea.Msg { return p }
}

func newHelp() help.Model {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#909090",
		Dark:  "#FFFF00",
	})

	descStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#B2B2B2",
		Dark:  "#999999",
	})

	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#DDDADA",
		Dark:  "#3C3C3C",
	})

	m := help.New()
	m.Styles = help.Styles{
		ShortKey:       keyStyle,
		ShortDesc:      descStyle,
		ShortSeparator: sepStyle,
		Ellipsis:       sepStyle.Copy(),
		FullKey:        keyStyle.Copy(),
		FullDesc:       descStyle.Copy(),
		FullSeparator:  sepStyle.Copy(),
	}

	return m
}
