package unused

import (
	"testing"
	"time"
)

func TestDisksSort(t *testing.T) {
	var (
		now = time.Now()

		foo = provider("foo")
		baz = provider("baz")
		bar = provider("bar")

		gcp = disk{"ghi", &foo, now.Add(-10 * time.Minute)}
		aws = disk{"abc", &baz, now.Add(-5 * time.Minute)}
		az  = disk{"def", &bar, now.Add(-2 * time.Minute)}

		disks = Disks{gcp, aws, az}
	)

	tests := map[string]struct {
		exp []disk
		by  ByFunc
	}{
		"ByProvider":  {[]disk{az, aws, gcp}, ByProvider},
		"ByName":      {[]disk{aws, az, gcp}, ByName},
		"ByCreatedAt": {[]disk{gcp, aws, az}, ByCreatedAt},
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
