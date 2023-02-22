package interactive

import (
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

var (
	headerStyle = lipgloss.NewStyle().Align(lipgloss.Center).Bold(true)
	nameStyle   = lipgloss.NewStyle().Align(lipgloss.Left)
	ageStyle    = lipgloss.NewStyle().Align(lipgloss.Right)
)

func newProviderView(extraColumns []string) table.Model {
	cols := []table.Column{
		table.NewFlexColumn(columnName, "Name", 2).WithStyle(nameStyle),
		table.NewColumn(columnAge, "Age", 6).WithStyle(ageStyle),
		table.NewColumn(columnUnused, "Unused", 6).WithStyle(ageStyle),
		table.NewColumn(columnType, "Type", 6).WithStyle(ageStyle),
		table.NewColumn(columnSize, "Size (GB)", 10).WithStyle(ageStyle),
	}

	for _, c := range extraColumns {
		cols = append(cols, table.NewFlexColumn(c, c, 1).WithStyle(nameStyle))
	}

	return table.New(cols).
		HeaderStyle(headerStyle).
		Focused(true).
		WithSelectedText(" ", "✔").
		WithFooterVisibility(false).
		SelectableRows(true)
}

func disksToRows(disks unused.Disks, extraColumns []string) []table.Row {
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
		for _, c := range extraColumns {
			row[c] = meta[c]
		}

		rows[i] = table.NewRow(row)
	}

	return rows
}
