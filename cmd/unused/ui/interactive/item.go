package interactive

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/grafana/unused"
)

type item struct {
	disk    unused.Disk
	verbose bool
	marked  bool
}

var marked = lipgloss.NewStyle().Strikethrough(true)

func (i item) Title() string {
	if i.marked {
		return marked.Render(i.disk.Name())
	}
	return i.disk.Name()
}

func (i item) FilterValue() string {
	// TODO this doesn't looks right when filtering
	return i.disk.Name() + " " + i.disk.Meta().String() + " " + i.disk.Provider().Name() + " " + i.disk.Provider().Meta().String()
}

func (i item) Description() string {
	var s strings.Builder

	s.WriteString(i.disk.Provider().Name())
	s.WriteRune('{')
	s.WriteString(i.disk.Provider().Meta().String())
	s.WriteRune('}')

	s.WriteRune(' ')
	s.WriteString("age=")
	s.WriteString(age(i.disk.CreatedAt()))

	if i.verbose {
		s.WriteRune(' ')
		s.WriteString(i.disk.Meta().String())
	}

	return s.String()
}

func listDelegate(verbose bool) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.ShowDescription = verbose
	d.UpdateFunc = func(msg tea.Msg, list *list.Model) tea.Cmd {
		item, ok := list.SelectedItem().(item)
		if ok { // this should always happen
			return displayDiskDetails(item.disk)
		}
		return nil
	}
	return d
}
