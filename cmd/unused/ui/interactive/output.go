package interactive

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"text/template"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/grafana/unused"
)

type output struct {
	disks     unused.Disks
	viewport  viewport.Model
	w, h      int
	delete    bool
	deletidx  int
	delstatus map[int]error
	spinner   spinner.Model
	help      helpview
	tpl       *template.Template

	ctx    context.Context
	cancel context.CancelFunc
}

func NewOutput() *output {
	o := &output{
		delstatus: make(map[int]error),
		help:      NewHelp(outputKeyMap.Exec, outputKeyMap.Quit, outputKeyMap.Up, outputKeyMap.Down, outputKeyMap.PageUp, outputKeyMap.PageDown),
		spinner:   spinner.New(),
	}
	o.viewport.Style = o.viewport.Style.Border(lipgloss.RoundedBorder())
	o.spinner.Spinner = spinner.Dot

	o.tpl = template.Must(template.New("").
		Funcs(template.FuncMap{
			"error": errStyle.Render,
		}).
		Parse(outputTpl))

	return o
}

func (o *output) SetSize(w, h int) {
	o.w, o.h = w, h
}

func (o *output) Init() tea.Cmd { return nil }

func (o *output) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		o.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		return o.updateKeyMsg(msg)
	}

	var cmd [2]tea.Cmd
	o.viewport, cmd[0] = o.viewport.Update(msg)
	o.spinner, cmd[1] = o.spinner.Update(msg)
	return o, tea.Batch(cmd[0], cmd[1])
}

func (o *output) updateKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, outputKeyMap.Quit):
		return o, tea.Quit

	case key.Matches(msg, outputKeyMap.Cancel):
		if o.delete {
			o.cancel()
			o.delete = false
		}

		return o, nil

	case key.Matches(msg, outputKeyMap.Up, outputKeyMap.Down, outputKeyMap.PageUp, outputKeyMap.PageDown):
		var cmd tea.Cmd
		o.viewport, cmd = o.viewport.Update(msg)
		return o, cmd

	case key.Matches(msg, outputKeyMap.Exec):
		o.delete = true
		outputKeyMap.Quit.SetEnabled(false)
		o.ctx, o.cancel = context.WithCancel(context.Background())

		// TODO extract this to a method, and implement real deletion
		go func() {
			defer outputKeyMap.Quit.SetEnabled(true)
			defer o.cancel()

			for i := range o.disks {
				_, ok := o.delstatus[i]
				if ok {
					// disk was already processed
					continue
				}

				select {
				case <-o.ctx.Done():
					return
				default:
					o.deletidx = i
					time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
					if rand.Intn(100)%3 == 0 {
						o.delstatus[i] = errors.New("something went wrong")
					} else {
						o.delstatus[i] = nil
					}
				}
			}
		}()

		return o, o.spinner.Tick

	default:
		return o, nil
	}
}

type templateData struct {
	Disk     unused.Disk
	Done     bool
	Error    error
	Deleting bool
}

func (o *output) View() string {
	var (
		count = fmt.Sprintf("%d disks marked for deletion", len(o.disks))
		fill  = strings.Repeat(" ", o.w-lipgloss.Width(count)-2)
		title = titleStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, count, fill))
		help  = o.help.View()
	)

	o.viewport.Height = o.h - lipgloss.Height(title) - lipgloss.Height(help) - 2

	data := make([]templateData, len(o.disks))
	for i, d := range o.disks {
		d := templateData{
			Disk:     d,
			Deleting: o.delete && i == o.deletidx,
		}
		d.Error, d.Done = o.delstatus[i]
		data[i] = d
	}

	var s strings.Builder
	err := o.tpl.Execute(&s, data)
	if err == nil {
		// fill
		s.WriteString(strings.Repeat(" ", o.w-2))
		o.viewport.SetContent(s.String())
	} else {
		o.viewport.SetContent(errStyle.Render(err.Error()))
	}

	help = centerStyle.Copy().Width(o.w).Render(help)

	return lipgloss.JoinVertical(lipgloss.Left, title, o.viewport.View(), help)
}
