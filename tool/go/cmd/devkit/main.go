package main

import (
	"context"

	"github.com/innoai-tech/infra/pkg/cli"
)

var App = cli.NewApp("devkit", "devel")

func main() {
	cli.Exec(context.Background(), App)
}
