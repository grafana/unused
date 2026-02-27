package interactive

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	teav1 "github.com/charmbracelet/bubbletea"
)

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

func keyMsgV2toV1(msg tea.KeyMsg) teav1.KeyMsg {
	var msgv1 teav1.KeyMsg
	switch msg.Key().String() {
	case "space":
		msgv1.Type = teav1.KeySpace
	case "up":
		msgv1.Type = teav1.KeyUp
	case "down":
		msgv1.Type = teav1.KeyDown
	case "left":
		msgv1.Type = teav1.KeyLeft
	case "right":
		msgv1.Type = teav1.KeyRight
	}
	return msgv1
}
