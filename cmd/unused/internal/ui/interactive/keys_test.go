//go:build fake

package interactive

import (
	"testing"

	"charm.land/bubbles/v2/key"
)

func TestKeybindingsHelp(t *testing.T) {
	// Verify all key bindings have help text
	allKeys := []key.Binding{
		navKeys.Quit,
		navKeys.Up,
		navKeys.Down,
		navKeys.PageUp,
		navKeys.PageDown,
		navKeys.Home,
		navKeys.End,
		navKeys.Back,
	}

	for i, k := range allKeys {
		help := k.Help()
		if help.Key == "" || help.Desc == "" {
			t.Errorf("Key %d missing help text: key=%q desc=%q", i, help.Key, help.Desc)
		}
	}
}
