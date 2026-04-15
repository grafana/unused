package ui

import (
	"testing"
	"time"

	"github.com/grafana/unused"
	"github.com/grafana/unused/unusedtest"
)

func TestUI_Filter(t *testing.T) {
	var (
		csp1 = unusedtest.NewProvider("foo", nil)
		csp2 = unusedtest.NewProvider("bar", nil)

		now = time.Now()

		foo = unusedtest.NewDisk("foo", csp1, now.Add(-5*time.Hour), now.Add(-3*time.Hour))
		bar = unusedtest.NewDisk("bar", csp2, now.Add(-5*time.Hour), now.Add(-10*time.Second))
		baz = unusedtest.NewDisk("baz", csp1, now.Add(-2*time.Hour), now.Add(-1*time.Hour-30&time.Minute))
	)

	foo.SetMeta(unused.Meta{"lorem": "ipsum"})
	bar.SetMeta(unused.Meta{"lorem": "ipsum", "dolor": "sit amet"})
	baz.SetMeta(unused.Meta{"lorem": "quux"})

	disks := unused.Disks{foo, bar, baz}

	eq := func(a, b unused.Disks) bool {
		if len(a) != len(b) {
			return false
		}
		for i, e := range a {
			if b[i].Name() != e.Name() {
				return false
			}
		}
		return true
	}

	tests := map[string]struct {
		minAge   time.Duration
		unused   time.Duration
		key, val string
		exp      unused.Disks
	}{
		"no filter": {0, 0, "", "", disks},

		"minage": {3 * time.Hour, 0, "", "", unused.Disks{foo, bar}},
		"keyval": {0, 0, "dolor", "sit amet", unused.Disks{bar}},
		"both":   {2 * time.Hour, 0, "dolor", "", unused.Disks{foo, baz}},

		"!minage": {10 * time.Hour, 0, "", "", nil},
		"!keyval": {0, 0, "foo", "bar", nil},
		"!both":   {10 * time.Hour, 0, "foo", "bar", nil},

		"unused":  {0, 1 * time.Hour, "", "", unused.Disks{foo, baz}},
		"!unused": {0, 6 * time.Hour, "", "", nil},
	}

	for n, tt := range tests {
		t.Run(n, func(t *testing.T) {
			opts := UI{
				Filters: Filters{
					Key:       tt.key,
					Value:     tt.val,
					MinAge:    tt.minAge,
					MinUnused: tt.unused,
				},
			}

			got := disks.Filter(opts.Filter)
			if !eq(got, tt.exp) {
				for _, d := range disks {
					t.Error(tt.key, tt.val, d.Meta())
				}
				t.Errorf("slices are not equal\nexp: %v\ngot: %v", tt.exp, got)
			}
		})
	}
}
