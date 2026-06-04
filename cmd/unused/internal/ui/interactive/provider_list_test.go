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

	shortHelp := m.ShortHelp()
	if len(shortHelp) == 0 {
		t.Error("Expected short help to have bindings")
	}

	fullHelp := m.FullHelp()
	if len(fullHelp) == 0 {
		t.Error("Expected full help to have binding groups")
	}

	// Verify full help has more bindings than short help
	totalShort := len(shortHelp)
	totalFull := 0
	for _, group := range fullHelp {
		totalFull += len(group)
	}
	if totalFull <= totalShort {
		t.Errorf("Expected full help (%d bindings) to have more than short help (%d bindings)", totalFull, totalShort)
	}
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
