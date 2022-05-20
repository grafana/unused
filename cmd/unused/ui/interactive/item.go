package interactive

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/grafana/unused"
	"github.com/grafana/unused/cli"
)

type item struct {
	disk    unused.Disk
	verbose bool
	marked  bool

	extraCols []string
}

func (i item) Title() string {
	var s strings.Builder
	s.WriteString(i.disk.Name())

	for _, c := range i.extraCols {
		s.WriteRune(' ')
		s.WriteString(c)
		s.WriteRune('=')
		s.WriteString(i.disk.Meta()[c])
	}

	n := s.String()

	if i.marked {
		return markedStyle.Render(n)
	}
	return n
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
	s.WriteString(cli.Age(i.disk.CreatedAt()))

	s.WriteString(" unused=")
	s.WriteString(cli.Age(i.disk.LastUsedAt()))

	if i.verbose {
		s.WriteRune(' ')
		s.WriteString(i.disk.Meta().String())
	}

	return s.String()
}

func listDelegate(verbose bool) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.ShowDescription = verbose
	d.SetSpacing(0)
	d.UpdateFunc = func(msg tea.Msg, list *list.Model) tea.Cmd {
		item, ok := list.SelectedItem().(item)
		if ok { // this should always happen
			return displayDiskDetails(item.disk)
		}
		return nil
	}
	return d
}
