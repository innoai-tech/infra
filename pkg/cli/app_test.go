package cli_test

import (
	"context"
	"testing"

	"github.com/innoai-tech/infra/pkg/cli"
	. "github.com/octohelm/x/testing"
)

type DoFlags struct {
	Force  bool   `flag:"force" desc:"Do force"`
	Output string `flag:"output,o" desc:"Output dir"`
}

type Do struct {
	cli.Name `args:"INPUT" desc:"do something"`
	DoFlags
}

type VerboseFlags struct {
	V int `flag:"!verbose,v" desc:"verbose level"`
}

func (v *VerboseFlags) PreRun(ctx context.Context) context.Context {
	return ctx
}

func TestApp(t *testing.T) {
	t.Run("Setup cli app", func(t *testing.T) {
		vflags := &VerboseFlags{}

		a := cli.NewApp("app", "1.0.0", vflags)
		d := cli.Add(a, &Do{})

		t.Run("When execute `do` with flags and args", func(t *testing.T) {
			err := cli.Execute(context.Background(), a, []string{"do", "-v1", "--force", "-o", "build", "src"})
			Expect(t, err, Be[error](nil))

			t.Run("Flags should be parsed correct", func(t *testing.T) {
				Expect(t, d.Force, Be(true))
				Expect(t, d.Output, Be("build"))
			})

			t.Run("Args should be set", func(t *testing.T) {
				Expect(t, d.Args, Equal([]string{"src"}))
			})
		})
	})
}
