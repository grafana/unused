package interactive

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	providerView providerViewModel
	deleteView   deleteViewModel
	provider     unused.Provider
	spinner      spinner.Model
	disks        map[unused.Provider]unused.Disks
	state        state
	extraCols    []string
	key, value   string
	help         help.Model
	err          error
}

func New(providers []unused.Provider, extraColumns []string, key, value string) Model {
	m := Model{
		providerList: newProviderListModel(providers),
		providerView: newProviderViewModel(extraColumns),
		deleteView:   newDeleteViewModel(),
		disks:        make(map[unused.Provider]unused.Disks),
		state:        stateProviderList,
		spinner:      spinner.New(),
		extraCols:    extraColumns,
		key:          key,
		value:        value,
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
		cmds = append(cmds, func() tea.Msg { return m.provider })
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
				delete(m.disks, m.provider)
				m.state = stateFetchingDisks
				return m, tea.Batch(spinner.Tick, loadDisks(m.provider, m.disks, m.key, m.value))
			}

			return m, nil
		}

	case unused.Provider:
		if m.state == stateProviderList {
			m.provider = msg
			m.providerView = m.providerView.Empty()
			m.state = stateFetchingDisks

			return m, tea.Batch(spinner.Tick, loadDisks(m.provider, m.disks, m.key, m.value))
		}

	case unused.Disks:
		switch m.state {
		case stateFetchingDisks:
			m.providerView = m.providerView.WithDisks(msg)
			m.state = stateProviderView

		case stateProviderView:
			m.deleteView = m.deleteView.WithDisks(m.provider, msg)
			m.state = stateDeletingDisks
		}

	case spinner.TickMsg:
		if m.state == stateFetchingDisks {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
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

func loadDisks(provider unused.Provider, cache map[unused.Provider]unused.Disks, key, value string) tea.Cmd {
	return func() tea.Msg {
		if disks, ok := cache[provider]; ok {
			return disks
		}

		disks, err := provider.ListUnusedDisks(context.TODO())
		if err != nil {
			return err
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

		return disks
	}
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
