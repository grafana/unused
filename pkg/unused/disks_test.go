package unused_test

import (
	"testing"
	"time"

	"github.com/grafana/unused-pds/pkg/unused"
	"github.com/grafana/unused-pds/pkg/unused/unusedtest"
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
				if e := tt.exp[i]; e != got {
					t.Errorf("expecting disks[%d] %v, got %v", i, e, got)
				}
			}
		})
	}
}
