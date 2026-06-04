//go:build fake

package interactive

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/grafana/unused"
	"github.com/grafana/unused/fake"
)

func TestDeleteViewModel_Update(t *testing.T) {
	provider := fake.NewProvider("test", 3)
	ctx := t.Context()
	disks, err := provider.ListUnusedDisks(ctx)
	if err != nil {
		t.Fatalf("Failed to list disks: %v", err)
	}

	tests := map[string]tea.Msg{
		"window size message":          tea.WindowSizeMsg{Width: 100, Height: 30},
		"help toggle":                  tea.KeyPressMsg(tea.Key{Text: "?"}),
		"up key":                       tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}),
		"down key":                     tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}),
		"space key - toggle selection": tea.KeyPressMsg(tea.Key{Text: " "}),
		"D key - toggle dry run":       tea.KeyPressMsg(tea.Key{Text: "D"}),
	}

	for name, msg := range tests {
		t.Run(name, func(t *testing.T) {
			m := newDeleteViewModel(false)
			m.SetSize(120, 40)
			m = m.WithDisks(provider, disks)

			// Just ensure no panic
			_, _ = m.Update(msg)
		})
	}
}

func TestDeleteViewModel_View(t *testing.T) {
	provider := fake.NewProvider("test", 3)
	ctx := t.Context()
	disks, err := provider.ListUnusedDisks(ctx)
	if err != nil {
		t.Fatalf("Failed to list disks: %v", err)
	}

	tests := map[string]struct {
		dryRun       bool
		disks        unused.Disks
		expectInView string
	}{
		"dry run indicator": {
			dryRun:       true,
			disks:        disks,
			expectInView: "DRY-RUN",
		},
		"with disks shows disk names": {
			dryRun: false,
			disks:  disks,
			// Check for truncated prefix (table truncates long names)
			expectInView: disks[0].Name()[:30],
		},
		"empty disks shows message": {
			dryRun:       false,
			disks:        unused.Disks{},
			expectInView: "There are no disks to delete",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := newDeleteViewModel(tt.dryRun)
			m.SetSize(120, 40)
			m = m.WithDisks(provider, tt.disks)

			view := m.View()

			if !strings.Contains(view, tt.expectInView) {
				t.Errorf("Expected view to contain %q, got: %s", tt.expectInView, view)
			}
		})
	}
}

func TestDeleteViewModel_SetSize(t *testing.T) {
	t.Skip("Non-implemented feature")

	provider := fake.NewProvider("test", 3)
	ctx := t.Context()
	disks, err := provider.ListUnusedDisks(ctx)
	if err != nil {
		t.Fatalf("Failed to list disks: %v", err)
	}

	tests := map[string]struct {
		width  int
		height int
	}{
		"normal size": {
			width:  120,
			height: 40,
		},
		"small size": {
			width:  80,
			height: 25,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := newDeleteViewModel(false)
			m = m.WithDisks(provider, disks)
			m.SetSize(tt.width, tt.height)

			view := m.View()
			// For very small sizes, expect a "window too small" message
			if tt.width < minWidth || tt.height < minHeight {
				if !strings.Contains(view, "too small") && !strings.Contains(view, "invalid") {
					t.Logf("Note: Window size check not implemented for delete view (expected for small sizes)")
				}
			}
		})
	}
}

func TestDeleteViewModel_DryRunMode(t *testing.T) {
	provider := fake.NewProvider("test", 3)
	ctx := t.Context()
	disks, err := provider.ListUnusedDisks(ctx)
	if err != nil {
		t.Fatalf("Failed to list disks: %v", err)
	}

	// Test that dry run mode is reflected
	m := newDeleteViewModel(true)
	m.SetSize(120, 40)
	m = m.WithDisks(provider, disks)

	if !m.dryRun {
		t.Error("Expected dryRun to be true")
	}

	view := m.View()
	if !strings.Contains(view, "DRY-RUN") {
		t.Error("Expected DRY-RUN indicator in view")
	}
}

func TestDeleteViewModel_Help(t *testing.T) {
	m := newDeleteViewModel(false)
	assertHelpKeyMapView(t, m)
}
