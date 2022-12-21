package unused

import (
	"sort"
	"strings"
)

// Meta is a map of key/value pairs.
type Meta map[string]string

// Keys returns all the keys in the map sorted alphabetically.
func (m Meta) Keys() []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

// String returns a string representation of metadata.
func (m Meta) String() string {
	var s strings.Builder
	for i, k := range m.Keys() {
		s.WriteString(k)
		s.WriteRune('=')
		s.WriteString(m[k])
		if i < len(m)-1 {
			s.WriteRune(',')
		}
	}
	return s.String()
}

// Matches returns true when the given key exists in the map with the
// given value.
func (m Meta) Matches(key, val string) bool {
	return m[key] == val
}
