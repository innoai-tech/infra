package apis

import (
	"github.com/octohelm/courier/pkg/courierhttp"

	"github.com/innoai-tech/infra/cmd/example/apis/archive"
	"github.com/innoai-tech/infra/cmd/example/apis/org"
)

var R = courierhttp.GroupRouter("/api/example").With(
	courierhttp.GroupRouter("/v0").With(
		org.R,
		archive.R,
	),
)
