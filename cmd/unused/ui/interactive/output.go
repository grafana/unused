package interactive

import (
	"context"
	"fmt"
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

type diskStatus struct {
	Disk     unused.Disk
	Done     bool
	Error    error
	Deleting bool
}

type output struct {
	viewport viewport.Model
	w, h     int
	delete   bool
	spinner  spinner.Model
	help     helpview
	tpl      *template.Template
	status   []*diskStatus
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewOutput() *output {
	o := &output{
		help:    NewHelp(outputKeyMap.Exec, outputKeyMap.Quit, outputKeyMap.Up, outputKeyMap.Down, outputKeyMap.PageUp, outputKeyMap.PageDown),
		spinner: spinner.New(),
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

func (o *output) SetDisks(disks unused.Disks) {
	o.status = make([]*diskStatus, len(disks))
	for i, d := range disks {
		o.status[i] = &diskStatus{Disk: d}
	}
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

	case stopExecMsg:
		o.delete = false
		outputKeyMap.Quit.SetEnabled(true)

		if o.cancel != nil {
			o.cancel()
		}

	case resumeExecMsg:
		o.delete = true
		outputKeyMap.Quit.SetEnabled(false)
		o.ctx, o.cancel = context.WithCancel(context.Background())

		return o, deleteNextDisk

	case deleteNextDiskMsg:
		for _, s := range o.status {
			if s.Done {
				// disk was already processed
				continue
			}

			select {
			case <-o.ctx.Done():
				return o, nil
			default:
				if !s.Deleting {
					s.Deleting = true

					go func() {
						s.Error = s.Disk.Provider().Delete(o.ctx, s.Disk)
						s.Done = true
						s.Deleting = false
					}()
				}

				return o, tea.Tick(50*time.Millisecond, deleteNextDiskTick)
			}
		}

		return o, stopExec
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
		return o, stopExec

	case key.Matches(msg, outputKeyMap.Up, outputKeyMap.Down, outputKeyMap.PageUp, outputKeyMap.PageDown):
		var cmd tea.Cmd
		o.viewport, cmd = o.viewport.Update(msg)
		return o, cmd

	case key.Matches(msg, outputKeyMap.Exec):
		return o, tea.Batch(o.spinner.Tick, resumeExec)

	default:
		return o, nil
	}
}

func (o *output) progressView() string {
	total := len(o.status)
	var deleted int
	for _, s := range o.status {
		if s.Done {
			deleted++
		}
	}

	var progress, spinner, fill string

	progress = fmt.Sprintf("%d/%d disks deleted ", deleted, total) // leave the space for the spinner
	if o.delete {
		spinner = o.spinner.View()
	}
	fill = strings.Repeat(" ", o.w-lipgloss.Width(progress)-lipgloss.Width(spinner)-2)

	return titleStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, progress, spinner, fill))
}

func (o *output) View() string {
	var (
		title = o.progressView()
		help  = o.help.View()
	)

	o.viewport.Height = o.h - lipgloss.Height(title) - lipgloss.Height(help) - 2

	var s strings.Builder
	err := o.tpl.Execute(&s, o.status)
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

type resumeExecMsg struct{}

func resumeExec() tea.Msg { return resumeExecMsg{} }

type stopExecMsg struct{}

func stopExec() tea.Msg { return stopExecMsg{} }

type deleteNextDiskMsg struct{}

func deleteNextDisk() tea.Msg              { return deleteNextDiskMsg{} }
func deleteNextDiskTick(time.Time) tea.Msg { return deleteNextDisk() }
