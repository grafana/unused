package unusedtest_test

import (
	"testing"

	"github.com/grafana/unused"
	"github.com/grafana/unused/unusedtest"
)

func TestAssertEqualMeta(t *testing.T) {
	var err error

	err = unusedtest.AssertEqualMeta(unused.Meta{"foo": "bar"}, unused.Meta{"baz": "quux", "lorem": "ipsum"})
	if err == nil {
		t.Fatal("expecting error with different metadata lengths")
	}

	err = unusedtest.AssertEqualMeta(unused.Meta{"foo": "bar"}, unused.Meta{"foo": "quux"})
	if err == nil {
		t.Fatal("expecting error with different metadata value")
	}

	err = unusedtest.AssertEqualMeta(unused.Meta{"foo": "bar"}, unused.Meta{"lorem": "bar"})
	if err == nil {
		t.Fatal("expecting error with different metadata")
	}

	err = unusedtest.AssertEqualMeta(unused.Meta{"foo": "bar"}, unused.Meta{"foo": "bar"})
	if err != nil {
		t.Fatalf("unexpected error with equal metadata: %v", err)
	}
}
