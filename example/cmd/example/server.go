package main

import (
	nethttp "net/http"
	"strings"

	"github.com/innoai-tech/infra/pkg/cli"
	infrahttp "github.com/innoai-tech/infra/pkg/http"
	"github.com/innoai-tech/infra/pkg/otel"

	exampleroutes "example/cmd/example/routes"
	"example/cmd/example/ui"
	archivedomain "example/domain/archive"
	orgdomain "example/domain/org"
)

func init() {
	serve := &Serve{}
	cli.AddTo(App, serve)
	serve.Server.ApplyRouter(exampleroutes.R)
	serve.Server.ApplyGlobalHandlers(func(handler nethttp.Handler) nethttp.Handler {
		return nethttp.HandlerFunc(func(rw nethttp.ResponseWriter, req *nethttp.Request) {
			if strings.HasPrefix(req.URL.Path, "/api/") || strings.HasPrefix(req.URL.Path, "/.sys/") {
				handler.ServeHTTP(rw, req)
				return
			}
			ui.UI.ServeHTTP(rw, req)
		})
	})
}

type Serve struct {
	cli.C `component:"example"`
	otel.Otel
	Server infrahttp.Server

	Orgs     orgdomain.Service
	Archives archivedomain.Service
}
