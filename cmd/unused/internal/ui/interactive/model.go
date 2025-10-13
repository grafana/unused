package interactive

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/grafana/unused"
)

const (
	minWidth  = 100
	minHeight = 30

	timeout = 1 * time.Minute
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
	help         help.Model
	provider     unused.Provider
	err          error
	cache        map[unused.Provider]unused.Disks
	filter       unused.FilterFunc
	extraCols    []string
	spinner      spinner.Model
	providerList providerListModel
	providerView providerViewModel
	deleteView   deleteViewModel
	state        state
	w, h         int
}

func New(providers []unused.Provider, extraColumns []string, filter unused.FilterFunc, dryRun bool) Model {
	m := Model{
		providerList: newProviderListModel(providers),
		providerView: newProviderViewModel(extraColumns),
		deleteView:   newDeleteViewModel(dryRun),
		cache:        make(map[unused.Provider]unused.Disks),
		state:        stateProviderList,
		spinner:      spinner.New(),
		extraCols:    extraColumns,
		filter:       filter,
		help:         newHelp(),
	}

	if len(providers) == 1 {
		m.provider = providers[0]
	}

	return m
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{tea.EnterAltScreen}
	if m.provider != nil { // No need to show the providers list if there's only one provider
		cmds = append(cmds, sendMsg(m.provider))
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, navKeys.Quit):
			return m, tea.Quit

		case key.Matches(msg, navKeys.Back):
			switch m.state {
			case stateProviderView:
				m.state = stateProviderList
				return m, nil

			case stateDeletingDisks:
				delete(m.cache, m.provider)
				m.state = stateFetchingDisks
				m.providerView = m.providerView.Empty()
				return m, tea.Batch(m.spinner.Tick, m.loadDisks())
			}

			return m, nil
		}

	case unused.Provider:
		if m.state == stateProviderList {
			m.provider = msg
			m.providerView = m.providerView.Empty()
			m.state = stateFetchingDisks

			return m, tea.Batch(m.spinner.Tick, m.loadDisks())
		}

	case unused.Disks:
		switch m.state {
		case stateFetchingDisks:
			m.cache[m.provider] = msg
			m.providerView = m.providerView.WithDisks(msg)
			m.state = stateProviderView

		case stateProviderView:
			m.deleteView = m.deleteView.WithDisks(m.provider, msg)
			m.state = stateDeletingDisks
		}

	case refreshMsg:
		delete(m.cache, m.provider)
		m.state = stateFetchingDisks
		return m, tea.Batch(m.spinner.Tick, m.loadDisks())

	case spinner.TickMsg:
		if m.state == stateFetchingDisks {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.providerList.SetSize(msg.Width, msg.Height)
		m.providerView.SetSize(msg.Width, msg.Height)
		m.deleteView.SetSize(msg.Width, msg.Height)

	case error:
		m.err = msg
		return m, nil
	}

	var cmd tea.Cmd

	switch m.state {
	case stateProviderList:
		m.providerList, cmd = m.providerList.Update(msg)

	case stateProviderView:
		m.providerView, cmd = m.providerView.Update(msg)

	case stateDeletingDisks:
		m.deleteView, cmd = m.deleteView.Update(msg)
	}

	return m, cmd
}

var errorStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#cb4b16", Dark: "#d87979"})

func (m Model) View() string {
	if m.w < minWidth || m.h < minHeight {
		return errorStyle.Render(fmt.Sprintf("invalid window size %dx%d, expecting at least %dx%d", m.w, m.h, minWidth, minHeight))
	}
	if m.err != nil {
		return errorStyle.Render(m.err.Error())
	}

	switch m.state {
	case stateProviderList:
		return m.providerList.View()

	case stateProviderView:
		return m.providerView.View()

	case stateFetchingDisks:
		return fmt.Sprintf("Fetching disks for %s %s %s\n", m.provider.Name(), m.provider.Meta().String(), m.spinner.View())

	case stateDeletingDisks:
		return m.deleteView.View()

	default:
		return "WHAT"
	}
}

func (m Model) loadDisks() tea.Cmd {
	return func() tea.Msg {
		if disks, ok := m.cache[m.provider]; ok {
			return disks
		}

		ctx, cancel := context.WithTimeout(context.TODO(), timeout)
		defer cancel()

		disks, err := m.provider.ListUnusedDisks(ctx)
		if err != nil {
			return fmt.Errorf("listing unused disks for %s %s: %w", m.provider.Name(), m.provider.Meta(), err)
		}

		return disks.Filter(m.filter)
	}
}

// sendMsg is a tea.Cmd that will send whatever is passed as an
// argument as a tea.Msg.
func sendMsg(msg tea.Msg) tea.Cmd { return func() tea.Msg { return msg } }

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
		Ellipsis:       sepStyle,
		FullKey:        keyStyle,
		FullDesc:       descStyle,
		FullSeparator:  sepStyle,
	}

	return m
}

// refreshMsg is a message used to mark that we need to clear the
// cache for the current provider and reload its unused disks.
type refreshMsg struct{}
