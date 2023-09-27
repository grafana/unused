package interactive

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/grafana/unused"
)

type deleteViewModel struct {
	output   viewport.Model
	help     help.Model
	w, h     int
	confirm  key.Binding
	provider unused.Provider
	spinner  spinner.Model
	delete   bool
	disks    unused.Disks
	cur      int
	status   []*deleteStatus
	dryRun   bool
}

func newDeleteViewModel(dryRun bool) deleteViewModel {
	return deleteViewModel{
		help:    newHelp(),
		confirm: key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "confirm delete")),
		spinner: spinner.New(),
		dryRun:  dryRun,
	}
}

func (m deleteViewModel) WithDisks(provider unused.Provider, disks unused.Disks) deleteViewModel {
	m.provider = provider
	m.disks = disks
	m.status = make([]*deleteStatus, len(disks))
	m.cur = 0
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
			cmd = tea.Batch(m.spinner.Tick, sendMsg(deleteNextMsg{}))
		}

	case deleteNextMsg:
		if m.cur == len(m.disks) {
			m.delete = false
			return m, nil
		}

		ds := m.status[m.cur]
		if ds == nil {
			ds = &deleteStatus{}
			m.status[m.cur] = ds

			return m, deleteDisk(m.provider, m.disks[m.cur], ds, m.dryRun)
		} else if ds.done {
			m.cur++
		}

		cmd = tea.Batch(m.spinner.Tick, sendMsg(deleteNextMsg{}))

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
	}

	return m, cmd
}

type deleteStatus struct {
	done bool
	err  error
}

func deleteDisk(p unused.Provider, d unused.Disk, s *deleteStatus, dryRun bool) tea.Cmd {
	return func() tea.Msg {
		if !dryRun {
			s.err = d.Provider().Delete(context.TODO(), d)
		}

		s.done = true

		return deleteNextMsg{}
	}
}

var bold = lipgloss.NewStyle().Bold(true)

func (m deleteViewModel) View() string {
	sb := &strings.Builder{}

	if m.delete {
		fmt.Fprintf(sb, "Deleting %d/%d disks from %s %s\n\n", m.cur+1, len(m.disks), m.provider.Name(), m.provider.Meta().String())
	} else if m.cur == len(m.disks) {
		fmt.Fprintf(sb, "Deleted %d disks from %s %s\n\n", len(m.disks), m.provider.Name(), m.provider.Meta().String())
	} else {
		confirm := bold.Render("Press `x` to start deleting the following disks:")
		fmt.Fprintf(sb, "You're about to delete %d disks from %s %s\n%s\n", len(m.disks), m.provider.Name(), m.provider.Meta(), confirm)

		if m.dryRun {
			dryRun := bold.Render("DISKS WON'T BE DELETED BECAUSE DRY-RUN MODE IS ENABLED")
			fmt.Fprintf(sb, "\n%s\n\n", dryRun)
		}
	}

	for i, d := range m.disks {
		s := m.status[i]

		switch {
		case s == nil:
			fmt.Fprintf(sb, "  %s\n", d.Name())

		case m.cur == i:
			fmt.Fprintf(sb, "➤ %s %s\n", d.Name(), m.spinner.View())

		case !s.done:

		case s.err != nil:
			fmt.Fprintf(sb, "𐄂 %s\n  %s\n", d.Name(), errorStyle.Render(s.err.Error()))

		default:
			fmt.Fprintf(sb, "✓ %s\n", d.Name())
		}
	}

	m.output.SetContent(sb.String())

	return lipgloss.JoinVertical(lipgloss.Left, m.output.View(), m.help.View(m))
}

func (m deleteViewModel) ShortHelp() []key.Binding {
	return []key.Binding{navKeys.Quit, m.confirm, navKeys.Back}
}

func (m deleteViewModel) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.ShortHelp()}
}

func (m *deleteViewModel) resetSize() {
	hh := lipgloss.Height(m.help.View(m))
	m.output.Width, m.output.Height = m.w, m.h-hh
	m.help.Width = m.w
}

func (m *deleteViewModel) SetSize(w, h int) {
	m.w, m.h = w, h
	m.resetSize()
}
