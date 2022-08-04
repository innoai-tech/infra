package ui

import (
	"embed"
	"io/fs"

	"github.com/innoai-tech/infra/pkg/http/webapp"
)

//go:embed dist
var content embed.FS

var root, _ = fs.Sub(content, "dist")

var UI = webapp.ServeFS(root, webapp.WithBaseHref("/example"))
