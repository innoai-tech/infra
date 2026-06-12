package middleware

import (
	"net/http"

	"github.com/innoai-tech/infra/pkg/http/compress"
)

// CompressHandlerMiddleware 根据指定压缩级别创建 HTTP 压缩中间件。
func CompressHandlerMiddleware(level int) func(h http.Handler) http.Handler {
	return compress.HandlerLevel(level)
}
