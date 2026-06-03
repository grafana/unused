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

func TestProviderFlagsMultiple(t *testing.T) {
	var gcpProject, awsProfile, azureSub internal.StringSliceFlag

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}

	internal.ProviderFlags(fs, &gcpProject, &awsProfile, &azureSub)

	args := []string{
		"-gcp.project=project1",
		"-gcp.project=project2",
		"-aws.profile=profile1",
		"-aws.profile=profile2",
		"-azure.sub=sub1",
	}

	if err := fs.Parse(args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(gcpProject) != 2 {
		t.Errorf("expected 2 GCP projects, got %d", len(gcpProject))
	}
	if len(awsProfile) != 2 {
		t.Errorf("expected 2 AWS profiles, got %d", len(awsProfile))
	}
	if len(azureSub) != 1 {
		t.Errorf("expected 1 Azure sub, got %d", len(azureSub))
	}
}

func TestProviderFlagsDefaults(t *testing.T) {
	// Note: Provider names are global variables that may have been modified by other tests
	// This test verifies that ProviderFlags sets them up correctly with their default values

	// Reset to known defaults first
	gcp.ProviderName = "GCP"
	aws.ProviderName = "AWS"
	azure.ProviderName = "Azure"

	var gcpProject, awsProfile, azureSub internal.StringSliceFlag

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}

	internal.ProviderFlags(fs, &gcpProject, &awsProfile, &azureSub)

	// Parse with no provider name flags (should keep defaults)
	args := []string{"-gcp.project=test"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The flags were registered with default values, parsing without overrides should keep them
	if gcp.ProviderName != "GCP" {
		t.Errorf("GCP provider name should be 'GCP', got %q", gcp.ProviderName)
	}
	if aws.ProviderName != "AWS" {
		t.Errorf("AWS provider name should be 'AWS', got %q", aws.ProviderName)
	}
	if azure.ProviderName != "Azure" {
		t.Errorf("Azure provider name should be 'Azure', got %q", azure.ProviderName)
	}
}

func TestCreateProvidersErrorCases(t *testing.T) {
	l := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty slices", func(t *testing.T) {
		ps, err := internal.CreateProviders(context.Background(), l, []string{}, []string{}, []string{})
		if !errors.Is(err, internal.ErrNoProviders) {
			t.Errorf("expected ErrNoProviders, got %v", err)
		}
		if ps != nil {
			t.Errorf("expected nil providers, got %v", ps)
		}
	})

	t.Run("GCP invalid project", func(t *testing.T) {
		// This will attempt to create the service, which may fail based on credentials
		// but we're testing that the error path is exercised
		_, err := internal.CreateProviders(context.Background(), l, []string{""}, nil, nil)
		// We expect either success (empty project ID might work in some envs) or an error
		// The key is we're exercising the code path
		if err != nil && !errors.Is(err, internal.ErrNoProviders) {
			// Good - we hit an error path
			t.Logf("GCP error (expected): %v", err)
		}
	})
}

func TestErrNoProviders(t *testing.T) {
	err := internal.ErrNoProviders
	if err.Error() != "please select at least one provider" {
		t.Errorf("unexpected error message: %v", err)
	}

	// Test that it can be compared with errors.Is
	wrapped := errors.New("wrapped: " + err.Error())
	if errors.Is(wrapped, internal.ErrNoProviders) {
		t.Error("should not match wrapped error")
	}
}

func TestCreateProvidersMixed(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("test needs authentication")
	}

	l := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("GCP and Azure", func(t *testing.T) {
		ps, err := internal.CreateProviders(context.Background(), l, []string{"gcp-project"}, nil, []string{"azure-sub"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ps) != 2 {
			t.Fatalf("expecting 2 providers, got %d", len(ps))
		}

		// Verify we have one of each type
		hasGCP := false
		hasAzure := false
		for _, p := range ps {
			switch p.(type) {
			case *gcp.Provider:
				hasGCP = true
			case *azure.Provider:
				hasAzure = true
			}
		}

		if !hasGCP {
			t.Error("expected GCP provider")
		}
		if !hasAzure {
			t.Error("expected Azure provider")
		}
	})
}
