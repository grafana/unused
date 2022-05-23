package interactive

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
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
