package unused_test

import (
	"testing"
	"time"

	"github.com/grafana/unused"
	"github.com/grafana/unused/unusedtest"
)

func TestDisksSort(t *testing.T) {
	var (
		now = time.Now()

		foo = unusedtest.NewProvider("foo", nil)
		baz = unusedtest.NewProvider("baz", nil)
		bar = unusedtest.NewProvider("bar", nil)

		gcp = unusedtest.NewDisk("ghi", foo, now.Add(-10*time.Minute))
		aws = unusedtest.NewDisk("abc", baz, now.Add(-5*time.Minute))
		az  = unusedtest.NewDisk("def", bar, now.Add(-2*time.Minute))

		disks = unused.Disks{gcp, aws, az}
	)

	tests := map[string]struct {
		exp []unused.Disk
		by  unused.ByFunc
	}{
		"ByProvider":  {[]unused.Disk{az, aws, gcp}, unused.ByProvider},
		"ByName":      {[]unused.Disk{aws, az, gcp}, unused.ByName},
		"ByCreatedAt": {[]unused.Disk{gcp, aws, az}, unused.ByCreatedAt},
	}

	for n, tt := range tests {
		t.Run(n, func(t *testing.T) {
			disks.Sort(tt.by)

			for i, got := range disks {
				assertEqualDisks(t, tt.exp[i], got)
			}
		})
	}
}

func assertEqualDisks(t *testing.T, p, q unused.Disk) {
	t.Helper()

	if e, g := p.Name(), q.Name(); e != g {
		t.Errorf("expecting name %q, got %q", e, g)
	}

	if e, g := p.Provider(), q.Provider(); e != g {
		t.Errorf("expecting provider %v, got %v", e, g)
	}

	if e, g := p.CreatedAt(), q.CreatedAt(); !e.Equal(g) {
		t.Errorf("expecting created at %v, got %v", e, g)
	}

	mp, mq := p.Meta(), q.Meta()

	if e, g := len(mp), len(mq); e != g {
		t.Fatalf("expecting %d metadata items, got %d", e, g)
	}

	for k, v := range mp {
		if mq[k] != v {
			t.Errorf("expecting metadata %q with value %q, got %q", k, v, mq[k])
		}
	}
}
