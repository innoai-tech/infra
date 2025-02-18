package appinfo

import (
	"fmt"
	"net/url"
)

// Info provide app info
// +gengo:injectable:provider
type Info struct {
	App       *App
	Name      string
	Desc      string
	Component *Component
}

type App struct {
	Name           string
	Version        string
	ImageNamespace string
}

func (a App) String() string {
	return fmt.Sprintf("%s@%s", a.Name, a.Version)
}

type Component struct {
	Name    string
	Options url.Values
}
