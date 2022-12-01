package internal_test

import (
	"flag"
	"io"
	"testing"

	"github.com/grafana/unused/cmd/internal"
)

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
