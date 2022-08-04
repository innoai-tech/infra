package main

import (
	"context"
	"os"

	"github.com/innoai-tech/infra/pkg/cli"
)

var App = cli.NewApp(
	"example",
	"1.0.0",
	cli.WithImageNamespace("ghcr.io/octohelm"),
)

func main() {
	if err := cli.Execute(context.Background(), App, os.Args[1:]); err != nil {
		panic(err)
	}
}
