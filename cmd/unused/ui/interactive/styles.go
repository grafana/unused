package interactive

import "github.com/charmbracelet/lipgloss"

var (
	// header in disk details view
	headerStyle = lipgloss.NewStyle().Bold(true)

	// main list, disk details, and output sections
	sectionStyle       = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
	activeSectionStyle = sectionStyle.Copy().Border(lipgloss.RoundedBorder())

	// provider tabs
	tab       = lipgloss.NewStyle().Faint(true).Padding(0, 2).BorderStyle(lipgloss.NormalBorder())
	activeTab = tab.Copy().Faint(false).Bold(true).BorderStyle(lipgloss.RoundedBorder())

	// list item when marked
	markedColor = lipgloss.AdaptiveColor{Light: "#cb4b16", Dark: "#d87979"}
	markedStyle = lipgloss.NewStyle().Strikethrough(true).Foreground(markedColor)

	// output styles
	titleStyle = lipgloss.NewStyle().Bold(true).Border(lipgloss.RoundedBorder())
	errStyle   = lipgloss.NewStyle().Foreground(markedColor)

	// align things to center
	centerStyle = lipgloss.NewStyle().Align(lipgloss.Center)
)
