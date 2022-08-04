package main

import (
	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/innoai-tech/infra/pkg/http/webapp"
	"github.com/innoai-tech/infra/pkg/otel"
)

func init() {
	cli.AddTo(App, &Webapp{})
}

// Start webapp serve
type Webapp struct {
	cli.C `component:"webapp"`
	otel.Otel
	webapp.Server
}
