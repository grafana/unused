package internal_test

import (
	"context"
	"errors"
	"flag"
	"io"
	"os"
	"testing"

	"github.com/grafana/unused/aws"
	"github.com/grafana/unused/azure"
	"github.com/grafana/unused/cmd/internal"
	"github.com/grafana/unused/gcp"
)

func TestCreateProviders(t *testing.T) {
	t.Run("fail when no provider is given", func(t *testing.T) {
		ps, err := internal.CreateProviders(context.Background(), nil, nil, nil)

		if !errors.Is(err, internal.ErrNoProviders) {
			t.Fatalf("expecting error %v, got %v", internal.ErrNoProviders, err)
		}
		if ps != nil {
			t.Fatalf("expecting nil providers, got %v", ps)
		}
	})

	if os.Getenv("CI") == "true" {
		t.Skip("the following tests need authentication") // TODO
	}

	t.Run("GCP", func(t *testing.T) {
		ps, err := internal.CreateProviders(context.Background(), []string{"foo", "bar"}, nil, nil)
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
		ps, err := internal.CreateProviders(context.Background(), nil, []string{"foo", "bar"}, nil)
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
		ps, err := internal.CreateProviders(context.Background(), nil, nil, []string{"foo", "bar"})
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

func TestProviderFlags(t *testing.T) {
	var gcp, aws, azure internal.StringSliceFlag

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}

	internal.ProviderFlags(fs, &gcp, &aws, &azure)

	args := []string{
		"-gcp.project=my-project",
		"-azure.sub=my-subscription",
		"-aws.profile=my-profile",
	}

	if err := fs.Parse(args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := map[*internal.StringSliceFlag]string{
		&gcp:   "my-project",
		&aws:   "my-profile",
		&azure: "my-subscription",
	}

	for v, exp := range tests {
		if len(*v) != 1 || (*v)[0] != exp {
			t.Errorf("expecting one value (%q), got %v", exp, v)
		}
	}
}
