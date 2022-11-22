package main

import (
	"time"

	"github.com/grafana/unused/cmd/clicommon"
	"github.com/inkel/logfmt"
)

type config struct {
	Providers struct {
		GCP   clicommon.StringSliceFlag
		AWS   clicommon.StringSliceFlag
		Azure clicommon.StringSliceFlag
	}

	Web struct {
		Address string
		Path    string
	}

	Collector struct {
		Timeout time.Duration
	}

	Logger *logfmt.Logger
}
