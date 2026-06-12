package middleware

import (
	"net/http"
)

// HealthCheckHandler 创建健康检查中间件，对根路径的 HEAD/GET 请求返回 204。
func HealthCheckHandler() func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return &healthCheckHandler{
			nextHandler: handler,
		}
	}
}

type healthCheckHandler struct {
	nextHandler http.Handler
}

func (h *healthCheckHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if (req.Method == http.MethodHead || req.Method == http.MethodGet) && req.URL.Path == "/" {
		rw.WriteHeader(http.StatusNoContent)
		return
	}
	h.nextHandler.ServeHTTP(rw, req)
}
