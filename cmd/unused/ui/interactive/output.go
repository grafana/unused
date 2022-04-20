package interactive

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/grafana/unused"
)

var outputKeyMap = struct {
	Exec, Quit, Up, Down, PageUp, PageDown key.Binding
}{
	Exec:     key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "delete")),
	Quit:     key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	Up:       key.NewBinding(key.WithKeys("↑"), key.WithHelp("up", "move up one line")),
	Down:     key.NewBinding(key.WithKeys("↓"), key.WithHelp("down", "move down one line")),
	PageUp:   key.NewBinding(key.WithKeys("pgup"), key.WithHelp("page up", "move up one page")),
	PageDown: key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("page down", "move down one page")),
}

type output struct {
	disks     unused.Disks
	viewport  viewport.Model
	w, h      int
	delete    bool
	deletidx  int
	delstatus map[int]error
	spinner   spinner.Model
	help      helpview
}

func NewOutput() *output {
	o := &output{
		delstatus: make(map[int]error),
		help:      NewHelp(outputKeyMap.Exec, outputKeyMap.Quit, outputKeyMap.Up, outputKeyMap.Down, outputKeyMap.PageUp, outputKeyMap.PageDown),
	}
	o.viewport.Style = o.viewport.Style.Border(lipgloss.RoundedBorder())
	o.spinner.Spinner = spinner.Points
	return o
}

func (o *output) SetSize(w, h int) {
	o.w, o.h = w, h
}

func (o *output) Init() tea.Cmd { return nil }

func (o *output) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		o.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, outputKeyMap.Quit):
			return o, tea.Quit

		case key.Matches(msg, outputKeyMap.Up, outputKeyMap.Down, outputKeyMap.PageUp, outputKeyMap.PageDown):
			o.viewport, cmd = o.viewport.Update(msg)
			return o, cmd

		case key.Matches(msg, outputKeyMap.Exec):
			o.delete = true

			// TODO extract this to a method, and implement real deletion
			go func() {
				for i := range o.disks {
					o.deletidx = i
					time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
					if rand.Intn(100)%3 == 0 {
						o.delstatus[i] = errors.New("something went wrong")
					} else {
						o.delstatus[i] = nil
					}
				}
			}()

			return o, o.spinner.Tick
		}

	case spinner.TickMsg:
		o.spinner, cmd = o.spinner.Update(msg)
		return o, cmd

	default:
		o.viewport, cmd = o.viewport.Update(msg)
		return o, cmd
	}

	return o, nil
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Border(lipgloss.RoundedBorder())
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#990000", Dark: "#ff0000"})
)

func (o *output) renderTitle() string {
	var (
		count = fmt.Sprintf("%d disks marked for deletion", len(o.disks))
		fill  = strings.Repeat(" ", o.w-lipgloss.Width(count)-2)
	)
	return titleStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, count, fill))
}

func (o *output) View() string {
	var (
		count = fmt.Sprintf("%d disks marked for deletion", len(o.disks))
		fill  = strings.Repeat(" ", o.w-lipgloss.Width(count)-2)
		title = titleStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, count, fill))
		help  = o.help.View()
	)

	o.viewport.Height = o.h - lipgloss.Height(title) - lipgloss.Height(help) - 2

	var ds strings.Builder
	renderDisk := func(i int, d unused.Disk) string {
		ds.Reset()

		fmt.Fprintf(&ds, "* %s (%s %s) ", d.Name(), d.Provider().Name(), d.Provider().Meta())
		if o.delete {
			if err, ok := o.delstatus[i]; err != nil {
				fmt.Fprintf(&ds, "ERROR\n")
				fmt.Fprintf(&ds, errStyle.Render(err.Error()))
			} else if ok {
				fmt.Fprintf(&ds, "DONE")
			} else if o.deletidx == i {
				fmt.Fprintf(&ds, o.spinner.View())
			}
		}

		return ds.String()
	}

	var s strings.Builder
	for i, d := range o.disks {
		disk := renderDisk(i, d)
		fill = strings.Repeat(" ", o.w-lipgloss.Width(disk)-2)
		s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, disk, fill))
		s.WriteRune('\n')
	}
	o.viewport.SetContent(s.String())

	fill = strings.Repeat(" ", (o.w-lipgloss.Width(help))/2)
	help = lipgloss.JoinHorizontal(lipgloss.Top, fill, help, fill)

	return lipgloss.JoinVertical(lipgloss.Left, title, o.viewport.View(), help)
}
