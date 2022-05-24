package clicommon_test

import (
	"errors"
	"testing"

	"github.com/grafana/unused/cmd/clicommon"
)

func TestFilterFlag(t *testing.T) {
	t.Run("String()", func(t *testing.T) {
		tests := []struct {
			f   *clicommon.FilterFlag
			exp string
		}{
			{&clicommon.FilterFlag{Key: "foo"}, "foo="},
			{&clicommon.FilterFlag{Key: "bar", Value: "quux"}, "bar=quux"},
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
			exp *clicommon.FilterFlag
			err error
		}{
			{"foo", &clicommon.FilterFlag{Key: "foo"}, nil},
			{"bar=", &clicommon.FilterFlag{Key: "bar"}, nil},
			{"quux=baz", &clicommon.FilterFlag{Key: "quux", Value: "baz"}, nil},
			{"=foo", nil, clicommon.ErrFilterFlagMissingKey},
			{"=", nil, clicommon.ErrFilterFlagMissingKey},
			{"", nil, clicommon.ErrFilterFlagMissingKey},
		}

		for _, tt := range tests {
			t.Run(tt.in, func(t *testing.T) {
				f := &clicommon.FilterFlag{}

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
