package middleware

import (
	"net/http"
	"net/http/pprof"
	"strings"
)

// PProfHandler 创建 pprof 调试中间件，在启用时挂载 /.sys/debug/pprof 路径。
func PProfHandler(enabled bool) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return &pprofHandler{
			enabled:     enabled,
			nextHandler: handler,
		}
	}
}

type pprofHandler struct {
	enabled     bool
	nextHandler http.Handler
}

func (h *pprofHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if h.enabled && strings.HasPrefix(req.URL.Path, "/.sys/debug/pprof") {
		switch req.URL.Path {
		case "/.sys/debug/pprof/cmdline":
			pprof.Cmdline(rw, req)
			return
		case "/.sys/debug/pprof/profile":
			pprof.Profile(rw, req)
			return
		case "/.sys/debug/pprof/symbol":
			pprof.Symbol(rw, req)
			return
		case "/.sys/debug/pprof/trace":
			pprof.Trace(rw, req)
			return
		default:
			// 去掉 /.sys 前缀以让 pprof.Index 正常工作
			req.URL.Path = req.URL.Path[len("/.sys"):]
			pprof.Index(rw, req)
			return
		}
	}
	h.nextHandler.ServeHTTP(rw, req)
}
