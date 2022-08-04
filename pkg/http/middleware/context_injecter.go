package middleware

import (
	"net/http"

	"github.com/innoai-tech/infra/pkg/configuration"
)

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
