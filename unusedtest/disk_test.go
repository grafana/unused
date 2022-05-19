package unusedtest_test

import (
	"testing"
	"time"

	"github.com/grafana/unused/unusedtest"
)

func TestDisk(t *testing.T) {
	p := unusedtest.NewProvider("my-provider", nil)
	createdAt := time.Now().Round(0)
	d := unusedtest.NewDisk("my-disk", p, createdAt)

	if exp, got := "my-disk", d.ID(); exp != got {
		t.Errorf("expecting ID() %q, got %q", exp, got)
	}
	if d.Name() != "my-disk" {
		t.Errorf("expecting Name() my-disk, got %s", d.Name())
	}
	if got := d.CreatedAt(); !createdAt.Equal(got) {
		t.Errorf("expectng CreatedAt() %v, got %v", createdAt, got)
	}
	if got := d.Provider(); got != p {
		t.Errorf("expecting Provider() %v, got %v", p, got)
	}
	if got, exp := d.LastUsedAt(), d.CreatedAt().Add(time.Minute); !got.Equal(exp) {
		t.Errorf("expecting LastUsedAt() %v, got %v", exp, got)
	}
}
