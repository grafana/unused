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
)
