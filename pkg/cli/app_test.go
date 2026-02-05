package cli_test

import (
	"context"
	"testing"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"

	"github.com/innoai-tech/infra/pkg/cli"
)

type Do struct {
	cli.C `args:"INPUT"`
	Shared
}

type Shared struct {
	X
}

type X struct {
	Src    []string `arg:"INPUT"`
	Force  bool     `flag:",omitzero" alias:"f"`
	Output string   `flag:",omitzero" alias:"o"`
}

func (x *X) InjectContext(ctx context.Context) context.Context {
	return ctx
}

func TestApp(t *testing.T) {
	t.Run("GIVEN a cli app with 'do' command", func(t *testing.T) {
		a := cli.NewApp("app", "1.0.0")
		do := cli.AddTo(a, &Do{})

		t.Run("WHEN execute with combined flags and arguments", func(t *testing.T) {
			Must(t, func() error {
				return cli.Execute(context.Background(), a, []string{"do", "--force", "-o", "build", "src"})
			})

			Then(t, "arguments and flags should be mapped to the command struct",
				Expect(do.Src, Equal([]string{"src"})),
				Expect(do.Output, Equal("build")),
				Expect(do.Force, Be(cmp.True())),
			)
		})
	})
}
