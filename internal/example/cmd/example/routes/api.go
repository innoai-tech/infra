package routes

import (
	"github.com/octohelm/courier/pkg/courierhttp"

	archivev0routes "github.com/innoai-tech/infra/internal/example/cmd/example/routes/archive/v0"
	orgv0routes "github.com/innoai-tech/infra/internal/example/cmd/example/routes/org/v0"
	archivedomain "github.com/innoai-tech/infra/internal/example/domain/archive"
	orgdomain "github.com/innoai-tech/infra/internal/example/domain/org"
)

var _ = orgdomain.Service{}
var _ = archivedomain.Service{}

var R = courierhttp.GroupRouter("/api/example").With(
	courierhttp.GroupRouter("/v0").With(
		orgv0routes.R,
		archivev0routes.R,
	),
)
