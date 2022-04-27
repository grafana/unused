package interactive

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

type helpview struct {
	model help.Model
	keys  []key.Binding
}

func NewHelp(keys ...key.Binding) helpview {
	return helpview{
		model: help.New(),
		keys:  keys,
	}
}

func (h helpview) View() string {
	return h.model.ShortHelpView(h.keys)
}
