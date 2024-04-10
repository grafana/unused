package main

import (
	"log/slog"
	"time"

	"github.com/grafana/unused/cmd/internal"
)

type config struct {
	Providers internal.ProviderConfig

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
