package internal_test

import (
	"errors"
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

func TestParseAge(t *testing.T) {
	tests := []struct {
		in  string
		exp time.Duration
		err error
	}{
		{"23d", 23 * 24 * time.Hour, nil},
		{"0d3h", 3 * time.Hour, nil},
		{"d3h", 0, internal.ErrInvalidAge},
		{"23dh", 0, internal.ErrInvalidAge},
		{"23d3", 0, internal.ErrInvalidAge},
		{"", 0, internal.ErrInvalidAge},
		{"3x", 0, internal.ErrInvalidAge},
		{"2x4d6h", 0, internal.ErrInvalidAge},
		{"1y", 365 * 24 * time.Hour, nil},
		{"3y7d14h", 3*365*24*time.Hour + 7*24*time.Hour + 14*time.Hour, nil},
		{"1y2y3d", 0, internal.ErrInvalidAge},
		{"1d2d3h", 0, internal.ErrInvalidAge},
	}

	for _, tt := range tests {
		got, err := internal.ParseAge(tt.in)
		if !errors.Is(err, tt.err) || got != tt.exp {
			t.Errorf("expecting ParseAge(%q) = (%v, %v), got (%v, %v)", tt.in, tt.exp, tt.err, got, err)
		}
		t.Logf("s: %q, dur: %v, err: %v", tt.in, got, err)
	}
}
