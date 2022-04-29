package unused

import (
	"sort"
	"strings"
)

type Meta map[string]string

func (m Meta) Keys() []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

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

func (m Meta) Matches(key, val string) bool {
	return m[key] == val
}
