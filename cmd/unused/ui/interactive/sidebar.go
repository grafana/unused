package interactive

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Sidebar struct {
	strings.Builder

	Style       lipgloss.Style
	HeaderStyle lipgloss.Style
}

func NewSidebar() *Sidebar {
	return &Sidebar{
		HeaderStyle: lipgloss.NewStyle().Bold(true),
		Style:       lipgloss.NewStyle(),
	}
}

func (sb *Sidebar) SetSize(w, h int) {
	sb.Style.Width(w)
	sb.Style.Height(h)
}

func (sb *Sidebar) View() string {
	return sb.Style.Render(sb.Builder.String())
}

func (sb *Sidebar) WriteHeader(s string) {
	sb.WriteString(sb.HeaderStyle.Render(s))
}

func (sb *Sidebar) Printf(format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(&sb.Builder, format, a...)
}

func (sb *Sidebar) Println(a ...interface{}) (int, error) {
	return fmt.Fprintln(&sb.Builder, a...)
}
