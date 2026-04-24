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

	Then(t, "应用字符串使用 name@version 形式",
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

	Then(t, "应用和组件元数据可直接组合使用",
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

	Then(t, "生成的注入辅助函数可读写上下文",
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

func TestRuntimeDoc(t *testing.T) {
	t.Parallel()

	infoDoc, infoOK := (&Info{}).RuntimeDoc()
	appFieldDoc, appFieldOK := (&App{}).RuntimeDoc("Name")
	appVersionDoc, appVersionOK := (&App{}).RuntimeDoc("Version")
	appImageDoc, appImageOK := (&App{}).RuntimeDoc("ImageNamespace")
	appMissingDoc, appMissingOK := (&App{}).RuntimeDoc("missing")
	componentNameDoc, componentNameOK := (&Component{}).RuntimeDoc("Name")
	componentFieldDoc, componentFieldOK := (&Component{}).RuntimeDoc("Options")
	componentMissingDoc, componentMissingOK := (&Component{}).RuntimeDoc("missing")
	infoAppDoc, infoAppOK := (&Info{}).RuntimeDoc("App")
	infoNameDoc, infoNameOK := (&Info{}).RuntimeDoc("Name")
	infoDescDoc, infoDescOK := (&Info{}).RuntimeDoc("Desc")
	infoComponentDoc, infoComponentOK := (&Info{}).RuntimeDoc("Component")
	infoMissingDoc, infoMissingOK := (&Info{}).RuntimeDoc("missing")
	prefixed, prefixedOK := runtimeDoc(&Info{}, "prefix: ")
	_, helperMissingOK := runtimeDoc(struct{}{}, "prefix: ")

	Then(t, "生成的运行时文档可暴露类型与字段说明",
		Expect(infoOK, Equal(true)),
		Expect(infoDoc, Equal([]string{"provide app info"})),
		Expect(appFieldOK, Equal(true)),
		Expect(appFieldDoc, Equal([]string{})),
		Expect(appVersionOK, Equal(true)),
		Expect(appVersionDoc, Equal([]string{})),
		Expect(appImageOK, Equal(true)),
		Expect(appImageDoc, Equal([]string{})),
		Expect(appMissingOK, Equal(false)),
		Expect(appMissingDoc, Equal([]string(nil))),
		Expect(componentNameOK, Equal(true)),
		Expect(componentNameDoc, Equal([]string{})),
		Expect(componentFieldOK, Equal(true)),
		Expect(componentFieldDoc, Equal([]string{})),
		Expect(componentMissingOK, Equal(false)),
		Expect(componentMissingDoc, Equal([]string(nil))),
		Expect(infoAppOK, Equal(true)),
		Expect(infoAppDoc, Equal([]string{})),
		Expect(infoNameOK, Equal(true)),
		Expect(infoNameDoc, Equal([]string{})),
		Expect(infoDescOK, Equal(true)),
		Expect(infoDescDoc, Equal([]string{})),
		Expect(infoComponentOK, Equal(true)),
		Expect(infoComponentDoc, Equal([]string{})),
		Expect(infoMissingOK, Equal(false)),
		Expect(infoMissingDoc, Equal([]string(nil))),
		Expect(prefixedOK, Equal(true)),
		Expect(prefixed, Equal([]string{"prefix: provide app info"})),
		Expect(helperMissingOK, Equal(false)),
	)
}
