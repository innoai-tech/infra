package http

import (
	"cmp"
	"compress/gzip"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp/handler"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
	"github.com/octohelm/x/logr"

	"github.com/innoai-tech/infra/pkg/appinfo"
	"github.com/innoai-tech/infra/pkg/configuration"
	"github.com/innoai-tech/infra/pkg/http/middleware"
	otelmetric "github.com/innoai-tech/infra/pkg/otel/metric"
)

// +gengo:injectable
type Server struct {
	// Listen addr
	Addr string `flag:",omitempty,expose=http"`
	// Enable debug mode
	EnableDebug bool `flag:",omitempty"`

	corsOptions []middleware.CORSOption

	name string
	root courier.Router
	svc  *http.Server

	tlsProvider    Provider
	metricReader   sdkmetric.Reader
	globalHandlers []handler.Middleware
	routerHandlers []handler.Middleware

	info *appinfo.Info `inject:",opt"`

	ready    sync.WaitGroup
	endpoint atomic.Pointer[string]
}

// SetDefaults 根据 TLS 配置补齐默认监听地址。
func (s *Server) SetDefaults() {
	if s.tlsProvider != nil {
		if s.Addr == "" {
			s.Addr = ":443"
		}
		return
	}

	if s.Addr == "" {
		s.Addr = ":80"
	}
}

// SetCorsOptions 设置全局 CORS 选项。
func (s *Server) SetCorsOptions(options ...middleware.CORSOption) {
	s.corsOptions = options
}

// ApplyRouter 绑定 courier 路由树。
func (s *Server) ApplyRouter(r courier.Router) {
	s.root = r
}

// SetName 覆盖默认服务名。
func (s *Server) SetName(name string) {
	s.name = name
}

// SetTLSProvoder 设置 TLS 配置提供者。
func (s *Server) SetTLSProvoder(tlsProvider Provider) {
	s.tlsProvider = tlsProvider
}

// SetMetricReader 显式设置指标 reader，避免完全依赖上下文注入。
func (s *Server) SetMetricReader(reader sdkmetric.Reader) {
	s.metricReader = reader
}

// ApplyRouterHandlers 为业务路由追加中间件。
func (s *Server) ApplyRouterHandlers(handlers ...handler.Middleware) {
	s.routerHandlers = append(s.routerHandlers, handlers...)
}

// ApplyGlobalHandlers 为整个 HTTP 服务追加中间件。
func (s *Server) ApplyGlobalHandlers(handlers ...handler.Middleware) {
	s.globalHandlers = append(s.globalHandlers, handlers...)
}

func (s *Server) serviceName(ctx context.Context) string {
	if s.info == nil {
		if value, ok := appinfo.InfoFromContext(ctx); ok {
			s.info = value
		}
	}

	if info := s.info; info != nil {
		return cmp.Or(s.name, info.App.Name) + "/" + info.App.Version
	}

	return "unknown/v0"
}

// NewHandler 基于给定路由构建最终 HTTP handler。
func (s *Server) NewHandler(ctx context.Context, root courier.Router) (http.Handler, error) {
	return httprouter.New(
		root,
		s.serviceName(ctx),
		s.buildRouterHandlers(ctx)...,
	)
}

func (s *Server) buildRouterHandlers(ctx context.Context) []handler.Middleware {
	return slices.Concat(
		[]handler.Middleware{
			middleware.ContextInjectorMiddleware(configuration.ContextInjectorFromContext(ctx)),
			middleware.CompressHandlerMiddleware(gzip.DefaultCompression),
			middleware.LogAndMetricHandler(),
		},
		s.routerHandlers,
	)
}

func (s *Server) afterInit(ctx context.Context) error {
	if s.svc != nil {
		return nil
	}

	metricReader := s.metricReader
	if metricReader == nil {
		if injected, ok := otelmetric.ReaderContext.MayFrom(ctx); ok {
			metricReader = injected
		} else {
			metricReader = sdkmetric.NewManualReader()
		}
	}

	var r http.Handler = http.NewServeMux()

	if s.root != nil {
		h, err := s.NewHandler(ctx, s.root)
		if err != nil {
			return err
		}
		r = h
	}

	globalHandlers := slices.Concat(
		[]handler.Middleware{
			middleware.MetricHandler(metricReader),
			middleware.DefaultCORS(s.corsOptions...),
			middleware.PProfHandler(s.EnableDebug),
		},
		s.globalHandlers,
		[]handler.Middleware{
			middleware.HealthCheckHandler(),
		},
	)

	s.svc = &http.Server{
		Addr:              s.Addr,
		ReadHeaderTimeout: 30 * time.Second,
		Handler:           h2c.NewHandler(handler.ApplyMiddlewares(globalHandlers...)(r), &http2.Server{}),
	}

	// for listen waiting
	s.ready.Add(1)

	return nil
}

// Endpoint 返回实际监听成功后的对外地址。
func (s *Server) Endpoint() string {
	s.ready.Wait()

	if v := s.endpoint.Load(); v != nil {
		return *v
	}
	return ""
}

// Serve 启动 HTTP 服务并记录最终监听地址。
func (s *Server) Serve(ctx context.Context) error {
	if s.svc == nil {
		return nil
	}

	l := logr.FromContext(ctx)

	svc := s.svc

	proto := "http"
	if s.tlsProvider != nil {
		proto = "https"
	}

	host, port, err := net.SplitHostPort(svc.Addr)
	if err != nil {
		return nil
	}
	if host == "" {
		host = "0.0.0.0"
	}

	ln, err := net.Listen("tcp", svc.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	if port == "0" {
		_, p, _ := net.SplitHostPort(ln.Addr().String())
		port = p
	}

	addr := net.JoinHostPort(host, port)

	l.Info("serve %s %s://%s (%s/%s)", s.serviceName(ctx), proto, addr, runtime.GOOS, runtime.GOARCH)

	s.endpoint.Store(new(fmt.Sprintf("%s://%s", proto, addr)))

	s.ready.Done()

	if s.tlsProvider != nil {
		svc.TLSConfig = s.tlsProvider.TLSConfig()
		return svc.ServeTLS(ln, "", "")
	}

	return svc.Serve(ln)
}

// Shutdown 优雅关闭底层 HTTP 服务。
func (s *Server) Shutdown(ctx context.Context) error {
	if s.svc == nil {
		return nil
	}
	return s.svc.Shutdown(ctx)
}

// Provider 表示可提供 TLS 配置的对象。
type Provider interface {
	TLSConfig() *tls.Config
}
