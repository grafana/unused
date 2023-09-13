package interactive

import "github.com/charmbracelet/bubbles/key"

type customKeyMap struct {
	Back, Quit       key.Binding
	Up, Down         key.Binding
	PageUp, PageDown key.Binding
	Home, End        key.Binding
	Select           key.Binding
	Toggle           key.Binding
	Delete           key.Binding
}

func (km customKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		km.Back, km.Up, km.Down, km.PageUp, km.PageDown,
		km.Select, km.Toggle, km.Delete, km.Quit,
	}
}

func (km customKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.Up, km.Down, km.PageUp, km.PageDown, km.PageUp, km.PageDown},
		{km.Select, km.Toggle, km.Delete, km.Back, km.Quit},
	}
}

var keyMap = customKeyMap{
	Back:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("<esc>", "back")),
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Up:       key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
	Down:     key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
	PageUp:   key.NewBinding(key.WithKeys("pgup", "right"), key.WithHelp("→", "page up")),
	PageDown: key.NewBinding(key.WithKeys("pgdown", "left"), key.WithHelp("←", "page down")),
	Select:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("⮐", "select provider")),
	Toggle:   key.NewBinding(key.WithKeys(" "), key.WithHelp("<space>", "toggle mark")),
	Delete:   key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "delete marked")),
}

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
