package interactive

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/grafana/unused"
	"github.com/grafana/unused/cmd/internal"
)

const (
	columnDisk   = "disk"
	columnName   = "name"
	columnAge    = "age"
	columnUnused = "ageUnused"
	columnSize   = "size"
	columnType   = "type"
)

// Custom Kubernetes columns.
// These constants are copied from the ui package.
const (
	KubernetesNS  = "__k8s:ns__"
	KubernetesPV  = "__k8s:pv__"
	KubernetesPVC = "__k8s:pvc__"
)

var k8sHeaders = map[string]string{
	KubernetesNS:  "Namespace",
	KubernetesPVC: "PVC",
	KubernetesPV:  "PV",
}

var (
	headerStyle = lipgloss.NewStyle().Align(lipgloss.Center).Bold(true)
	nameStyle   = lipgloss.NewStyle().Align(lipgloss.Left)
	ageStyle    = lipgloss.NewStyle().Align(lipgloss.Right)
)

type providerViewModel struct {
	help      help.Model
	toggle    key.Binding
	delete    key.Binding
	toggleCur key.Binding
	selAll    key.Binding
	unselAll  key.Binding
	refresh   key.Binding
	extraCols []string
	table     table.Model
	w         int
	h         int
}

func newProviderViewModel(extraColumns []string) providerViewModel {
	cols := []table.Column{
		table.NewFlexColumn(columnName, "Name", 2).WithStyle(nameStyle),
		table.NewColumn(columnAge, "Age", 6).WithStyle(ageStyle),
		table.NewColumn(columnUnused, "Unused", 6).WithStyle(ageStyle),
		table.NewColumn(columnType, "Type", 6).WithStyle(ageStyle),
		table.NewColumn(columnSize, "Size (GB)", 10).WithStyle(ageStyle),
	}

	for _, c := range extraColumns {
		h, ok := k8sHeaders[c]
		if !ok {
			h = c
		}
		cols = append(cols, table.NewFlexColumn(c, h, 1).WithStyle(nameStyle))
	}

	table := table.New(cols).
		HeaderStyle(headerStyle).
		Focused(true).
		WithSelectedText(" ", "✔").
		WithFooterVisibility(true).
		SelectableRows(true)

	return providerViewModel{
		table:  table,
		help:   newHelp(),
		toggle: key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle mark")),
		delete: key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "delete marked")),

		toggleCur: key.NewBinding(key.WithKeys("*"), key.WithHelp("*", "toggle current page")),
		selAll:    key.NewBinding(key.WithKeys("A"), key.WithHelp("A", "select all")),
		unselAll:  key.NewBinding(key.WithKeys("N"), key.WithHelp("N", "deselect all")),
		refresh:   key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "refresh disks")),

		extraCols: extraColumns,
	}
}

func (m providerViewModel) Update(msg tea.Msg) (providerViewModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.delete):
			if rows := m.table.SelectedRows(); len(rows) > 0 {
				disks := make(unused.Disks, len(rows))
				for i, r := range rows {
					disks[i] = r.Data[columnDisk].(unused.Disk)
				}
				cmd = sendMsg(disks)
			}

		case key.Matches(msg, m.toggleCur):
			rows := m.table.GetVisibleRows()
			sel := m.table.SelectedRows()
			s := (m.table.CurrentPage() - 1) * m.table.PageSize()
			e := min(s+m.table.PageSize(), len(rows))
			for i := s; i < e; i++ {
				rows[i] = rows[i].Selected(!slices.ContainsFunc(sel, func(r table.Row) bool {
					a := r.Data[columnDisk].(unused.Disk)
					b := rows[i].Data[columnDisk].(unused.Disk)
					return a.ID() == b.ID()
				}))
			}
			m.table = m.table.WithRows(rows)

		case key.Matches(msg, m.selAll, m.unselAll):
			sel := key.Matches(msg, m.selAll)
			rows := m.table.GetVisibleRows()
			for i := range rows {
				rows[i] = rows[i].Selected(sel)
			}
			m.table = m.table.WithRows(rows)

		case key.Matches(msg, m.refresh):
			return m, sendMsg(refreshMsg{})

		case msg.String() == "?":
			m.help.ShowAll = !m.help.ShowAll
			m.resetSize()

		case key.Matches(msg, navKeys.Quit):
			cmd = tea.Quit

		default:
			m.table, cmd = m.table.Update(msg)
		}
	}

	m.updateTableFooter()

	return m, cmd
}

func (m providerViewModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.table.View(), m.help.View(m))
}

func (m providerViewModel) ShortHelp() []key.Binding {
	return []key.Binding{navKeys.Quit, navKeys.Back, m.toggle, m.toggleCur, m.delete, m.refresh}
}

func (m providerViewModel) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		m.ShortHelp(),
		{m.selAll, m.unselAll, navKeys.Up, navKeys.Down, navKeys.PageUp, navKeys.PageDown, navKeys.Home, navKeys.End},
	}
}

func (m *providerViewModel) updateTableFooter() {
	var (
		t    = m.table
		sel  = fmt.Sprintf(" %d of %d disks selected", len(t.SelectedRows()), t.TotalRows())
		page = fmt.Sprintf("Page %d of %d ", t.CurrentPage(), t.MaxPages())
		f    = sel + strings.Repeat(" ", m.w-2-len(sel)-len(page)) + page
	)

	m.table = t.WithStaticFooter(f)
}

func (m *providerViewModel) resetSize() {
	hh := lipgloss.Height(m.help.View(m))
	// 4 is the table borders plus header height, 2 is the footer height.
	m.table = m.table.WithTargetWidth(m.w).WithPageSize(m.h - 4 - hh - 2)
	m.help.Width = m.w
}

func (m *providerViewModel) SetSize(w, h int) {
	m.w, m.h = w, h
	m.resetSize()
}

func (m providerViewModel) Empty() providerViewModel {
	m.table = m.table.WithRows(nil)
	return m
}

func (m providerViewModel) WithDisks(disks unused.Disks) providerViewModel {
	rows := make([]table.Row, len(disks))

	for i, d := range disks {
		row := table.RowData{
			columnDisk:   d,
			columnName:   d.Name(),
			columnAge:    internal.Age(d.CreatedAt()),
			columnUnused: internal.Age(d.LastUsedAt()),
			columnType:   d.DiskType(),
			columnSize:   d.SizeGB(),
		}

		meta := d.Meta()
		for _, c := range m.extraCols {
			var v string
			switch c {
			case KubernetesNS:
				v = meta.CreatedForNamespace()
			case KubernetesPV:
				v = meta.CreatedForPV()
			case KubernetesPVC:
				v = meta.CreatedForPVC()
			default:
				v = meta[c]
			}
			row[c] = v
		}

		rows[i] = table.NewRow(row)
	}

	m.table = m.table.WithRows(rows).WithAllRowsDeselected()
	return m
}
