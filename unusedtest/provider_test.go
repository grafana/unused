package unusedtest_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/grafana/unused"
	"github.com/grafana/unused/unusedtest"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name  string
		disks unused.Disks
	}{
		{"no-disks", nil},
	}

	for _, tt := range tests {
		ctx := context.Background()
		p := unusedtest.NewProvider(tt.name, nil, tt.disks...)

		if p.Name() != tt.name {
			t.Errorf("unexpected provider.Name() %q", p.Name())
		}

		disks, err := p.ListUnusedDisks(ctx)
		if err != nil {
			t.Fatal("unexpected error:", err)
		}
		if got, exp := len(disks), len(tt.disks); exp != got {
			t.Fatalf("expecting %d disks, got %d", exp, got)
		}

		for i, exp := range tt.disks {
			got := disks[i]
			if exp != got {
				t.Errorf("expecting disk %d to be %v, got %v", i, exp, got)
			}
		}
	}
}

func TestProviderMeta(t *testing.T) {
	unusedtest.TestProviderMeta(t, func(meta unused.Meta) (unused.Provider, error) {
		return unusedtest.NewProvider("my-provider", meta, nil), nil
	})
}

func TestProviderDelete(t *testing.T) {
	ctx := context.Background()

	setup := func() (unused.Provider, unused.Disks) {
		now := time.Now()

		disks := make(unused.Disks, 10)
		p := unusedtest.NewProvider("my-provider", nil, disks...)
		for i := 0; i < cap(disks); i++ {
			disks[i] = unusedtest.NewDisk(fmt.Sprintf("disk-%03d", i), p, now)
		}

		return p, disks
	}

	run := func(t *testing.T, p unused.Provider, disks unused.Disks, idx int) {
		t.Helper()

		t.Logf("deleting disk at index %d", idx)

		d := disks[idx]

		if err := p.Delete(ctx, d); err != nil {
			t.Fatalf("unexpected error when deleting: %v", err)
		}

		ds, err := p.ListUnusedDisks(context.Background())
		if err != nil {
			t.Fatalf("unexpected error listing: %v", err)
		}

		if exp, got := len(disks)-1, len(ds); exp != got {
			t.Errorf("expecting %d disks, got %d", exp, got)
		}

		for i := range ds {
			if ds[i].Name() == d.Name() {
				t.Fatalf("found disk %v at index %d", d, i)
			}
		}
	}

	t.Run("first", func(t *testing.T) {
		p, disks := setup()
		run(t, p, disks, 0)
	})

	t.Run("last", func(t *testing.T) {
		p, disks := setup()
		run(t, p, disks, len(disks)-1)
	})

	t.Run("random", func(t *testing.T) {
		p, disks := setup()
		run(t, p, disks, rand.Intn(len(disks)))
	})
}
