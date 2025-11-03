package main

import (
	"context"
	"os"

	"github.com/octohelm/gengo/pkg/format"

	"github.com/innoai-tech/infra/devpkg/gengo"
	"github.com/innoai-tech/infra/pkg/cli"
)

import (
	_ "github.com/octohelm/courier/devpkg/clientgen"
	_ "github.com/octohelm/courier/devpkg/injectablegen"
	_ "github.com/octohelm/courier/devpkg/operatorgen"
	_ "github.com/octohelm/enumeration/devpkg/enumgen"
	_ "github.com/octohelm/gengo/devpkg/runtimedocgen"
)

var App = cli.NewApp("devtool", "dev")

func init() {
	cli.AddTo(App, &struct {
		cli.C `name:"gen"`
		gengo.Gengo
	}{})

	cli.AddTo(App, &struct {
		cli.C `name:"fmt"`
		format.Project
	}{})
}

func main() {
	if err := cli.Execute(context.Background(), App, os.Args[1:]); err != nil {
		panic(err)
	}
}
