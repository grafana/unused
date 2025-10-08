//go:build fake

package internal

import (
	"context"
	"errors"
	"flag"
	"log/slog"

	"github.com/grafana/unused"
	"github.com/grafana/unused/fake"
)

var ErrNoProviders = errors.New("please select at least one provider")

var large, medium, empty bool

func CreateProviders(ctx context.Context, logger *slog.Logger, _, _, _ []string) ([]unused.Provider, error) {
	logger.Warn("Using fake provider")

	var ps []unused.Provider

	if large {
		ps = append(ps, fake.NewProvider("large", 14+23+36))
	}
	if medium {
		ps = append(ps, fake.NewProvider("medium", 14))
	}
	if empty {
		ps = append(ps, fake.NewProvider("empty", 0))
	}

	if len(ps) == 0 {
		return nil, ErrNoProviders
	}

	return ps, nil
}

// ProviderFlags adds the provider configuration flags to the given
// flag set.
func ProviderFlags(fs *flag.FlagSet, _, _, _ *StringSliceFlag) {
	fs.BoolVar(&large, "large", false, "Add a provider with a large number of disks")
	fs.BoolVar(&medium, "medium", true, "Add a provider with a medium number of disks")
	fs.BoolVar(&empty, "empty", false, "Add a provider with no unused disks")
}
