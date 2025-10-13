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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/grafana/unused"
)

type deleteViewModel struct {
	help     help.Model
	start    time.Time
	provider unused.Provider
	confirm  key.Binding
	toggle   key.Binding
	disks    []diskToDelete
	spinner  spinner.Model
	table    table.Model
	progress progress.Model
	cur      int
	delete   bool
	dryRun   bool
}

const (
	columnMark   = "mark"
	columnStatus = "status"
)

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
		table: table.New([]table.Column{
			table.NewColumn(columnMark, " ", 2),
			table.NewFlexColumn(columnName, "Name", 1).WithStyle(nameStyle),
			table.NewFlexColumn(columnStatus, "Status", 2).WithStyle(nameStyle),
		}).
			HeaderStyle(headerStyle).
			Focused(true).
			WithPageSize(10).WithPaginationWrapping(true).
			WithFooterVisibility(true),
	}
}

func (m deleteViewModel) WithDisks(provider unused.Provider, disks unused.Disks) deleteViewModel {
	m.provider = provider
	m.cur = 0

	m.disks = make([]diskToDelete, len(disks))
	rows := make([]table.Row, len(disks))
	for i, d := range disks {
		m.disks[i] = diskToDelete{disk: d}
		rows[i] = table.NewRow(table.RowData{
			columnName: d.Name(),
		})
	}

	m.table = m.table.WithRows(rows)

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

		default:
			m.table, cmd = m.table.Update(msg)
		}

	case deleteNextMsg:
		if m.cur == len(m.disks) {
			m.delete = false
			return m, nil
		}

		rows := m.table.GetVisibleRows()

		ds := m.disks[m.cur].status
		if ds == nil {
			m.disks[m.cur].status = &deleteStatus{}
			rows[m.cur] = rows[m.cur].Selected(true)

			return m, deleteDisk(&m.disks[m.cur], m.dryRun)
		} else if ds.done {
			data := rows[m.cur].Data
			if ds.err == nil {
				data[columnMark] = "✔"
			} else {
				data[columnMark] = "❌"
				data[columnStatus] = errorStyle.Render(ds.err.Error())
			}
			rows[m.cur].Data = data
			rows[m.cur] = rows[m.cur].Selected(false)

			m.table = m.table.WithCurrentPage((m.cur / m.table.PageSize()) + 1)

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
			ctx, cancel := context.WithTimeout(context.TODO(), timeout)
			defer cancel()

			d.status.err = d.disk.Provider().Delete(ctx, d.disk)
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
		eta := "N/A"
		if m.cur > 0 {
			eta = (time.Since(m.start) * time.Duration(len(m.disks)) / time.Duration(m.cur)).Truncate(time.Second).String()
		}

		sb.WriteString(" ETA " + eta)
		sb.WriteString("\n")

		if m.cur < len(m.disks) {
			if s := m.disks[m.cur].status; s != nil {
				sb.WriteString("➤ ")
				sb.WriteString(m.disks[m.cur].disk.Name())
				sb.WriteString(" ")
				sb.WriteString(m.spinner.View())
			}
		}

		sb.WriteString("\n")

	case m.cur == len(m.disks):
		fmt.Fprintf(sb, "Deleted %d disks from %s %s\n\n\n", len(m.disks), m.provider.Name(), m.provider.Meta().String())

	default:
		fmt.Fprintf(sb, "You're about to delete %d disks from %s %s\n\n", len(m.disks), m.provider.Name(), m.provider.Meta())

		if m.dryRun {
			fmt.Fprintln(sb, bold.Render("DISKS WON'T BE DELETED BECAUSE DRY-RUN MODE IS ENABLED"))
		} else {
			fmt.Fprintln(sb, bold.Render("Press `x` to start deleting the following disks:"))
		}

	}

	sb.WriteString(m.table.View())

	return lipgloss.JoinVertical(lipgloss.Left, sb.String(), m.help.View(m))
}

func (m deleteViewModel) ShortHelp() []key.Binding {
	return []key.Binding{navKeys.Quit, m.confirm, m.toggle, navKeys.Back}
}

func (m deleteViewModel) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.ShortHelp()}
}

func (m *deleteViewModel) SetSize(w, h int) {
	m.progress.Width = w / 2
	m.help.Width = w
	hh := lipgloss.Height(m.help.View(m))
	m.table = m.table.WithMaxTotalWidth(w - 2).WithTargetWidth(w - 4).WithPageSize(h - hh - 3 - 6)
}
