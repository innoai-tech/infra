package main

import (
	"github.com/innoai-tech/infra/devpkg/workrun"
	"github.com/innoai-tech/infra/pkg/cli"
)

func init() {
	cli.AddTo(App, &Workrun{})
}

type Workrun struct {
	cli.C

	workrun.Runner
}
