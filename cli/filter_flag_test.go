package cli_test

import (
	"errors"
	"testing"

	"github.com/grafana/unused/cli"
)

func TestFilterFlag(t *testing.T) {
	t.Run("String()", func(t *testing.T) {
		tests := []struct {
			f   *cli.FilterFlag
			exp string
		}{
			{&cli.FilterFlag{Key: "foo"}, "foo="},
			{&cli.FilterFlag{Key: "bar", Value: "quux"}, "bar=quux"},
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
			exp *cli.FilterFlag
			err error
		}{
			{"foo", &cli.FilterFlag{Key: "foo"}, nil},
			{"bar=", &cli.FilterFlag{Key: "bar"}, nil},
			{"quux=baz", &cli.FilterFlag{Key: "quux", Value: "baz"}, nil},
			{"=foo", nil, cli.ErrFilterFlagMissingKey},
			{"=", nil, cli.ErrFilterFlagMissingKey},
			{"", nil, cli.ErrFilterFlagMissingKey},
		}

		for _, tt := range tests {
			t.Run(tt.in, func(t *testing.T) {
				f := &cli.FilterFlag{}

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
