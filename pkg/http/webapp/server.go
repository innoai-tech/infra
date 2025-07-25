package webapp

import (
	"bytes"
	"cmp"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/innoai-tech/infra/pkg/http/compress"
	"github.com/octohelm/courier/pkg/courierhttp/handler"

	"github.com/go-courier/logr"
	"github.com/innoai-tech/infra/pkg/http/webapp/appconfig"

	_ "github.com/innoai-tech/infra/pkg/http/webapp/etc"
)

type Server struct {
	// app env name
	Env string `flag:",omitempty"`
	// app display version
	Ver string `flag:",omitempty"`
	// base href
	BaseHref string `flag:",omitempty"`
	// config
	Config string `flag:",omitempty"`
	// Disable http history fallback, only used for static pages
	DisableHistoryFallback bool `flag:",omitempty"`
	// Disable Content-Security-Policy
	DisableCSP bool `flag:",omitempty"`
	// AppRoot for host in fs
	Root string `flag:",omitempty"`
	// Webapp serve on
	Addr string `flag:",omitempty,expose=http"`

	fs fs.FS

	svc *http.Server
}

func (s *Server) BindFS(f fs.FS) {
	s.fs = f
}

func (s *Server) SetDefaults() {
	if s.BaseHref == "" {
		s.BaseHref = "/"
	}

	if s.Addr == "" {
		s.Addr = ":80"
	}

	if s.Env == "" {
		s.Env = os.Getenv("ENV")
	}
}

func (s *Server) Init(ctx context.Context) error {
	if s.svc != nil {
		return nil
	}

	if s.fs == nil {
		s.fs = os.DirFS(s.Root)
		if _, err := fs.Stat(s.fs, "index.html"); err != nil {
			return fmt.Errorf("index.html not found in root dir %s: %w", s.Root, err)
		}
	}

	ac := appconfig.ParseAppConfig(s.Config)
	ac.LoadFromEnviron(os.Environ())

	s.svc = &http.Server{
		Addr:              s.Addr,
		ReadHeaderTimeout: 10 * time.Second,
		Handler: ServeFS(
			s.fs,
			WithAppEnv(s.Env),
			WithAppVersion(s.Ver),
			WithAppConfig(ac),
			WithBaseHref(s.BaseHref),
			DisableHistoryFallback(s.DisableHistoryFallback),
			DisableCSP(s.DisableCSP),
		),
	}

	return nil
}

func (s *Server) Serve(ctx context.Context) error {
	if s.svc == nil {
		return nil
	}
	l := logr.FromContext(ctx)
	if s.Root != "" {
		l = l.WithValues("staticRoot", s.Root)
	}
	l.Info("serve on %s (%s/%s)", s.svc.Addr, runtime.GOOS, runtime.GOARCH)
	return s.svc.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.svc.Shutdown(ctx)
}

type OptFunc func(o *opt)

func WithAppConfig(appConfig map[string]string) OptFunc {
	return func(o *opt) {
		o.appConfig = appConfig
	}
}

func WithAppEnv(appEnv string) OptFunc {
	return func(o *opt) {
		o.appEnv = appEnv
	}
}

func WithAppVersion(appVersion string) OptFunc {
	return func(o *opt) {
		o.appVersion = appVersion
	}
}

func WithBaseHref(baseHref string) OptFunc {
	return func(o *opt) {
		o.baseHref = baseHref

		if !strings.HasSuffix(o.baseHref, "/") {
			o.baseHref = o.baseHref + "/"
		}
	}
}

func DisableHistoryFallback(disableHistoryFallback bool) OptFunc {
	return func(o *opt) {
		o.disableHistoryFallback = disableHistoryFallback
	}
}

func DisableCSP(disableCSP bool) OptFunc {
	return func(o *opt) {
		o.disableCSP = disableCSP
	}
}

type opt struct {
	appEnv                 string
	appVersion             string
	appConfig              appconfig.AppConfig
	baseHref               string
	disableHistoryFallback bool
	disableCSP             bool

	processed sync.Map
}

func (o *opt) loadOrProcess(f fs.FS, path string, baseHref string) ([]byte, error) {
	fn, _ := o.processed.LoadOrStore(path+"?baseHref="+baseHref, func() ([]byte, error) {
		file, err := f.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}

		data = bytes.ReplaceAll(data, []byte("__ENV__"), []byte(o.appEnv))
		data = bytes.ReplaceAll(data, []byte("__VERSION__"), []byte(o.appVersion))
		data = bytes.ReplaceAll(data, []byte("__APP_CONFIG__"), []byte(o.appConfig.String()))

		data = bytes.ReplaceAll(data, []byte("/__APP_BASE_HREF__/"), []byte(baseHref))
		data = bytes.ReplaceAll(data, []byte("__APP_BASE_HREF__"), []byte(baseHref))

		return data, nil
	})

	return fn.(func() ([]byte, error))()
}

func (o *opt) sendFile(f fs.FS, w http.ResponseWriter, r *http.Request, path string) {
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	data, err := o.loadOrProcess(f, path, o.resolveBaseHref(r.Header.Get(HeaderAppBaseHref)))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, bytes.NewBuffer(data)); err != nil {
	}
}

func (o *opt) staticFileHandler(f fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ext := path.Ext(r.URL.Path); ext != "" {
			w.Header().Set("Content-Type", mime.TypeByExtension(ext))
		}
		o.sendFile(f, w, r, r.URL.Path)
	})
}

func (o *opt) htmlHandler(f fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", mime.TypeByExtension(".html"))

		if !o.disableCSP {
			w.Header().Set("X-Frame-Options", "sameorigin")
		}

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		requestPath := "index.html"

		if !o.disableHistoryFallback {
			o.sendFile(f, w, r, requestPath)
			return
		}

		requestPath = strings.Trim(r.URL.Path, "/")

		for _, suffix := range []string{".html", "/index.html"} {
			requestPath := requestPath + suffix

			if requestPath[0] == '/' {
				requestPath = requestPath[1:]
			}

			if _, err := fs.Stat(f, requestPath); err == nil {
				o.sendFile(f, w, r, requestPath)
				return
			}
		}

		writeErr(w, http.StatusNotFound, fmt.Errorf("`%s` not exists", requestPath))
	})
}

func (o *opt) resolveBaseHref(headerAppBaseHref string) string {
	baseHref := o.baseHref
	p := path.Clean(cmp.Or(headerAppBaseHref) + baseHref)
	if p == "/" {
		return "/"
	}
	return p + "/"
}

func (o *opt) build(optFns ...OptFunc) *opt {
	for i := range optFns {
		optFns[i](o)
	}
	return o
}

const (
	HeaderAppBaseHref = "X-App-Base-Href"
)

func ServeFS(f fs.FS, optFns ...OptFunc) http.Handler {
	o := (&opt{baseHref: "/"}).build(optFns...)

	html := o.htmlHandler(f)
	static := o.staticFileHandler(f)

	return handler.ApplyMiddlewares(
		compress.HandlerLevel(gzip.DefaultCompression),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		baseHref := o.resolveBaseHref(r.Header.Get(HeaderAppBaseHref))

		if baseHref != "/" {
			if !strings.HasPrefix(r.URL.Path+"/", o.baseHref) {
				http.Redirect(w, r, path.Clean(o.baseHref+r.URL.Path), http.StatusFound)
				return
			} else {
				r.URL.Path = strings.TrimPrefix(r.URL.Path, o.baseHref[0:len(o.baseHref)-1])
			}
		}

		upath := r.URL.Path
		if !strings.HasPrefix(upath, "/") {
			upath = "/" + upath
			r.URL.Path = upath
		}

		requestPath := path.Clean(upath)

		if requestPath == "/" {
			html.ServeHTTP(w, r)
			return
		}

		if ext := path.Ext(requestPath); ext != "" && ext != ".html" {
			switch requestPath {
			case "/favicon.ico":
				expires(w.Header(), 24*time.Hour)
			case "/sw.js":
			default:
				if ext != ".json" {
					expires(w.Header(), 30*24*time.Hour)
				}
			}
			static.ServeHTTP(w, r)
			return
		}
		html.ServeHTTP(w, r)
	}))
}

func expires(header http.Header, d time.Duration) {
	header.Set("Cache-Control", fmt.Sprintf("max-age=%d", d/time.Second))
}

func writeErr(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	_, _ = w.Write([]byte(err.Error()))
}
