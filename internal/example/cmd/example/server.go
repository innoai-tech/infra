package main

import (
	nethttp "net/http"
	"strings"

	exampleroutes "github.com/innoai-tech/infra/internal/example/cmd/example/routes"
	"github.com/innoai-tech/infra/internal/example/cmd/example/ui"
	archivedomain "github.com/innoai-tech/infra/internal/example/domain/archive"
	orgdomain "github.com/innoai-tech/infra/internal/example/domain/org"
	"github.com/innoai-tech/infra/pkg/cli"
	infrahttp "github.com/innoai-tech/infra/pkg/http"
	"github.com/innoai-tech/infra/pkg/otel"
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
	cli.C `component:"server"`
	otel.Otel
	Server infrahttp.Server

	Orgs     orgdomain.Service
	Archives archivedomain.Service
}
