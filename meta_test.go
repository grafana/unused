package unused_test

import (
	"sort"
	"testing"

	"github.com/grafana/unused"
)

func TestMeta(t *testing.T) {
	m := &unused.Meta{
		"def": "123",
		"ghi": "456",
		"abc": "789",
	}

	t.Run("keys return sorted", func(t *testing.T) {
		keys := m.Keys()
		if len(keys) != 3 {
			t.Fatalf("expecting 3 keys, got %d", len(keys))
		}
		if !sort.StringsAreSorted(keys) {
			t.Fatalf("expecting keys to be sorted, got %q", keys)
		}
	})

	t.Run("string is sorted and comma separated", func(t *testing.T) {
		exp := "abc=789,def=123,ghi=456"
		got := m.String()

		if exp != got {
			t.Errorf("expecting String() %q, got %q", exp, got)
		}
	})
}