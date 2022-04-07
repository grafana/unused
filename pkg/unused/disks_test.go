package unused

import (
	"context"
	"fmt"
	"testing"
	"time"
)

var _ Disk = disk{}

type disk struct {
	name      string
	provider  *provider
	createdAt time.Time
}

func (d disk) Provider() Provider   { return d.provider }
func (d disk) Name() string         { return d.name }
func (d disk) CreatedAt() time.Time { return d.createdAt }

func (d disk) String() string {
	return fmt.Sprintf("disk{Provider:%q, Name:%q, CreatedAt:%q}", d.provider.Name(), d.name, d.createdAt.Format(time.RFC3339))
}

type provider string

func (p *provider) Name() string { return string(*p) }

func (p *provider) ListUnusedDisks(ctx context.Context) (Disks, error) {
	return nil, nil
}

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
