package routes

import (
	"github.com/octohelm/courier/pkg/courierhttp"

	archivev0routes "example/cmd/example/routes/archive/v0"
	orgv0routes "example/cmd/example/routes/org/v0"
	archivedomain "example/domain/archive"
	orgdomain "example/domain/org"
)

var (
	_ = orgdomain.Service{}
	_ = archivedomain.Service{}
)

var R = courierhttp.GroupRouter("/api/example").With(
	courierhttp.GroupRouter("/v0").With(
		orgv0routes.R,
		archivev0routes.R,
	),
)
