package interactive

import "github.com/charmbracelet/bubbles/key"

var listKeyMap = struct {
	Mark, Exec, Quit, Up, Down, PageUp, PageDown, Right, Left, Verbose key.Binding
}{
	Mark:     key.NewBinding(key.WithKeys("m", " "), key.WithHelp("m", "toggle mark")),
	Exec:     key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "delete")),
	Quit:     key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	Up:       key.NewBinding(key.WithKeys("up"), key.WithHelp("up", "up one line")),
	Down:     key.NewBinding(key.WithKeys("down"), key.WithHelp("down", "down one line")),
	Right:    key.NewBinding(key.WithKeys("right"), key.WithHelp("→", "next provider")),
	Left:     key.NewBinding(key.WithKeys("left"), key.WithHelp("←", "previous provider")),
	Verbose:  key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "toggle verbose")),
	PageUp:   key.NewBinding(key.WithKeys("pgup"), key.WithHelp("page up", "up one page")),
	PageDown: key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("page down", "down one page")),
}
