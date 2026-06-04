package interactive

import (
	"strings"
	"testing"

	"github.com/grafana/unused"
	"github.com/grafana/unused/unusedtest"
)

func TestProviderListModel_Help(t *testing.T) {
	m := newProviderListModel([]unused.Provider{})
	assertHelpKeyMapView(t, m)
}

func TestProviderItem(t *testing.T) {
	meta := unused.Meta{
		"name": "test",
	}
	provider := unusedtest.NewProvider("Test", meta, nil)
	item := providerItem{Provider: provider}

	if item.Title() != provider.Name() {
		t.Errorf("Expected title %q, got %q", provider.Name(), item.Title())
	}

	filterValue := item.FilterValue()
	// Filter value should contain both name and metadata
	if !strings.Contains(filterValue, provider.Name()) || !strings.Contains(filterValue, meta.String()) {
		t.Errorf("Expected filter value to contain %q, got %q", provider.Name(), filterValue)
	}

	// Description should contain the provider metadata
	if desc := item.Description(); !strings.Contains(desc, meta.String()) {
		t.Errorf("Expected description to contain provider metadata, got %q", desc)
	}
}
