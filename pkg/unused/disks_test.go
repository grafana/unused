package unused

import (
	"fmt"
	"testing"
	"time"
)

var _ Disk = disk{}

type disk struct {
	provider, name string
	createdAt      time.Time
}

func (d disk) Provider() string     { return d.provider }
func (d disk) Name() string         { return d.name }
func (d disk) CreatedAt() time.Time { return d.createdAt }

func (d disk) String() string {
	return fmt.Sprintf("disk{Provider:%q, Name:%q, CreatedAt:%q}", d.provider, d.name, d.createdAt.Format(time.RFC3339))
}

func TestDisksSort(t *testing.T) {
	var (
		now = time.Now()

		gcp = disk{"gcp", "foo", now.Add(-10 * time.Minute)}
		aws = disk{"aws", "baz", now.Add(-5 * time.Minute)}
		az  = disk{"az", "bar", now.Add(-2 * time.Minute)}

		disks = Disks{gcp, aws, az}
	)

	tests := map[string]struct {
		exp []disk
		by  ByFunc
	}{
		"ByProvider":  {[]disk{aws, az, gcp}, ByProvider},
		"ByName":      {[]disk{az, aws, gcp}, ByName},
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
