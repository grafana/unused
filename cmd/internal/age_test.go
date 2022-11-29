package internal_test

import (
	"testing"
	"time"

	"github.com/grafana/unused/cmd/internal"
)

func TestAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		in  time.Time
		exp string
	}{
		{now.Add(-30 * time.Second), "1m"},
		{now.Add(-30 * time.Minute), "30m"},
		{now.Add(-5 * time.Hour), "5h"},
		{now.Add(-7 * 24 * time.Hour), "7d"},
		{now.Add(-364 * 24 * time.Hour), "364d"},
		{now.Add(-600 * 24 * time.Hour), "1y"},
		{now.Add(-740 * 24 * time.Hour), "2y"},
		{time.Time{}, "n/a"},
	}

	for _, tt := range tests {
		t.Run(tt.exp, func(t *testing.T) {
			if got := internal.Age(tt.in); tt.exp != got {
				t.Errorf("expecting Age(%s) = %s, got %s", tt.in.Format(time.RFC3339), tt.exp, got)
			}
		})
	}
}
