package main

import (
	"log/slog"
	"time"

	"github.com/grafana/unused/cmd/internal"
)

type config struct {
	Providers struct {
		GCP   internal.StringSliceFlag
		AWS   internal.StringSliceFlag
		Azure internal.StringSliceFlag
	}

	Web struct {
		Address string
		Path    string
		Timeout time.Duration
	}

	Collector struct {
		Timeout      time.Duration
		PollInterval time.Duration
	}

	Logger         *slog.Logger
	VerboseLogging bool
}
