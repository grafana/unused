package unusedtest_test

import (
	"context"
	"testing"

	"github.com/grafana/unused-pds/pkg/unused"
	"github.com/grafana/unused-pds/pkg/unused/unusedtest"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name  string
		disks unused.Disks
		meta  unused.Meta
	}{
		{"no-disks", nil, nil},
		{"with metadata", nil, map[string]string{"foo": "bar"}},
	}

	for _, tt := range tests {
		ctx := context.Background()
		p := unusedtest.NewProvider(tt.name, tt.meta, tt.disks...)

		if p.Name() != tt.name {
			t.Errorf("unexpected provider.Name() %q", p.Name())
		}

		meta := p.Meta()
		if meta == nil {
			t.Error("expecting metadata, got nil")
		}

		if exp, got := len(tt.meta), len(meta); exp != got {
			t.Errorf("expecting %d metadata value, got %d", exp, got)
		}
		for k, v := range tt.meta {
			if exp, got := v, meta[k]; exp != got {
				t.Errorf("expecting metadata %q with value %q, got %q", k, exp, got)
			}
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
