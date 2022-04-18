package unusedtest_test

import (
	"context"
	"testing"

	"github.com/grafana/unused"
	"github.com/grafana/unused/unusedtest"
)

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
