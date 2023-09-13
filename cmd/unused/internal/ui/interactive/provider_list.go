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

type providerListKeymap struct {
	Quit, Select     key.Binding
	Up, Down         key.Binding
	PageUp, PageDown key.Binding
	Home, End        key.Binding
}

func (km providerListKeymap) ShortHelp() []key.Binding {
	return []key.Binding{km.Quit, km.Select, km.Up, km.Down}
}

func (km providerListKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		km.ShortHelp(),
		{km.PageUp, km.PageDown, km.Home, km.End},
	}
}

type providerListModel struct {
	list list.Model
	help help.Model
	km   providerListKeymap
	w, h int
}

func newProviderListModel(providers []unused.Provider) providerListModel {
	return providerListModel{
		list: newProviderList(providers),
		help: newHelp(),
		km: providerListKeymap{
			Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
			Select:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select provider")),
			Up:       key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
			Down:     key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
			PageUp:   key.NewBinding(key.WithKeys("pgup", "right"), key.WithHelp("→", "page up")),
			PageDown: key.NewBinding(key.WithKeys("pgdown", "left"), key.WithHelp("←", "page down")),
			Home:     key.NewBinding(key.WithKeys("home"), key.WithHelp("home", "first")),
			End:      key.NewBinding(key.WithKeys("end"), key.WithHelp("end", "last")),
		},
	}
}

func (m providerListModel) Update(msg tea.Msg) (providerListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.km.Select):
			return m, func() tea.Msg { return m.list.SelectedItem().(providerItem).Provider }

		case msg.String() == "?":
			m.help.ShowAll = !m.help.ShowAll
			m.resetSize()
			return m, nil

		case key.Matches(msg, m.km.Quit):
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
	return lipgloss.JoinVertical(lipgloss.Left, m.list.View(), m.help.View(m.km))
}

func (m *providerListModel) resetSize() {
	hh := lipgloss.Height(m.help.View(m.km))
	m.list.SetSize(m.w, m.h-hh)
	m.help.Width = m.w
}

func (m *providerListModel) SetSize(w, h int) {
	m.w, m.h = w, h
	m.resetSize()
}
