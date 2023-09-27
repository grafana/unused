package interactive

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/grafana/unused"
)

type providerItem struct {
	unused.Provider
}

func (i providerItem) FilterValue() string {
	return i.Provider.Name() + " " + i.Provider.Meta().String()
}

func (i providerItem) Title() string {
	return i.Provider.Name()
}

func (i providerItem) Description() string {
	return i.Provider.Meta().String()
}

func newProviderList(providers []unused.Provider) list.Model {
	items := make([]list.Item, len(providers))
	for i, p := range providers {
		items[i] = providerItem{p}
	}

	m := list.New(items, list.NewDefaultDelegate(), 0, 0)
	m.Title = "Please select which provider to use for checking unused disks"
	m.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select provider"),
		)}
	}
	m.AdditionalShortHelpKeys = m.AdditionalFullHelpKeys
	m.SetFilteringEnabled(false)
	m.SetShowHelp(false)
	m.DisableQuitKeybindings()

	return m
}

type providerListModel struct {
	list list.Model
	help help.Model
	sel  key.Binding
	w, h int
}

func newProviderListModel(providers []unused.Provider) providerListModel {
	return providerListModel{
		list: newProviderList(providers),
		help: newHelp(),
		sel:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select provider")),
	}
}

func (m providerListModel) Update(msg tea.Msg) (providerListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.sel):
			return m, sendMsg(m.list.SelectedItem().(providerItem).Provider)

		case msg.String() == "?":
			m.help.ShowAll = !m.help.ShowAll
			m.resetSize()
			return m, nil

		case key.Matches(msg, navKeys.Quit):
			return m, tea.Quit

		default:
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m providerListModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.list.View(), m.help.View(m))
}

func (m *providerListModel) resetSize() {
	hh := lipgloss.Height(m.help.View(m))
	m.list.SetSize(m.w, m.h-hh)
	m.help.Width = m.w
}

func (m *providerListModel) SetSize(w, h int) {
	m.w, m.h = w, h
	m.resetSize()
}

func (m providerListModel) ShortHelp() []key.Binding {
	return []key.Binding{navKeys.Quit, m.sel, navKeys.Up, navKeys.Down}
}

func (m providerListModel) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		m.ShortHelp(),
		{navKeys.PageUp, navKeys.PageDown, navKeys.Home, navKeys.End},
	}
}
