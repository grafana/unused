package interactive

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/grafana/unused"
	"github.com/grafana/unused/unusedtest"
)

func TestModel_Init(t *testing.T) {
	tests := map[string]struct {
		providers     []unused.Provider
		expectMsg     bool
		expectMsgType string
	}{
		"single provider auto-selects": {
			providers:     []unused.Provider{unusedtest.NewProvider("p1", nil, nil)},
			expectMsg:     true,
			expectMsgType: "unused.Provider",
		},
		"multiple providers show list": {
			providers:     []unused.Provider{unusedtest.NewProvider("p1", nil, nil), unusedtest.NewProvider("p2", nil, nil)},
			expectMsg:     false,
			expectMsgType: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := New(tt.providers, []string{}, noFilter, false)
			cmd := m.Init()

			if tt.expectMsg && cmd == nil {
				t.Error("Expected Init() to return a command")
			}
			if !tt.expectMsg && cmd != nil {
				t.Error("Expected Init() to return nil")
			}

			if tt.expectMsg && cmd != nil {
				msg := cmd()
				if _, ok := msg.(unused.Provider); !ok && tt.expectMsgType == "unused.Provider" {
					t.Errorf("Expected message type %s, got %T", tt.expectMsgType, msg)
				}
			}
		})
	}
}

func TestModel_Update_Navigation(t *testing.T) {
	provider := unusedtest.NewProvider("p1", nil, nil)
	m := New([]unused.Provider{provider}, []string{}, noFilter, false)

	tests := map[string]struct {
		initialState state
		key          tea.Key
		expectQuit   bool
	}{
		"quit from provider list": {
			initialState: stateProviderList,
			key:          tea.Key{Text: "q"},
			expectQuit:   true,
		},
		"quit from provider view": {
			initialState: stateProviderView,
			key:          tea.Key{Text: "q"},
			expectQuit:   true,
		},
		"back from provider view": {
			initialState: stateProviderView,
			key:          tea.Key{Code: tea.KeyEsc},
			expectQuit:   false,
		},
		"back from deleting disks": {
			initialState: stateDeletingDisks,
			key:          tea.Key{Code: tea.KeyEsc},
			expectQuit:   false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m.state = tt.initialState
			m.provider = provider

			keyMsg := tea.KeyPressMsg(tt.key)

			newModel, cmd := m.Update(keyMsg)
			m = newModel.(Model)

			if tt.expectQuit && cmd == nil {
				t.Error("Expected Quit command")
			}

			// Verify state transitions for back key
			if tt.key.Code == tea.KeyEsc && tt.initialState == stateProviderView {
				if m.state != stateProviderList {
					t.Errorf("Expected state to be stateProviderList, got %v", m.state)
				}
			}
		})
	}
}

func TestModel_Update_StateTransitions(t *testing.T) {
	provider1 := unusedtest.NewProvider("p1", nil, nil)
	provider2 := unusedtest.NewProvider("p2", nil, nil)

	tests := map[string]struct {
		initialState state
		msg          tea.Msg
		expectState  state
	}{
		"provider selection triggers disk loading": {
			initialState: stateProviderList,
			msg:          provider1,
			expectState:  stateFetchingDisks,
		},
		"disks loaded shows provider view": {
			initialState: stateFetchingDisks,
			msg:          unused.Disks{},
			expectState:  stateProviderView,
		},
		"delete action from provider view": {
			initialState: stateProviderView,
			msg:          unused.Disks{},
			expectState:  stateDeletingDisks,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := New([]unused.Provider{provider1, provider2}, []string{}, noFilter, false)
			m.state = tt.initialState
			m.provider = provider1
			// Set window size to avoid rendering issues
			m.w, m.h = 120, 40
			m.providerView.SetSize(120, 40)

			newModel, _ := m.Update(tt.msg)
			m = newModel.(Model)

			if m.state != tt.expectState {
				t.Errorf("Expected state %v, got %v", tt.expectState, m.state)
			}
		})
	}
}

func TestModel_Update_MessageHandling(t *testing.T) {
	provider := unusedtest.NewProvider("p1", nil, nil)

	tests := map[string]struct {
		msg tea.Msg
	}{
		"window size message": {
			msg: tea.WindowSizeMsg{Width: 120, Height: 40},
		},
		"error message": {
			msg: errors.New("test error"),
		},
		"refresh message": {
			msg: refreshMsg{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := New([]unused.Provider{provider}, []string{}, noFilter, false)
			m.state = stateProviderView
			m.provider = provider

			_, _ = m.Update(tt.msg)
			// Just ensure no panic
		})
	}
}

func TestModel_View(t *testing.T) {
	provider := unusedtest.NewProvider("p1", nil, nil)

	tests := map[string]struct {
		state        state
		width        int
		height       int
		err          error
		expectInView string
	}{
		"window too small": {
			state:        stateProviderList,
			width:        50,
			height:       20,
			expectInView: "invalid window size",
		},
		"error state": {
			state:        stateProviderView,
			width:        120,
			height:       40,
			err:          errors.New("test error"),
			expectInView: "test error",
		},
		"fetching disks state": {
			state:        stateFetchingDisks,
			width:        120,
			height:       40,
			expectInView: "Fetching disks",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := New([]unused.Provider{provider}, []string{}, noFilter, false)
			m.state = tt.state
			m.provider = provider
			m.w = tt.width
			m.h = tt.height
			if tt.err != nil {
				m.err = tt.err
			}

			view := m.View()

			if tt.expectInView != "" && !strings.Contains(view.Content, tt.expectInView) {
				t.Errorf("Expected view to contain %q, got: %s", tt.expectInView, view.Content)
			}
		})
	}
}

// Helper types and functions

func noFilter(d unused.Disk) bool {
	return true
}
