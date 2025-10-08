//go:build !fake

package internal_test

import (
	"context"
	"errors"
	"flag"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/grafana/unused/aws"
	"github.com/grafana/unused/azure"
	"github.com/grafana/unused/cmd/internal"
	"github.com/grafana/unused/gcp"
)

func TestCreateProviders(t *testing.T) {
	l := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("fail when no provider is given", func(t *testing.T) {
		ps, err := internal.CreateProviders(context.Background(), l, nil, nil, nil)

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
		ps, err := internal.CreateProviders(context.Background(), l, []string{"foo", "bar"}, nil, nil)
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
		t.Skip("AWS now fails when it cannot find the profile in the configuration")
		ps, err := internal.CreateProviders(context.Background(), l, nil, []string{"foo", "bar"}, nil)
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
		ps, err := internal.CreateProviders(context.Background(), l, nil, nil, []string{"foo", "bar"})
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
	var gcpProject, awsProfile, azureSub internal.StringSliceFlag

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}

	internal.ProviderFlags(fs, &gcpProject, &awsProfile, &azureSub)

	args := []string{
		"-gcp.project=my-project",
		"-azure.sub=my-subscription",
		"-aws.profile=my-profile",
		"-gcp.providername=GKE",
		"-azure.providername=AKS",
		"-aws.providername=EKS",
	}

	if err := fs.Parse(args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testSlices := map[*internal.StringSliceFlag]string{
		&gcpProject: "my-project",
		&awsProfile: "my-profile",
		&azureSub:   "my-subscription",
	}
	testStrings := map[*string]string{
		&gcp.ProviderName:   "GKE",
		&aws.ProviderName:   "EKS",
		&azure.ProviderName: "AKS",
	}

	for v, exp := range testSlices {
		if len(*v) != 1 || (*v)[0] != exp {
			t.Errorf("expecting one value (%q), got %v", exp, v)
		}
	}
	for v, exp := range testStrings {
		if *v != exp {
			t.Errorf("expecting %q, got %v", exp, v)
		}
	}
}
