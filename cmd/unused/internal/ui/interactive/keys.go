package interactive

import "github.com/charmbracelet/bubbles/key"

var navKeys = struct {
	Quit, Up, Down, PageUp, PageDown, Home, End, Back key.Binding
}{
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Up:       key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
	Down:     key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
	PageUp:   key.NewBinding(key.WithKeys("pgup", "right"), key.WithHelp("→", "page up")),
	PageDown: key.NewBinding(key.WithKeys("pgdown", "left"), key.WithHelp("←", "page down")),
	Home:     key.NewBinding(key.WithKeys("home"), key.WithHelp("home", "first")),
	End:      key.NewBinding(key.WithKeys("end"), key.WithHelp("end", "last")),
	Back:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("<esc>", "back")),
}
