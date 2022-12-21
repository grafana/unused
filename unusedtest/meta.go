package unusedtest

import (
	"fmt"

	"github.com/grafana/unused"
)

// AssertEqualMeta returns nil if both [unused.Meta] arguments are
// equal.
func AssertEqualMeta(p, q unused.Meta) error {
	if e, g := len(p), len(q); e != g {
		return fmt.Errorf("expecting %d metadata items, got %d", e, g)
	}

	for k, v := range p {
		if g := q[k]; v != g {
			return fmt.Errorf("expecting metadata item %q with value %q, got %q", k, v, g)
		}
	}

	return nil
}
