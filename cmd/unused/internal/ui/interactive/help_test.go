package interactive

import (
	"testing"

	"charm.land/bubbles/v2/help"
)

func assertHelpKeyMapView(t *testing.T, m help.KeyMap) {
	t.Helper()

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
	if totalFull < totalShort {
		t.Errorf("Expected full help (%d bindings) to have at least short help (%d bindings)", totalFull, totalShort)
	}
}
