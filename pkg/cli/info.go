package cli

import (
	"context"
	"fmt"

	contextx "github.com/octohelm/x/context"
)

type Info struct {
	App       *App
	Name      string
	Desc      string
	Component string
}

func (info Info) InjectContext(ctx context.Context) context.Context {
	return ContextWithInfo(ctx, &info)
}

type AppOptionFunc = func(*App)

func WithImageNamespace(imageNamespace string) AppOptionFunc {
	return func(a *App) {
		a.ImageNamespace = imageNamespace
	}
}

type App struct {
	Name    string
	Version string

	ImageNamespace string
}

func (a App) String() string {
	return fmt.Sprintf("%s@%s", a.Name, a.Version)
}

type infoCtx struct {
}

func InfoFromContext(ctx context.Context) *Info {
	if info, ok := ctx.Value(infoCtx{}).(*Info); ok {
		return info
	}
	return nil
}

func ContextWithInfo(ctx context.Context, c *Info) context.Context {
	return contextx.WithValue(ctx, infoCtx{}, c)
}
