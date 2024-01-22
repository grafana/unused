package unused

import (
	"sort"
	"testing"
)

func TestMeta(t *testing.T) {
	m := &Meta{
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

func TestMetaMatches(t *testing.T) {
	m := &Meta{
		"def": "123",
		"ghi": "456",
		"abc": "789",
	}

	if ok := m.Matches("ghi", "456"); !ok {
		t.Error("expecting match")
	}
	if ok := m.Matches("zyx", "123"); ok {
		t.Error("expecting no match for unrecognized key")
	}
	if ok := m.Matches("def", "789"); ok {
		t.Error("expecting no match for different value")
	}
}

func TestCoalesce(t *testing.T) {
	tests := []struct {
		name     string
		m        Meta
		input    []string
		expected string
	}{
		{
			name: "single key returns self",
			m: Meta{
				"foo": "bar",
			},
			input:    []string{"foo"},
			expected: "bar",
		},
		{
			name: "multiple keys returns first non-nil, single match",
			m: Meta{
				"foo": "bar",
			},
			input:    []string{"buz", "foo"},
			expected: "bar",
		},
		{
			name: "multiple keys returns first non-nil, many possible matches",
			m: Meta{
				"foo": "bar",
				"buz": "qux",
			},
			input:    []string{"buz", "foo"},
			expected: "qux",
		},
		{
			name: "any value is returned if key is present",
			m: Meta{
				"foo": "",
				"buz": "qux",
			},
			input:    []string{"foo", "buz"},
			expected: "",
		},
		{
			name: "no given keys returns zero value",
			m: Meta{
				"foo": "bar",
				"buz": "qux",
			},
			input:    []string{},
			expected: "",
		},
		{
			name: "no matching keys returns zero value",
			m: Meta{
				"foo": "bar",
				"buz": "qux",
			},
			input:    []string{"nope"},
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.m.coalesce(tt.input...)
			if tt.expected != actual {
				t.Fatalf("expected %v but got %v", tt.expected, actual)
			}
		})
	}
}

func TestEquals(t *testing.T) {
	tests := []struct {
		name     string
		m        Meta
		input    Meta
		expected bool
	}{
		{
			name:     "nil values are equal",
			m:        Meta{},
			input:    Meta{},
			expected: true,
		},
		{
			name:     "nil & non-nil values are not equal",
			m:        Meta{"not": "nil"},
			input:    Meta{},
			expected: false,
		},
		{
			name:     "same keys but different values are not equal",
			m:        Meta{"a": "b"},
			input:    Meta{"a": "c"},
			expected: false,
		},
		{
			name:     "same values but different keys are not equal",
			m:        Meta{"a": "b"},
			input:    Meta{"c": "b"},
			expected: false,
		},
		{
			name:     "same keys & values are equal",
			m:        Meta{"a": "b", "c": "d"},
			input:    Meta{"a": "b", "c": "d"},
			expected: true,
		},
		{
			name:     "order is irrelevant",
			m:        Meta{"a": "b", "c": "d"},
			input:    Meta{"c": "d", "a": "b"},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.m.Equals(tt.input)
			if tt.expected != actual {
				t.Fatalf("expected %v but got %v", tt.expected, actual)
			}
		})
	}
}
