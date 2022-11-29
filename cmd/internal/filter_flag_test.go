package internal_test

import (
	"errors"
	"testing"

	"github.com/grafana/unused/cmd/internal"
)

func TestFilterFlag(t *testing.T) {
	t.Run("String()", func(t *testing.T) {
		tests := []struct {
			f   *internal.FilterFlag
			exp string
		}{
			{&internal.FilterFlag{Key: "foo"}, "foo="},
			{&internal.FilterFlag{Key: "bar", Value: "quux"}, "bar=quux"},
		}

		for _, tt := range tests {
			if got := tt.f.String(); got != tt.exp {
				t.Errorf("expecting %#v String() %q, got %q", tt.f, tt.exp, got)
			}
		}
	})

	t.Run("Set()", func(t *testing.T) {
		tests := []struct {
			in  string
			exp *internal.FilterFlag
			err error
		}{
			{"foo", &internal.FilterFlag{Key: "foo"}, nil},
			{"bar=", &internal.FilterFlag{Key: "bar"}, nil},
			{"quux=baz", &internal.FilterFlag{Key: "quux", Value: "baz"}, nil},
			{"=foo", nil, internal.ErrFilterFlagMissingKey},
			{"=", nil, internal.ErrFilterFlagMissingKey},
			{"", nil, internal.ErrFilterFlagMissingKey},
		}

		for _, tt := range tests {
			t.Run(tt.in, func(t *testing.T) {
				f := &internal.FilterFlag{}

				if err := f.Set(tt.in); !errors.Is(err, tt.err) {
					t.Fatalf("expecting error %q, got %q", tt.err, err)
				}

				if tt.exp != nil && *f != *tt.exp {
					t.Fatalf("expecting %#v, got %#v", *tt.exp, *f)
				}
			})
		}
	})
}
