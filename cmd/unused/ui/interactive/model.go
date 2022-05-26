package interactive

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/grafana/unused"
	"github.com/grafana/unused/aws"
	"github.com/grafana/unused/cmd/clicommon"
)

type state int

const (
	stateProviderList state = iota
	stateProviderView
	stateFetchingDisks
	stateDeletingDisks
)

type stateChange struct {
	prev, next state
}

var _ tea.Model = Model{}

type Model struct {
	providerList list.Model
	providerView table.Model
	provider     unused.Provider
	spinner      spinner.Model
	disks        map[unused.Provider]unused.Disks
	state        state
	loadingDone  chan struct{}
	extraCols    []string
	key, value   string
	output       viewport.Model
	deleteStatus map[string]*deleteStatus
}

type deleteStatus struct {
	cur  bool
	done bool
	err  error
}

func New(providers []unused.Provider, extraColumns []string, key, value string) Model {
	return Model{
		providerList: newProviderList(providers),
		providerView: newProviderView(extraColumns),
		disks:        make(map[unused.Provider]unused.Disks),
		state:        stateProviderList,
		spinner:      spinner.New(),
		loadingDone:  make(chan struct{}),
		extraCols:    extraColumns,
		key:          key,
		value:        value,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit

		case "esc":
			switch m.state {
			case stateProviderView:
				return m, m.changeState(stateProviderList)

			case stateDeletingDisks:
				delete(m.disks, m.provider)
				return m, m.changeState(stateFetchingDisks)
			}

			return m, nil

		case "enter":
			if m.state == stateProviderList {
				return m, tea.Batch(spinner.Tick, m.changeState(stateFetchingDisks))
			}

		case "x":
			if m.state == stateProviderView {
				if rows := m.providerView.SelectedRows(); len(rows) > 0 {
					m.deleteStatus = make(map[string]*deleteStatus, len(rows))

					go m.deleteDisks()

					return m, tea.Batch(spinner.Tick, m.changeState(stateDeletingDisks))
				}
			}

		case "r":
			if m.state == stateProviderView {
				// which rows to refresh
				sel := m.providerView.SelectedRows()
				tor := make(map[string]bool, len(sel))
				for _, r := range sel {
					tor[r.Data[columnDisk].(unused.Disk).ID()] = true
				}

				ninetyDays := time.Now().Add(-90 * 24 * time.Hour)

				rows := m.providerView.GetVisibleRows()
				for i, r := range rows {
					d, ok := r.Data[columnDisk].(*aws.Disk)
					if !ok {
						continue
					}

					if tor[d.ID()] {
						if err := d.RefreshLastUsedAt(context.TODO()); err != nil {
							// TODO handle error
							rows[i].Data[columnUnused] = "ERR"
							continue
						}

						// AWS only allows to lookup events for the
						// past 90 days. If refreshing doesn't trigger
						// an error and the last used date is still
						// zero and the creation date is 90+ days it's
						// safe to assume it has been detached for
						// over 90 days.
						if d.LastUsedAt().IsZero() && d.CreatedAt().Before(ninetyDays) {
							rows[i].Data[columnUnused] = "90+d"
						} else {
							rows[i].Data[columnUnused] = clicommon.Age(d.LastUsedAt())
						}
					}
				}

				m.providerView = m.providerView.WithRows(rows)
			}
		}

	case stateChange:
		switch msg.next {
		case stateFetchingDisks:
			m.providerView = m.providerView.WithRows(nil)
			m.provider = m.providerList.SelectedItem().(providerItem).Provider

			go m.loadDisks()

		case stateProviderView:
			m.providerView = m.providerView.WithRows(disksToRows(m.disks[m.provider], m.extraCols))
		}

		m.state = msg.next

		return m, nil

	case spinner.TickMsg:
		select {
		case <-m.loadingDone:
			return m, m.changeState(stateProviderView)
		default:
		}

	case tea.WindowSizeMsg:
		m.providerList.SetSize(msg.Width, msg.Height)
		m.providerView = m.providerView.WithTargetWidth(msg.Width).WithPageSize(msg.Height - 6)
		m.output.Width = msg.Width
		m.output.Height = msg.Height - 1
	}

	var cmd tea.Cmd

	switch m.state {
	case stateFetchingDisks:
		m.spinner, cmd = m.spinner.Update(msg)

	case stateProviderList:
		m.providerList, cmd = m.providerList.Update(msg)

	case stateProviderView:
		m.providerView, cmd = m.providerView.Update(msg)

	case stateDeletingDisks:
		m.spinner, cmd = m.spinner.Update(msg)
	}

	return m, cmd
}

var errorStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#cb4b16", Dark: "#d87979"})

func (m Model) View() string {
	switch m.state {
	case stateProviderList:
		return m.providerList.View()

	case stateProviderView:
		return m.providerView.View()

	case stateFetchingDisks:
		return fmt.Sprintf("Fetching disks for %s %s %s\n", m.provider.Name(), m.provider.Meta().String(), m.spinner.View())

	case stateDeletingDisks:
		rows := m.providerView.SelectedRows()
		title := fmt.Sprintf("Deleting %d disks for %s %s\n", len(rows), m.provider.Name(), m.provider.Meta().String())

		var sb strings.Builder

		for _, r := range rows {
			d := r.Data[columnDisk].(unused.Disk)
			s := m.deleteStatus[d.ID()]
			if s != nil {
				if s.cur {
					fmt.Fprintf(&sb, "➤ %s %s\n", d.Name(), m.spinner.View())
				} else if s.done {
					if s.err != nil {
						fmt.Fprintf(&sb, "𐄂 %s\n  %s\n", d.Name(), errorStyle.Render(s.err.Error()))
					} else {
						fmt.Fprintf(&sb, "✓ %s\n", d.Name())
					}
				}
			} else {
				fmt.Fprintf(&sb, "  %s\n", d.Name())
			}
		}

		return lipgloss.JoinVertical(lipgloss.Left, title, sb.String())

	default:
		return "WHAT"
	}
}

func (m Model) loadDisks() {
	if _, ok := m.disks[m.provider]; !ok {
		disks, _ := m.provider.ListUnusedDisks(context.TODO()) // TODO handle error

		if m.key != "" {
			filtered := make(unused.Disks, 0, len(disks))
			for _, d := range disks {
				if d.Meta().Matches(m.key, m.value) {
					filtered = append(filtered, d)
				}
			}
			disks = filtered
		}

		m.disks[m.provider] = disks
	}

	m.loadingDone <- struct{}{}
}

func (m Model) deleteDisks() {
	for _, r := range m.providerView.SelectedRows() {
		d := r.Data[columnDisk].(unused.Disk)
		s := &deleteStatus{cur: true}
		m.deleteStatus[d.ID()] = s

		s.err = d.Provider().Delete(context.TODO(), d)

		s.done = true
		s.cur = false
	}
}

func (m Model) changeState(next state) tea.Cmd {
	return func() tea.Msg {
		return stateChange{m.state, next}
	}
}
