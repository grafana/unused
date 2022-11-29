package internal

import "strings"

type StringSliceFlag []string

func (s *StringSliceFlag) String() string { return strings.Join(*s, ",") }

func (s *StringSliceFlag) Set(v string) error {
	*s = append(*s, v)
	return nil
}
