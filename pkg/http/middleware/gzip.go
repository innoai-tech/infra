package middleware

import (
	"net/http"

	"github.com/innoai-tech/infra/pkg/http/compress"
)

func CompressHandlerMiddleware(level int) func(h http.Handler) http.Handler {
	return compress.HandlerLevel(level)
}
