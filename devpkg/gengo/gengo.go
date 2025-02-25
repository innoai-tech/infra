package gengo

import (
	"context"

	"github.com/octohelm/gengo/pkg/gengo"
)

type Gengo struct {
	Entrypoint []string `arg:""`
	// generate for all packages
	All bool `flag:",omitzero" alias:"a"`
	// force generate without cache
	Force bool `flag:",omitzero"`
}

func (g *Gengo) Run(ctx context.Context) error {
	c, err := gengo.NewContext(&gengo.GeneratorArgs{
		Entrypoint:         g.Entrypoint,
		OutputFileBaseName: "zz_generated",
		All:                g.All,
		Force:              g.Force,
		Globals: map[string][]string{
			"gengo:runtimedoc": {},
		},
	})
	if err != nil {
		return err
	}
	return c.Execute(ctx, gengo.GetRegisteredGenerators()...)
}
