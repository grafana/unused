//go:build fake

package interactive

import (
	"strings"
	"testing"

	"github.com/grafana/unused"
	"github.com/grafana/unused/fake"
)

func TestProviderViewModel_WithDisks(t *testing.T) {
	provider := fake.NewProvider("test", 5)
	ctx := t.Context()
	disks, err := provider.ListUnusedDisks(ctx)
	if err != nil {
		t.Fatalf("Failed to list disks: %v", err)
	}

	tests := map[string]struct {
		extraCols []string
		disks     unused.Disks
	}{
		"no extra columns": {
			extraCols: []string{},
			disks:     disks,
		},
		"with kubernetes columns": {
			extraCols: []string{KubernetesNS, KubernetesPVC},
			disks:     disks,
		},
		"empty disks": {
			extraCols: []string{},
			disks:     unused.Disks{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := newProviderViewModel(tt.extraCols)
			m.SetSize(120, 40)
			m = m.WithDisks(tt.disks)

			// Verify view can be rendered
			view := m.View()
			if view == "" {
				t.Error("Expected non-empty view")
			}
		})
	}
}

func TestProviderViewModel_Empty(t *testing.T) {
	m := newProviderViewModel([]string{})
	m.SetSize(120, 40)

	provider := fake.NewProvider("test", 5)
	ctx := t.Context()
	disks, err := provider.ListUnusedDisks(ctx)
	if err != nil {
		t.Fatalf("Failed to list disks: %v", err)
	}

	m = m.WithDisks(disks)

	// Verify view shows disk data before emptying
	viewBefore := m.View()
	if viewBefore == "" {
		t.Error("Expected non-empty view with disks")
	}
	foundDisk := false
	for _, d := range disks {
		if strings.Contains(viewBefore, d.Name()) {
			foundDisk = true
			break
		}
	}
	if !foundDisk {
		t.Error("Expected to find at least one disk name in view before Empty()")
	}

	// Now empty and verify view is still renderable
	m = m.Empty()
	viewAfter := m.View()
	if viewAfter == "" {
		t.Error("Expected non-empty view even after Empty()")
	}
}

func TestProviderViewModel_Help(t *testing.T) {
	m := newProviderViewModel([]string{})
	assertHelpKeyMapView(t, m)
}
