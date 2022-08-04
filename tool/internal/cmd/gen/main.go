package main

import (
	"context"

	"github.com/go-courier/logr"
	"github.com/octohelm/gengo/pkg/gengo"

	_ "github.com/octohelm/courier/devpkg/clientgen"
	_ "github.com/octohelm/courier/devpkg/operatorgen"
	_ "github.com/octohelm/gengo/devpkg/runtimedocgen"
	_ "github.com/octohelm/storage/devpkg/enumgen"
)

func main() {
	c, err := gengo.NewContext(&gengo.GeneratorArgs{
		Globals: map[string][]string{
			"gengo:runtimedoc": {},
		},
		Entrypoint: []string{
			"./cmd/example",
			"./cmd/example/apis/org",
		},
		OutputFileBaseName: "zz_generated",
	})
	if err != nil {
		panic(err)
	}

	ctx := logr.WithLogger(context.Background(), logr.StdLogger())
	if err := c.Execute(ctx, gengo.GetRegisteredGenerators()...); err != nil {
		panic(err)
	}
}
