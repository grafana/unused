package unusedtest_test

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/grafana/unused"
	"github.com/grafana/unused/unusedtest"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		disks unused.Disks
	}{
		{"no-disks", "my-id", nil},
	}

	for _, tt := range tests {
		ctx := context.Background()
		p := unusedtest.NewProvider(tt.name, nil, tt.disks...)

		if p.Name() != tt.name {
			t.Errorf("unexpected provider.Name() %q", p.Name())
		}

		if p.ID() != tt.id {
			t.Errorf("unexpected provider.ID() %q", p.ID())
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
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		run(t, p, disks, r.Intn(len(disks)))
	})

	t.Run("not found", func(t *testing.T) {
		p, _ := setup()

		if err := p.Delete(ctx, unusedtest.NewDisk("foo-bar-baz", p, time.Now())); !errors.Is(err, unusedtest.ErrDiskNotFound) {
			t.Fatalf("expecting error %v, got %v", unusedtest.ErrDiskNotFound, err)
		}
	})
}

func TestTestProviderMeta(t *testing.T) {
	t.Run("fail to create provider", func(t *testing.T) {
		err := unusedtest.TestProviderMeta(func(unused.Meta) (unused.Provider, error) {
			return nil, errors.New("foo")
		})
		if err == nil {
			t.Fatal("expecting error")
		}
	})

	t.Run("returns nil metadata", func(t *testing.T) {
		err := unusedtest.TestProviderMeta(func(unused.Meta) (unused.Provider, error) {
			p := unusedtest.NewProvider("my-provider", nil)
			p.SetMeta(nil)
			return p, nil
		})
		if err == nil {
			t.Fatal("expecting error")
		}
	})

	t.Run("returns different metadata length", func(t *testing.T) {
		err := unusedtest.TestProviderMeta(func(meta unused.Meta) (unused.Provider, error) {
			// ensure we are always sending at least twice the length
			newMeta := make(unused.Meta)
			for k, v := range meta {
				newMeta[k] = v
				newMeta[v] = k
			}
			return unusedtest.NewProvider("my-provider", newMeta), nil
		})
		if err == nil {
			t.Fatal("expecting error")
		}
	})

	t.Run("metadata is unchanged", func(t *testing.T) {
		err := unusedtest.TestProviderMeta(func(meta unused.Meta) (unused.Provider, error) {
			// ensure we are always sending at least twice the length
			newMeta := make(unused.Meta)
			for k := range meta {
				newMeta[k] = k
			}
			return unusedtest.NewProvider("my-provider", newMeta), nil
		})
		if err == nil {
			t.Fatal("expecting error")
		}
	})

	t.Run("passes all testes", func(t *testing.T) {
		err := unusedtest.TestProviderMeta(func(meta unused.Meta) (unused.Provider, error) {
			return unusedtest.NewProvider("my-provider", meta), nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
