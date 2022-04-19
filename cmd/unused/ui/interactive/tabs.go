package interactive

import "github.com/charmbracelet/lipgloss"

var (
	tab       = lipgloss.NewStyle().Faint(true).Padding(0, 2).BorderStyle(lipgloss.NormalBorder())
	activeTab = tab.Copy().Faint(false).Bold(true).BorderStyle(lipgloss.RoundedBorder())
)

type Tabs struct {
	Titles []string
	cur    int
}

func (t *Tabs) Selected() string {
	return t.Titles[t.cur]
}

func (t *Tabs) View() string {
	tabs := make([]string, len(t.Titles))

	for i, title := range t.Titles {
		style := tab
		if i == t.cur {
			style = activeTab
		}
		tabs[i] = style.Render(title)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (t *Tabs) Next() {
	t.cur = (t.cur + 1) % len(t.Titles)
}

func (t *Tabs) Prev() {
	if t.cur == 0 {
		t.cur = len(t.Titles) - 1
	} else {
		t.cur = (t.cur - 1) % len(t.Titles)
	}
}
