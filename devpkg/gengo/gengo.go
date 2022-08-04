package gengo

import (
	"context"

	"github.com/octohelm/gengo/pkg/gengo"
)

type Gengo struct {
	Entrypoint []string `arg:""`
}

func (g *Gengo) Run(ctx context.Context) error {
	c, err := gengo.NewContext(&gengo.GeneratorArgs{
		Globals: map[string][]string{
			"gengo:runtimedoc": {},
		},
		Entrypoint:         g.Entrypoint,
		OutputFileBaseName: "zz_generated",
	})
	if err != nil {
		return err
	}
	return c.Execute(ctx, gengo.GetRegisteredGenerators()...)
}
