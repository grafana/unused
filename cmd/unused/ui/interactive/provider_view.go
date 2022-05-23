package interactive

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/grafana/unused"
	"github.com/grafana/unused/cli"
)

const (
	columnDisk   = "disk"
	columnName   = "name"
	columnAge    = "age"
	columnUnused = "ageUnused"
)

var (
	headerStyle = lipgloss.NewStyle().Align(lipgloss.Center).Bold(true)
	nameStyle   = lipgloss.NewStyle().Align(lipgloss.Left)
	ageStyle    = lipgloss.NewStyle().Align(lipgloss.Right)
)

func newProviderView() table.Model {
	return table.New([]table.Column{
		table.NewFlexColumn(columnName, "Name", 2).WithStyle(nameStyle),
		table.NewColumn(columnAge, "Age", 6).WithStyle(ageStyle),
		table.NewColumn(columnUnused, "Unused", 6).WithStyle(ageStyle),
	}).
		HeaderStyle(headerStyle).
		Focused(true).
		SelectableRows(true)
}

func disksToRows(disks unused.Disks) []table.Row {
	rows := make([]table.Row, len(disks))

	for i, d := range disks {
		rows[i] = table.NewRow(table.RowData{
			columnDisk:   d,
			columnName:   d.Name(),
			columnAge:    cli.Age(d.CreatedAt()),
			columnUnused: cli.Age(d.LastUsedAt()),
		})
	}

	return rows
}
