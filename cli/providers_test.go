package cli_test

import (
	"context"
	"errors"
	"testing"

	"github.com/grafana/unused/aws"
	"github.com/grafana/unused/azure"
	"github.com/grafana/unused/cli"
	"github.com/grafana/unused/gcp"
)

func TestCreateProviders(t *testing.T) {
	t.Run("fail when no provider is given", func(t *testing.T) {
		ps, err := cli.CreateProviders(context.Background(), nil, nil, nil)

		if !errors.Is(err, cli.ErrNoProviders) {
			t.Fatalf("expecting error %v, got %v", cli.ErrNoProviders, err)
		}
		if ps != nil {
			t.Fatalf("expecting nil providers, got %v", ps)
		}
	})

	t.Run("GCP", func(t *testing.T) {
		ps, err := cli.CreateProviders(context.Background(), []string{"foo", "bar"}, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ps) != 2 {
			t.Fatalf("expecting 2 providers, got %d", len(ps))
		}

		for _, p := range ps {
			if _, ok := p.(*gcp.Provider); !ok {
				t.Fatalf("expecting *gcp.Provider, got %T", p)
			}
		}
	})

	t.Run("AWS", func(t *testing.T) {
		ps, err := cli.CreateProviders(context.Background(), nil, []string{"foo", "bar"}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ps) != 2 {
			t.Fatalf("expecting 2 providers, got %d", len(ps))
		}

		for _, p := range ps {
			if _, ok := p.(*aws.Provider); !ok {
				t.Fatalf("expecting *aws.Provider, got %T", p)
			}
		}
	})

	t.Run("Azure", func(t *testing.T) {
		ps, err := cli.CreateProviders(context.Background(), nil, nil, []string{"foo", "bar"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ps) != 2 {
			t.Fatalf("expecting 2 providers, got %d", len(ps))
		}

		for _, p := range ps {
			if _, ok := p.(*azure.Provider); !ok {
				t.Fatalf("expecting *azure.Provider, got %T", p)
			}
		}
	})
}
