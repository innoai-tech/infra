package main

import (
	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/octohelm/gengo/pkg/format"
)

func init() {
	cli.AddTo(App, &Fmt{})
}

type Fmt struct {
	cli.C

	format.Project
}
