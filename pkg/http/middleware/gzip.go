package middleware

import (
	"github.com/innoai-tech/infra/pkg/http/compress"
	"net/http"
)

func CompressLevelHandlerMiddleware(level int) func(h http.Handler) http.Handler {
	return compress.CompressHandlerLevel(level)
}
