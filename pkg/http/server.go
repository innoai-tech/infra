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
	"time"

	"github.com/go-courier/logr"
	"github.com/innoai-tech/infra/internal/otel"
	"github.com/innoai-tech/infra/pkg/appinfo"
	"github.com/innoai-tech/infra/pkg/configuration"
	"github.com/innoai-tech/infra/pkg/http/middleware"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp/handler"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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
	globalHandlers []handler.Middleware
	routerHandlers []handler.Middleware

	info *appinfo.Info `inject:",opt"`
}

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

func (s *Server) SetCorsOptions(options ...middleware.CORSOption) {
	s.corsOptions = options
}

func (s *Server) ApplyRouter(r courier.Router) {
	s.root = r
}

func (s *Server) SetName(name string) {
	s.name = name
}

func (s *Server) SetTLSProvoder(tlsProvider Provider) {
	s.tlsProvider = tlsProvider
}

func (s *Server) ApplyRouterHandlers(handlers ...handler.Middleware) {
	s.routerHandlers = append(s.routerHandlers, handlers...)
}

func (s *Server) ApplyGlobalHandlers(handlers ...handler.Middleware) {
	s.globalHandlers = append(s.globalHandlers, handlers...)
}

func (s *Server) serviceName() string {
	if info := s.info; info != nil {
		return cmp.Or(s.name, info.App.Name) + "/" + info.App.Version
	}
	return "unknown/v0"
}

func (s *Server) afterInit(ctx context.Context) error {
	if s.svc != nil {
		return nil
	}

	if s.root == nil {
		return fmt.Errorf("root router is not set")
	}

	h, err := httprouter.New(
		s.root,
		s.serviceName(),
		append(
			[]handler.Middleware{
				middleware.ContextInjectorMiddleware(configuration.ContextInjectorFromContext(ctx)),
				middleware.CompressHandlerMiddleware(gzip.DefaultCompression),
				middleware.LogAndMetricHandler(),
			},
			s.routerHandlers...,
		)...,
	)

	if err != nil {
		return err
	}

	globalHandlers := append([]handler.Middleware{
		middleware.MetricHandler(otel.GathererContext.From(ctx)),
		middleware.DefaultCORS(s.corsOptions...),
		middleware.PProfHandler(s.EnableDebug),
	}, s.globalHandlers...)

	globalHandlers = append(globalHandlers, middleware.HealthCheckHandler())

	h = handler.ApplyMiddlewares(globalHandlers...)(h)

	s.svc = &http.Server{
		Addr:              s.Addr,
		ReadHeaderTimeout: 30 * time.Second,
		Handler:           h2c.NewHandler(h, &http2.Server{}),
	}

	return nil
}

func (s *Server) Serve(ctx context.Context) error {
	if s.svc == nil {
		return nil
	}

	l := logr.FromContext(ctx)

	svc := s.svc

	tpe := "http"
	if s.tlsProvider != nil {
		tpe = "https"
	}

	l.Info("serve %s %s on %s (%s/%s)", s.serviceName(), tpe, svc.Addr, runtime.GOOS, runtime.GOARCH)

	ln, err := net.Listen("tcp", svc.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	if s.tlsProvider != nil {
		svc.TLSConfig = s.tlsProvider.TLSConfig()
		return svc.ServeTLS(ln, "", "")
	}

	return svc.Serve(ln)
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.svc == nil {
		return nil
	}
	return s.svc.Shutdown(ctx)
}

type Provider interface {
	TLSConfig() *tls.Config
}
