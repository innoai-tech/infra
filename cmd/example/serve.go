package main

import (
	"github.com/innoai-tech/infra/cmd/example/apis"
	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/innoai-tech/infra/pkg/http"
	"github.com/innoai-tech/infra/pkg/otel"
)

func init() {
	s := cli.AddTo(App, &Serve{})
	s.Server.SetRouter(apis.R)
}

// Start serve
type Serve struct {
	cli.C

	otel.Otel
	Server http.Server
}
