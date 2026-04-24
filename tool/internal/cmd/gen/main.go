package main

import (
	"context"
	"os"

	"github.com/octohelm/gengo/pkg/gengo"

	"github.com/octohelm/x/logr"
	"github.com/octohelm/x/logr/slog"
)

import (
	_ "github.com/octohelm/courier/devpkg/clientgen"
	_ "github.com/octohelm/courier/devpkg/injectablegen"
	_ "github.com/octohelm/courier/devpkg/operatorgen"
	_ "github.com/octohelm/enumeration/devpkg/enumgen"
	_ "github.com/octohelm/gengo/devpkg/deepcopygen"
	_ "github.com/octohelm/gengo/devpkg/runtimedocgen"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	c, err := gengo.NewContext(&gengo.GeneratorArgs{
		Entrypoint: []string{
			cwd,
		},
		Globals: map[string][]string{
			"gengo:runtimedoc": {},
		},
		OutputFileBaseName: "zz_generated",
	})
	if err != nil {
		panic(err)
	}

	ctx := logr.WithLogger(context.Background(), slog.Logger(slog.Default()))

	if err := c.Execute(ctx, gengo.GetRegisteredGenerators()...); err != nil {
		panic(err)
	}
}
