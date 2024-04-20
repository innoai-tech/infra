package http

import (
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"

	"github.com/go-courier/logr"
	openapiview "github.com/innoai-tech/openapi-playground"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp/handler"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/innoai-tech/infra/internal/otel"
	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/innoai-tech/infra/pkg/configuration"
	"github.com/innoai-tech/infra/pkg/http/middleware"
	"github.com/innoai-tech/infra/pkg/http/webapp"
)

func init() {
	httprouter.SetOpenAPIViewContents(&openapiView{})
}

type openapiView struct {
	once    sync.Once
	handler http.Handler
}

func (v *openapiView) Upgrade(w http.ResponseWriter, r *http.Request) error {
	v.once.Do(func() {
		basePath := strings.Split(r.URL.Path, "/_view/")[0]

		v.handler = webapp.ServeFS(
			openapiview.Contents,
			webapp.WithBaseHref(basePath+"/_view/"),
			webapp.WithAppConfig(map[string]string{
				"OPENAPI": base64.StdEncoding.EncodeToString([]byte(basePath + "/")),
			}),
		)
	})

	v.handler.ServeHTTP(w, r)

	return nil
}

type Server struct {
	// Listen addr
	Addr string `flag:",omitempty,expose=http"`
	// Enable debug mode
	EnableDebug bool `flag:",omitempty"`

	corsOptions []middleware.CORSOption

	root courier.Router
	svc  *http.Server

	routerHandlers []handler.HandlerMiddleware
	globalHandlers []handler.HandlerMiddleware
}

func (s *Server) SetDefaults() {
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

func (s *Server) ApplyRouterHandlers(handlers ...handler.HandlerMiddleware) {
	s.routerHandlers = append(s.routerHandlers, handlers...)
}

func (s *Server) ApplyGlobalHandlers(handlers ...handler.HandlerMiddleware) {
	s.globalHandlers = append(s.globalHandlers, handlers...)
}

func (s *Server) Init(ctx context.Context) error {
	if s.svc != nil {
		return nil
	}

	if s.root == nil {
		return fmt.Errorf("root router is not set")
	}

	info := cli.InfoFromContext(ctx)

	h, err := httprouter.New(
		s.root,
		info.App.String(),
		append(
			[]handler.HandlerMiddleware{
				middleware.ContextInjectorMiddleware(configuration.ContextInjectorFromContext(ctx)),
				middleware.LogAndMetricHandler(),
			},
			s.routerHandlers...,
		)...,
	)

	if err != nil {
		return err
	}

	globalHandlers := append([]handler.HandlerMiddleware{
		middleware.CompressLevelHandlerMiddleware(gzip.DefaultCompression),
		middleware.DefaultCORS(s.corsOptions...),
		middleware.MetricHandler(otel.GathererContext.From(ctx)),
		middleware.PProfHandler(s.EnableDebug),
	}, s.globalHandlers...)

	globalHandlers = append(globalHandlers, middleware.HealthCheckHandler())

	h = handler.ApplyHandlerMiddlewares(globalHandlers...)(h)

	s.svc = &http.Server{
		Addr:    s.Addr,
		Handler: h2c.NewHandler(h, &http2.Server{}),
	}

	return nil
}

func (s *Server) Serve(ctx context.Context) error {
	if s.svc == nil {
		return nil
	}

	logr.FromContext(ctx).Info("serve on %s (%s/%s)", s.svc.Addr, runtime.GOOS, runtime.GOARCH)

	return s.svc.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.svc == nil {
		return nil
	}
	return s.svc.Shutdown(ctx)
}
