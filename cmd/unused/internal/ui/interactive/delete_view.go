package interactive

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/grafana/unused"
)

type deleteViewModel struct {
	help     help.Model
	start    time.Time
	provider unused.Provider
	confirm  key.Binding
	toggle   key.Binding
	disks    []diskToDelete
	output   viewport.Model
	spinner  spinner.Model
	progress progress.Model
	cur      int
	delete   bool
	dryRun   bool
}

func newDeleteViewModel(dryRun bool) deleteViewModel {
	return deleteViewModel{
		help:    newHelp(),
		confirm: key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "confirm delete")),
		toggle:  key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "toggle dry-run")),
		spinner: spinner.New(),
		dryRun:  dryRun,
		progress: progress.New(
			progress.WithDefaultGradient(),
		),
	}
}

func (m deleteViewModel) WithDisks(provider unused.Provider, disks unused.Disks) deleteViewModel {
	m.provider = provider
	m.cur = 0

	m.disks = make([]diskToDelete, len(disks))
	for i, d := range disks {
		m.disks[i] = diskToDelete{disk: d}
	}

	return m
}

type deleteNextMsg struct{}

func (m deleteViewModel) Update(msg tea.Msg) (deleteViewModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.confirm):
			m.delete = true
			m.start = time.Now()
			cmd = tea.Batch(m.spinner.Tick, sendMsg(deleteNextMsg{}))

		case key.Matches(msg, m.toggle):
			m.dryRun = !m.dryRun
		}

	case deleteNextMsg:
		if m.cur == len(m.disks) {
			m.delete = false
			return m, nil
		}

		ds := m.disks[m.cur].status
		if ds == nil {
			m.disks[m.cur].status = &deleteStatus{}

			return m, deleteDisk(&m.disks[m.cur], m.dryRun)
		} else if ds.done {
			m.cur++
		}

		cmd = tea.Batch(m.spinner.Tick, sendMsg(deleteNextMsg{}))

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
	}

	return m, cmd
}

type diskToDelete struct {
	disk   unused.Disk
	status *deleteStatus
}

type deleteStatus struct {
	err  error
	done bool
}

func deleteDisk(d *diskToDelete, dryRun bool) tea.Cmd {
	return func() tea.Msg {
		if !dryRun {
			d.status.err = d.disk.Provider().Delete(context.TODO(), d.disk)
		}

		d.status.done = true

		return deleteNextMsg{}
	}
}

var bold = lipgloss.NewStyle().Bold(true)

func (m deleteViewModel) View() string {
	sb := &strings.Builder{}

	switch {
	case m.delete:
		fmt.Fprintf(sb, "Deleting %d/%d disks from %s %s\n", m.cur+1, len(m.disks), m.provider.Name(), m.provider.Meta().String())

		sb.WriteString(m.progress.ViewAs(float64(m.cur) / float64(len(m.disks))))
		if m.cur > 0 {
			eta := (time.Since(m.start) * time.Duration(len(m.disks)) / time.Duration(m.cur)).Truncate(time.Second)
			fmt.Fprintf(sb, " ETA %v", eta)
		}

		sb.WriteString("\n")

		if m.cur < len(m.disks) {
			if s := m.disks[m.cur].status; s != nil {
				sb.WriteString("➤ ")
				sb.WriteString(m.disks[m.cur].disk.Name())
				sb.WriteString(" ")
				sb.WriteString(m.spinner.View())
				sb.WriteString("\n")
			}
		}

	case m.cur == len(m.disks):
		fmt.Fprintf(sb, "Deleted %d disks from %s %s\n\n", len(m.disks), m.provider.Name(), m.provider.Meta().String())

		// TODO show table of deleted disks
		for _, d := range m.disks {
			if s := d.status; s != nil && s.err == nil && s.done {
				fmt.Fprintf(sb, "\n✓ %s", d.disk.Name())
			}
		}
	default:
		fmt.Fprintf(sb, "You're about to delete %d disks from %s %s\n\n", len(m.disks), m.provider.Name(), m.provider.Meta())

		if m.dryRun {
			fmt.Fprintln(sb, bold.Render("DISKS WON'T BE DELETED BECAUSE DRY-RUN MODE IS ENABLED"))
		} else {
			fmt.Fprintln(sb, bold.Render("Press `x` to start deleting the following disks:"))
		}

		// TODO show table of disks to be deleted
		for _, d := range m.disks {
			fmt.Fprintf(sb, "\n %s", d.disk.Name())
		}
	}

	// Print failed disks, if any
	var failed bool
	for _, d := range m.disks {
		if s := d.status; s != nil && s.err != nil {
			failed = true
			break
		}
	}

	if failed {
		sb.WriteString("\n\nThe following disks failed to be deleted:\n\n")
		for _, d := range m.disks {
			// Print current disk being deleted
			if s := d.status; s != nil && s.err != nil {
				fmt.Fprintf(sb, "𐄂 %s\n  %s\n", d.disk.Name(), errorStyle.Render(s.err.Error()))
			}
		}
	}

	m.output.SetContent(sb.String())

	return lipgloss.JoinVertical(lipgloss.Left, m.output.View(), m.help.View(m))
}

func (m deleteViewModel) ShortHelp() []key.Binding {
	return []key.Binding{navKeys.Quit, m.confirm, m.toggle, navKeys.Back}
}

func (m deleteViewModel) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.ShortHelp()}
}

func (m *deleteViewModel) SetSize(w, h int) {
	hh := lipgloss.Height(m.help.View(m))
	m.output.Width, m.output.Height = w, h-hh
	m.progress.Width = w / 2
	m.help.Width = w
}
