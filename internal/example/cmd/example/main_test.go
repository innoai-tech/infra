package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	exampleroutes "github.com/innoai-tech/infra/internal/example/cmd/example/routes"
	"github.com/innoai-tech/infra/internal/example/cmd/example/ui"
	apiv0 "github.com/innoai-tech/infra/internal/example/pkg/apis/org/v0"
	"github.com/innoai-tech/infra/pkg/appinfo"
	"github.com/innoai-tech/infra/pkg/cli"
	. "github.com/octohelm/x/testing/v2"
)

func TestServeCommandRoundTrip(t *testing.T) {
	s := &Serve{}
	s.Server.Addr = "127.0.0.1:0"
	s.Server.ApplyRouter(exampleroutes.R)
	s.Server.ApplyGlobalHandlers(exampleGlobalUIHandler())

	ctx := appinfo.InfoInjectContext(context.Background(), &appinfo.Info{
		App: &appinfo.App{
			Name:    "example",
			Version: "1.0.0",
		},
	})

	Must(t, func() error {
		return s.Server.Init(ctx)
	})

	serveDone := make(chan error, 1)
	go func() {
		serveDone <- s.Server.Serve(ctx)
	}()

	baseURL := s.Server.Endpoint()

	orgsResp := mustGet(t, baseURL+"/api/example/v0/orgs")
	defer orgsResp.Body.Close()

	var list apiv0.DataList
	Must(t, func() error {
		return json.NewDecoder(orgsResp.Body).Decode(&list)
	})

	orgResp := mustGet(t, baseURL+"/api/example/v0/orgs/demo")
	defer orgResp.Body.Close()

	notFoundResp := mustGet(t, baseURL+"/api/example/v0/orgs/NotFound")
	defer notFoundResp.Body.Close()

	zipResp := mustGet(t, baseURL+"/api/example/v0/archive/zip")
	defer zipResp.Body.Close()

	zipReader := mustZipReader(t, zipResp.Body)

	uiResp := mustGet(t, baseURL+"/")
	defer uiResp.Body.Close()
	uiBody := string(MustValue(t, func() ([]byte, error) {
		return io.ReadAll(uiResp.Body)
	}))

	Must(t, func() error {
		return s.Server.Shutdown(context.Background())
	})

	serveErr := <-serveDone

	Then(t, "Serve 命令可完整承载 API、压缩包下载和 UI fallback",
		Expect(orgsResp.StatusCode, Equal(http.StatusOK)),
		Expect(list.Total, Equal(1)),
		Expect(list.Data[0].Name, Equal("demo")),
		Expect(orgResp.StatusCode, Equal(http.StatusOK)),
		Expect(notFoundResp.StatusCode, Equal(http.StatusNotFound)),
		Expect(zipResp.Header.Get("Content-Type"), Equal("application/zip")),
		Expect(zipResp.Header.Get("Content-Disposition"), Equal(`attachment; filename="x.zip"`)),
		Expect(len(zipReader.File), Equal(2)),
		Expect(zipReader.File[0].Name, Equal("readme.md")),
		Expect(uiResp.StatusCode, Equal(http.StatusOK)),
		Expect(strings.Contains(uiBody, `<title>KubePkg Agent</title>`), Equal(true)),
		Expect(strings.Contains(uiBody, `<div id="root"></div>`), Equal(true)),
		Expect(errors.Is(serveErr, http.ErrServerClosed), Equal(true)),
	)
}

func TestWebappCommandRoundTrip(t *testing.T) {
	w := &Webapp{}
	w.Server.Addr = mustFreeAddr(t)
	w.Server.Root = filepath.Join(".", "ui", "dist")
	w.Server.Ver = "test"
	w.Server.SetDefaults()

	Must(t, func() error {
		return w.Server.Init(context.Background())
	})

	serveDone := make(chan error, 1)
	go func() {
		serveDone <- w.Server.Serve(context.Background())
	}()

	waitForHTTPReady(t, "http://"+w.Server.Addr+"/")

	resp := mustGet(t, "http://"+w.Server.Addr+"/")
	defer resp.Body.Close()
	body := string(MustValue(t, func() ([]byte, error) {
		return io.ReadAll(resp.Body)
	}))

	Must(t, func() error {
		return w.Server.Shutdown(context.Background())
	})

	serveErr := <-serveDone

	Then(t, "Webapp 命令可独立承载打包产物",
		Expect(resp.StatusCode, Equal(http.StatusOK)),
		Expect(strings.Contains(body, `<title>KubePkg Agent</title>`), Equal(true)),
		Expect(strings.Contains(body, `<div id="root"></div>`), Equal(true)),
		Expect(errors.Is(serveErr, http.ErrServerClosed), Equal(true)),
	)
}

func TestAppDumpK8sEntry(t *testing.T) {
	wd := MustValue(t, os.Getwd)
	tmp := t.TempDir()

	Must(t, func() error {
		return os.Chdir(tmp)
	})
	defer func() {
		_ = os.Chdir(wd)
	}()

	Must(t, func() error {
		return cli.Execute(context.Background(), App, []string{"serve", "--dump-k8s"})
	})

	raw := string(MustValue(t, func() ([]byte, error) {
		return os.ReadFile(filepath.Join(tmp, "cuepkg", "component", "example", "server.cue"))
	}))

	Then(t, "CLI 入口已正确挂载 serve 命令和 dump-k8s 能力",
		Expect(strings.Contains(raw, `ghcr.io/octohelm/example`), Equal(true)),
		Expect(strings.Contains(raw, `"serve"`), Equal(true)),
	)
}

func mustGet(t *testing.T, url string) *http.Response {
	t.Helper()

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("http get %s: %v", url, err)
	}

	return resp
}

func mustZipReader(t *testing.T, r io.Reader) *zip.Reader {
	t.Helper()

	data := MustValue(t, func() ([]byte, error) {
		return io.ReadAll(r)
	})

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("zip reader: %v", err)
	}

	return reader
}

func exampleGlobalUIHandler() func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if strings.HasPrefix(req.URL.Path, "/api/") || strings.HasPrefix(req.URL.Path, "/.sys/") {
				handler.ServeHTTP(rw, req)
				return
			}
			ui.UI.ServeHTTP(rw, req)
		})
	}
}

func mustFreeAddr(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen free addr: %v", err)
	}
	defer ln.Close()

	return ln.Addr().String()
}

func waitForHTTPReady(t *testing.T, url string) {
	t.Helper()

	client := &http.Client{Timeout: 200 * time.Millisecond}

	for range 50 {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("server not ready: %s", url)
}

func TestMain(m *testing.M) {
	time.Local = time.UTC
	os.Exit(m.Run())
}
