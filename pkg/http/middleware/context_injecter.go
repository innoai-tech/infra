package middleware

import (
	"net/http"

	"github.com/innoai-tech/infra/pkg/configuration"
)

// ContextInjectorMiddleware 创建一个将配置注入器应用到每个请求上下文的 HTTP 中间件。
func ContextInjectorMiddleware(injector configuration.ContextInjector) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if injector != nil {
				req = req.WithContext(injector.InjectContext(req.Context()))
			}
			next.ServeHTTP(rw, req)
		})
	}
}
