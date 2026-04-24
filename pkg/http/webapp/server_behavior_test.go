package webapp

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/innoai-tech/infra/pkg/http/basehref"
	. "github.com/octohelm/x/testing/v2"
)

func TestServerSetDefaults(t *testing.T) {
	t.Setenv("ENV", "prod")

	s := &Server{}
	s.SetDefaults()

	Then(t, "SetDefaults 会补齐基础默认值",
		Expect(s.BaseHref, Equal("/")),
		Expect(s.Addr, Equal(":80")),
		Expect(s.Env, Equal("prod")),
	)
}

func TestServerInitFromRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	Must(t, func() error {
		return os.WriteFile(filepath.Join(root, "index.html"), []byte("hello __ENV__ __VERSION__ __APP_BASE_HREF__ __APP_CONFIG__"), 0o644)
	})

	s := &Server{
		Root:     root,
		Addr:     ":8080",
		Env:      "dev",
		Ver:      "1.2.3",
		BaseHref: "/base",
		Config:   "feature=true",
	}

	Must(t, func() error {
		return s.Init(context.Background())
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/base/", nil)
	rr := httptest.NewRecorder()

	s.svc.Handler.ServeHTTP(rr, req)

	body := rr.Body.String()

	Then(t, "Init 会装配可用的 handler 并替换占位符",
		Expect(s.svc != nil, Equal(true)),
		Expect(rr.Code, Equal(http.StatusOK)),
		Expect(strings.Contains(body, "hello dev 1.2.3 /base/"), Equal(true)),
		Expect(strings.Contains(body, "feature=true"), Equal(true)),
	)
}

func TestServerInitMissingIndex(t *testing.T) {
	t.Parallel()

	s := &Server{
		Root: t.TempDir(),
	}

	err := s.Init(context.Background())

	Then(t, "缺少 index.html 时返回清晰错误",
		Expect(err == nil, Equal(false)),
		Expect(strings.Contains(err.Error(), "index.html not found"), Equal(true)),
	)
}

func TestServeWithoutInitIsNoop(t *testing.T) {
	t.Parallel()

	s := &Server{}

	Then(t, "未初始化时 Serve 直接返回",
		ExpectDo(func() error {
			return s.Serve(context.Background())
		}),
	)
}

func TestServerBindFSInitRuntimeDocAndShutdown(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	s := &Server{
		Addr: addr,
	}
	s.BindFS(makeTestFS(map[string]string{
		"index.html": "ok",
	}))

	Must(t, func() error {
		return s.Init(context.Background())
	})

	envDoc, envOK := s.RuntimeDoc("Env")
	var optFn OptFunc
	optDoc, optOK := (&optFn).RuntimeDoc()
	prefixedDoc, prefixedOK := runtimeDoc(s, "prefix: ", "Env")

	serveDone := make(chan error, 1)
	go func() {
		serveDone <- s.Serve(context.Background())
	}()

	Must(t, func() error {
		return waitForServer(addr)
	})

	resp, err := http.Get("http://" + addr + "/")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	Must(t, func() error {
		return s.Shutdown(context.Background())
	})

	serveErr := <-serveDone

	Then(t, "BindFS 与真实 Serve/Shutdown 回路可正常工作",
		Expect(string(body), Equal("ok")),
		Expect(errors.Is(serveErr, http.ErrServerClosed), Equal(true)),
		Expect(envOK, Equal(true)),
		Expect(envDoc, Equal([]string{"app env name"})),
		Expect(optOK, Equal(true)),
		Expect(optDoc, Equal([]string{})),
		Expect(prefixedOK, Equal(true)),
		Expect(prefixedDoc, Equal([]string{"prefix: app env name"})),
	)
}

func TestServeFSRootHTML(t *testing.T) {
	t.Parallel()

	h := ServeFS(
		makeTestFS(map[string]string{
			"index.html": "env=__ENV__ version=__VERSION__ base=__APP_BASE_HREF__ cfg=__APP_CONFIG__",
		}),
		WithAppEnv("test"),
		WithAppVersion("v1"),
		WithAppConfig(map[string]string{"feature": "true"}),
	)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	Then(t, "根路径返回 html 并写入安全头",
		Expect(rr.Code, Equal(http.StatusOK)),
		Expect(rr.Header().Get("Content-Type"), Equal("text/html; charset=utf-8")),
		Expect(rr.Header().Get("X-Frame-Options"), Equal("sameorigin")),
		Expect(rr.Header().Get("X-Content-Type-Options"), Equal("nosniff")),
		Expect(rr.Header().Get("X-XSS-Protection"), Equal("1; mode=block")),
		Expect(strings.Contains(rr.Body.String(), "env=test"), Equal(true)),
		Expect(strings.Contains(rr.Body.String(), "version=v1"), Equal(true)),
		Expect(strings.Contains(rr.Body.String(), "base=/"), Equal(true)),
	)
}

func TestServeFSDisableCSP(t *testing.T) {
	t.Parallel()

	h := ServeFS(
		makeTestFS(map[string]string{"index.html": "ok"}),
		DisableCSP(true),
	)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	Then(t, "关闭 CSP 后不再注入 X-Frame-Options",
		Expect(rr.Header().Get("X-Frame-Options"), Equal("")),
	)
}

func TestServeFSBaseHrefRedirectAndHeaderResolution(t *testing.T) {
	t.Parallel()

	h := ServeFS(
		makeTestFS(map[string]string{"index.html": "base=__APP_BASE_HREF__"}),
		WithBaseHref("/console"),
	)

	redirectReq := httptest.NewRequest(http.MethodGet, "http://example.com/assets/app.js", nil)
	redirectResp := httptest.NewRecorder()
	h.ServeHTTP(redirectResp, redirectReq)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/console/", nil)
	req.Header.Set(basehref.HeaderAppBaseHref, "/clusters/demo/")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	Then(t, "base href 会处理重定向并拼接代理前缀",
		Expect(redirectResp.Code, Equal(http.StatusFound)),
		Expect(redirectResp.Header().Get("Location"), Equal("/console/assets/app.js")),
		Expect(strings.Contains(rr.Body.String(), "base=/clusters/demo/console/"), Equal(true)),
	)
}

func TestServeFSStaticAssets(t *testing.T) {
	t.Parallel()

	h := ServeFS(makeTestFS(map[string]string{
		"index.html":    "ok",
		"assets/app.js": "console.log('ok')",
		"data.json":     "{}",
	}))

	jsReq := httptest.NewRequest(http.MethodGet, "http://example.com/assets/app.js", nil)
	jsResp := httptest.NewRecorder()
	h.ServeHTTP(jsResp, jsReq)

	jsonReq := httptest.NewRequest(http.MethodGet, "http://example.com/data.json", nil)
	jsonResp := httptest.NewRecorder()
	h.ServeHTTP(jsonResp, jsonReq)

	Then(t, "静态资源会按扩展名返回并设置缓存头",
		Expect(jsResp.Code, Equal(http.StatusOK)),
		Expect(strings.Contains(jsResp.Header().Get("Content-Type"), "javascript"), Equal(true)),
		Expect(jsResp.Header().Get("Cache-Control"), Equal("max-age=2592000")),
		Expect(jsonResp.Header().Get("Cache-Control"), Equal("")),
	)
}

func TestServeFSHistoryFallbackSwitch(t *testing.T) {
	t.Parallel()

	files := makeTestFS(map[string]string{
		"index.html":       "root",
		"docs.html":        "doc html",
		"guide/index.html": "guide index",
	})

	withFallback := ServeFS(files)
	fallbackReq := httptest.NewRequest(http.MethodGet, "http://example.com/unknown", nil)
	fallbackResp := httptest.NewRecorder()
	withFallback.ServeHTTP(fallbackResp, fallbackReq)

	withoutFallback := ServeFS(files, DisableHistoryFallback(true))

	docsReq := httptest.NewRequest(http.MethodGet, "http://example.com/docs", nil)
	docsResp := httptest.NewRecorder()
	withoutFallback.ServeHTTP(docsResp, docsReq)

	guideReq := httptest.NewRequest(http.MethodGet, "http://example.com/guide", nil)
	guideResp := httptest.NewRecorder()
	withoutFallback.ServeHTTP(guideResp, guideReq)

	missingReq := httptest.NewRequest(http.MethodGet, "http://example.com/missing", nil)
	missingResp := httptest.NewRecorder()
	withoutFallback.ServeHTTP(missingResp, missingReq)

	Then(t, "history fallback 可切换为显式 html 查找模式",
		Expect(fallbackResp.Code, Equal(http.StatusOK)),
		Expect(fallbackResp.Body.String(), Equal("root")),
		Expect(docsResp.Body.String(), Equal("doc html")),
		Expect(guideResp.Body.String(), Equal("guide index")),
		Expect(missingResp.Code, Equal(http.StatusNotFound)),
		Expect(strings.Contains(missingResp.Body.String(), "not exists"), Equal(true)),
	)
}

func TestServeFSRejectsNonGet(t *testing.T) {
	t.Parallel()

	h := ServeFS(makeTestFS(map[string]string{"index.html": "ok"}))

	req := httptest.NewRequest(http.MethodPost, "http://example.com/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	Then(t, "非 GET 请求直接返回 204",
		Expect(rr.Code, Equal(http.StatusNoContent)),
	)
}

func makeTestFS(files map[string]string) fstest.MapFS {
	m := fstest.MapFS{}
	for name, content := range files {
		m[name] = &fstest.MapFile{Data: []byte(content)}
	}
	return m
}

func waitForServer(addr string) error {
	for range 50 {
		resp, err := http.Get("http://" + addr + "/")
		if err == nil {
			resp.Body.Close()
			return nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	return errors.New("server not ready")
}
