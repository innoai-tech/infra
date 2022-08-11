package middleware

import (
	"net/http"

	"github.com/innoai-tech/infra/pkg/http/compress"
)

func CompressLevelHandlerMiddleware(level int) func(h http.Handler) http.Handler {
	return compress.CompressHandlerLevel(level)
}
