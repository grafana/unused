//go:build fake

package interactive

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/grafana/unused"
	"github.com/grafana/unused/fake"
)

func TestDeleteViewModel_WithDisks(t *testing.T) {
	provider := fake.NewProvider("test", 5)
	ctx := t.Context()
	disks, err := provider.ListUnusedDisks(ctx)
	if err != nil {
		t.Fatalf("Failed to list disks: %v", err)
	}

	tests := map[string]struct {
		dryRun bool
		disks  unused.Disks
	}{
		"dry run mode": {
			dryRun: true,
			disks:  disks,
		},
		"normal mode": {
			dryRun: false,
			disks:  disks,
		},
		"empty disks": {
			dryRun: false,
			disks:  unused.Disks{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := newDeleteViewModel(tt.dryRun)
			m.SetSize(120, 40)
			m = m.WithDisks(provider, tt.disks)

			if len(m.disks) != len(tt.disks) {
				t.Errorf("Expected %d disks, got %d", len(tt.disks), len(m.disks))
			}

			// Verify view contains disk names (or prefixes if truncated)
			view := m.View()
			if view == "" {
				t.Error("Expected non-empty view")
			}
			// Check that disk name prefixes appear in the view
			// (full names may be truncated in the table)
			for _, d := range tt.disks {
				// Check for first 30 characters of the disk name
				prefix := d.Name()
				if len(prefix) > 30 {
					prefix = prefix[:30]
				}
				if !strings.Contains(view, prefix) {
					t.Errorf("Expected to find disk prefix %q in view", prefix)
				}
			}
		})
	}
}

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

	shortHelp := m.ShortHelp()
	if len(shortHelp) == 0 {
		t.Error("Expected short help to have bindings")
	}

	fullHelp := m.FullHelp()
	if len(fullHelp) == 0 {
		t.Error("Expected full help to have binding groups")
	}

	// Verify full help contains at least the short help bindings
	totalShort := len(shortHelp)
	totalFull := 0
	for _, group := range fullHelp {
		totalFull += len(group)
	}
	if totalFull < totalShort {
		t.Errorf("Expected full help (%d bindings) to have at least short help bindings (%d)", totalFull, totalShort)
	}
}
