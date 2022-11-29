package internal

import (
	"flag"
	"testing"
)

func TestStringSliceFlag(t *testing.T) {
	var vs StringSliceFlag

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Var(&vs, "ss", "test")

	if err := fs.Parse([]string{"-ss", "foo", "-ss=bar"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if exp, got := 2, len(vs); exp != got {
		t.Errorf("expecting %d values, got %d", exp, got)
		t.Errorf("%q", vs)
	}
}

func TestStringSliceSet(t *testing.T) {
	ss := &StringSliceFlag{}

	for _, v := range []string{"foo", "bar"} {
		if err := ss.Set(v); err != nil {
			t.Fatalf("unexpected error setting %s: %v", v, err)
		}
	}

	if exp, got := "foo,bar", ss.String(); exp != got {
		t.Errorf("expecting %q, got %q", exp, got)
	}
}
