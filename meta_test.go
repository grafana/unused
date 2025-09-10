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

	t.Run("Kubernetes", func(t *testing.T) {
		m := &Meta{
			"kubernetes.io/created-for/pv/name":       "pv-foo",
			"kubernetes.io/created-for/pvc/name":      "pvc-bar",
			"kubernetes.io/created-for/pvc/namespace": "ns-quux",
		}

		if !m.Matches("k8s:pv", "pv-foo") {
			t.Error("expecting to match PV")
		}
		if !m.Matches("k8s:pvc", "pvc-bar") {
			t.Error("expecting to match PVC")
		}
		if !m.Matches("k8s:ns", "ns-quux") {
			t.Error("expecting to match namespace")
		}
	})
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

func TestCreatedForPV(t *testing.T) {
	tests := []struct {
		name     string
		m        Meta
		expected string
	}{
		{
			name:     "GCP disk",
			m:        Meta{"kubernetes.io/created-for/pv/name": "pvc-c898536e-1601-4357-af13-01bbe82f3055"},
			expected: "pvc-c898536e-1601-4357-af13-01bbe82f3055",
		},
		{
			name:     "AWS disk",
			m:        Meta{"kubernetes.io/created-for/pv/name": "pvc-b78d13ec-426f-4ec6-80aa-231a7d4e7db9"},
			expected: "pvc-b78d13ec-426f-4ec6-80aa-231a7d4e7db9",
		},
		{
			name:     "Azure disk",
			m:        Meta{"kubernetes.io-created-for-pv-name": "pvc-10df52de-2b9d-44a2-8901-4cbfc4871f8c"},
			expected: "pvc-10df52de-2b9d-44a2-8901-4cbfc4871f8c",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.m.CreatedForPV()
			if tt.expected != actual {
				t.Fatalf("expected %v but got %v", tt.expected, actual)
			}
		})
	}
}

func TestCreatedForPVC(t *testing.T) {
	tests := []struct {
		name     string
		m        Meta
		expected string
	}{
		{
			name:     "GCP disk",
			m:        Meta{"kubernetes.io/created-for/pvc/name": "qwerty"},
			expected: "qwerty",
		},
		{
			name:     "AWS disk",
			m:        Meta{"kubernetes.io/created-for/pvc/name": "asdf"},
			expected: "asdf",
		},
		{
			name:     "Azure disk",
			m:        Meta{"kubernetes.io-created-for-pvc-name": "zxcv"},
			expected: "zxcv",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.m.CreatedForPVC()
			if tt.expected != actual {
				t.Fatalf("expected %v but got %v", tt.expected, actual)
			}
		})
	}
}

func TestCreatedForNamespace(t *testing.T) {
	tests := []struct {
		name     string
		m        Meta
		expected string
	}{
		{
			name:     "GCP disk",
			m:        Meta{"kubernetes.io/created-for/pvc/namespace": "ns1"},
			expected: "ns1",
		},
		{
			name:     "AWS disk",
			m:        Meta{"kubernetes.io/created-for/pvc/namespace": "ns2"},
			expected: "ns2",
		},
		{
			name:     "Azure disk",
			m:        Meta{"kubernetes.io-created-for-pvc-namespace": "ns3"},
			expected: "ns3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.m.CreatedForNamespace()
			if tt.expected != actual {
				t.Fatalf("expected %v but got %v", tt.expected, actual)
			}
		})
	}
}

func TestZone(t *testing.T) {
	tests := []struct {
		name     string
		m        Meta
		expected string
	}{
		{
			name:     "GCP disk",
			m:        Meta{"zone": "asia-south1-a"},
			expected: "asia-south1-a",
		},
		{
			name:     "AWS disk",
			m:        Meta{"zone": "us-east-2a"},
			expected: "us-east-2a",
		},
		{
			name:     "Azure disk",
			m:        Meta{"location": "Central US"},
			expected: "Central US",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.m.Zone()
			if tt.expected != actual {
				t.Fatalf("expected %v but got %v", tt.expected, actual)
			}
		})
	}
}
