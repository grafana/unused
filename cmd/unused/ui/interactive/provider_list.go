package interactive

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
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

	return m
}
