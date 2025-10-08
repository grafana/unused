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

		foo = unusedtest.NewDisk("foo", csp1, now.Add(-5*time.Hour))
		bar = unusedtest.NewDisk("bar", csp2, now.Add(-5*time.Hour))
		baz = unusedtest.NewDisk("baz", csp1, now.Add(-2*time.Hour))
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
		key, val string
		exp      unused.Disks
	}{
		"no filter": {0, "", "", disks},

		"minage": {3 * time.Hour, "", "", unused.Disks{foo, bar}},
		"keyval": {0, "dolor", "sit amet", unused.Disks{bar}},
		"both":   {2 * time.Hour, "dolor", "", unused.Disks{foo, baz}},

		"!minage": {10 * time.Hour, "", "", nil},
		"!keyval": {0, "foo", "bar", nil},
		"!both":   {10 * time.Hour, "foo", "bar", nil},
	}

	for n, tt := range tests {
		t.Run(n, func(t *testing.T) {
			opts := UI{
				Filters: Filters{
					Key:    tt.key,
					Value:  tt.val,
					MinAge: tt.minAge,
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
