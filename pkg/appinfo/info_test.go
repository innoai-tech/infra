package appinfo

import (
	"context"
	"net/url"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestAppString(t *testing.T) {
	t.Parallel()

	app := App{
		Name:    "infra",
		Version: "1.2.3",
	}

	Then(
		t, "应用字符串使用 name@version 形式",
		Expect(app.String(), Equal("infra@1.2.3")),
	)
}

func TestInfoCarriesMetadata(t *testing.T) {
	t.Parallel()

	info := Info{
		App: &App{
			Name:           "infra",
			Version:        "1.2.3",
			ImageNamespace: "ghcr.io/innoai-tech",
		},
		Name: "example",
		Desc: "example command",
		Component: &Component{
			Name: "server",
			Options: url.Values{
				"expose": []string{"http"},
			},
		},
	}

	Then(
		t, "应用和组件元数据可直接组合使用",
		Expect(info.App.String(), Equal("infra@1.2.3")),
		Expect(info.App.ImageNamespace, Equal("ghcr.io/innoai-tech")),
		Expect(info.Name, Equal("example")),
		Expect(info.Component.Name, Equal("server")),
		Expect(info.Component.Options.Get("expose"), Equal("http")),
	)
}

func TestInfoInjectableHelpers(t *testing.T) {
	t.Parallel()

	info := &Info{Name: "example"}
	ctx := InfoInjectContext(context.Background(), info)
	fromCtx, ok := InfoFromContext(ctx)
	injected, injectedOK := InfoFromContext(info.InjectContext(context.Background()))
	_, missing := InfoFromContext(context.Background())

	Then(
		t, "生成的注入辅助函数可读写上下文",
		Expect(ok, Equal(true)),
		Expect(missing, Equal(false)),
		Expect(fromCtx, Equal(info)),
		ExpectDo(func() error {
			return info.Init(context.Background())
		}),
		Expect(injectedOK, Equal(true)),
		Expect(injected, Equal(info)),
	)
}
