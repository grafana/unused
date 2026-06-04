//go:build fake

package interactive

import (
	"strings"
	"testing"

	"github.com/grafana/unused"
	"github.com/grafana/unused/fake"
)

func TestProviderListModel_Help(t *testing.T) {
	provider := fake.NewProvider("test", 5)
	m := newProviderListModel([]unused.Provider{provider})
	assertHelpKeyMapView(t, m)
}

func TestProviderItem(t *testing.T) {
	provider := fake.NewProvider("test", 5)
	item := providerItem{Provider: provider}

	if item.Title() != "Test" {
		t.Errorf("Expected title 'Test', got %q", item.Title())
	}

	filterValue := item.FilterValue()
	if filterValue == "" {
		t.Error("Expected non-empty filter value")
	}
	// Filter value should contain both name and metadata
	if !strings.Contains(filterValue, "Test") {
		t.Errorf("Expected filter value to contain 'Test', got %q", filterValue)
	}

	desc := item.Description()
	if desc == "" {
		t.Error("Expected non-empty description")
	}
	// Description should contain the provider metadata
	if !strings.Contains(desc, "name") || !strings.Contains(desc, "test") {
		t.Errorf("Expected description to contain provider metadata, got %q", desc)
	}
}
