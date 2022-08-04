package http

import (
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"runtime"

	"github.com/go-courier/logr"

	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/innoai-tech/infra/pkg/configuration"
	"github.com/innoai-tech/infra/pkg/http/middleware"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp/handler"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
)

type Server struct {
	// Listen addr
	Addr string `flag:",omitempty,expose=http"`
	// Enable debug mode
	EnableDebug bool `flag:",omitempty"`

	root courier.Router
	svc  *http.Server
	h    handler.HandlerMiddleware
}

func (s *Server) SetDefaults() {
	if s.Addr == "" {
		s.Addr = ":80"
	}
}

func (s *Server) BindRouter(r courier.Router) {
	s.root = r
}

func (s *Server) ApplyHandler(h handler.HandlerMiddleware) {
	s.h = h
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
		middleware.ContextInjectorMiddleware(configuration.ContextInjectorFromContext(ctx)),
		middleware.LogHandler(),
	)
	if err != nil {
		return err
	}

	h = handler.ApplyHandlerMiddlewares(
		middleware.CompressLevelHandlerMiddleware(gzip.DefaultCompression),
		middleware.DefaultCORS(),
		middleware.HealthCheckHandler(),
		middleware.PProfHandler(s.EnableDebug),
	)(h)

	if s.h != nil {
		h = s.h(h)
	}

	s.svc = &http.Server{
		Addr:    s.Addr,
		Handler: h,
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
