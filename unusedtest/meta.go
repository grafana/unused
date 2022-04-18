package unusedtest

import (
	"testing"

	"github.com/grafana/unused"
)

func AssertEqualMeta(t *testing.T, p, q unused.Meta) {
	t.Helper()

	if e, g := len(p), len(q); e != g {
		t.Fatalf("expecting %d metadata items, got %d", e, g)
	}

	for k, v := range p {
		if g := q[k]; v != g {
			t.Errorf("expecting metadata item %q with value %q, got %q", k, v, g)
		}
	}
}
