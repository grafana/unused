package main

import (
	"time"

	"github.com/grafana/unused/cmd/internal"
	"github.com/inkel/logfmt"
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
		Timeout time.Duration
	}

	Logger *logfmt.Logger
}
