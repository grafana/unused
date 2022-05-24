package clicommon

import (
	"errors"
	"flag"
	"strings"
)

var _ flag.Value = &FilterFlag{}

var ErrFilterFlagMissingKey = errors.New("missing key")

type FilterFlag struct {
	Key, Value string
}

func (f *FilterFlag) Set(v string) error {
	ps := strings.SplitN(v, "=", 2)

	if len(ps) == 0 || ps[0] == "" {
		return ErrFilterFlagMissingKey
	}

	f.Key = ps[0]

	if len(ps) == 2 {
		f.Value = ps[1]
	}

	return nil
}

func (f *FilterFlag) String() string {
	return f.Key + "=" + f.Value
}
