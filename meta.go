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

func (m Meta) Equals(b Meta) bool {
	if len(m) != len(b) {
		return false
	}

	for ak, av := range m {
		bv, ok := b[ak]
		if !ok || av != bv {
			return false
		}
	}

	return true
}

// Matches returns true when the given key exists in the map with the
// given value.
func (m Meta) Matches(key, val string) bool {
	switch key {
	case "k8s:pv":
		return m.CreatedForPV() == val
	case "k8s:pvc":
		return m.CreatedForPVC() == val
	case "k8s:ns":
		return m.CreatedForNamespace() == val
	}
	return m[key] == val
}

func (m Meta) CreatedForPV() string {
	return m.coalesce("kubernetes.io/created-for/pv/name", "kubernetes.io-created-for-pv-name")
}

func (m Meta) CreatedForPVC() string {
	return m.coalesce("kubernetes.io/created-for/pvc/name", "kubernetes.io-created-for-pvc-name")
}

func (m Meta) CreatedForNamespace() string {
	return m.coalesce("kubernetes.io/created-for/pvc/namespace", "kubernetes.io-created-for-pvc-namespace")
}

func (m Meta) Zone() string {
	return m.coalesce("zone", "location")
}

func (m Meta) coalesce(keys ...string) string {
	for _, k := range keys {
		v, ok := m[k]
		if !ok {
			continue
		}

		return v
	}

	return ""
}
