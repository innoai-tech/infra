package main

import (
	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/octohelm/gengo/pkg/gengo"
)

import (
	_ "github.com/innoai-tech/infra/tool/go/gen"
)

func init() {
	cli.AddTo(App, &Gen{})
}

type Gen struct {
	cli.C

	gengo.Project
}
