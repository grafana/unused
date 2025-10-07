package interactive

import (
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
	selAll    key.Binding
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
		WithFooterVisibility(false).
		SelectableRows(true)

	return providerViewModel{
		table:  table,
		help:   newHelp(),
		toggle: key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle mark")),
		delete: key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "delete marked")),

		extraCols: extraColumns,
	}
}

func (m providerViewModel) Update(msg tea.Msg) (providerViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.delete):
			if rows := m.table.SelectedRows(); len(rows) > 0 {
				disks := make(unused.Disks, len(rows))
				for i, r := range rows {
					disks[i] = r.Data[columnDisk].(unused.Disk)
				}
				return m, sendMsg(disks)
			}

		case msg.String() == "?":
			m.help.ShowAll = !m.help.ShowAll
			m.resetSize()
			return m, nil

		case key.Matches(msg, navKeys.Quit):
			return m, tea.Quit

		default:
			var cmd tea.Cmd
			m.table, cmd = m.table.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m providerViewModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.table.View(), m.help.View(m))
}

func (m providerViewModel) ShortHelp() []key.Binding {
	return []key.Binding{navKeys.Quit, navKeys.Back, m.toggle, m.delete, navKeys.Up, navKeys.Down}
}

func (m providerViewModel) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		m.ShortHelp(),
		{navKeys.PageUp, navKeys.PageDown, navKeys.Home, navKeys.End},
	}
}

func (m *providerViewModel) resetSize() {
	hh := lipgloss.Height(m.help.View(m))
	m.table = m.table.WithTargetWidth(m.w).WithPageSize(m.h - 4 - hh)
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

	m.table = m.table.WithRows(rows)
	return m
}
