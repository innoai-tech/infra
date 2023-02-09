package cli_test

import (
	"context"
	"testing"

	"github.com/innoai-tech/infra/pkg/cli"
	. "github.com/octohelm/x/testing"
)

type Do struct {
	cli.C `args:"INPUT"`

	X
}

type X struct {
	Src    []string `arg:""`
	Force  bool     `flag:",omitempty" alias:"f"`
	Output string   `flag:",omitempty" alias:"o"`
}

func TestApp(t *testing.T) {
	t.Run("Setup cli app", func(t *testing.T) {
		a := cli.NewApp("app", "1.0.0")
		do := cli.AddTo(a, &Do{})

		t.Run("When execute `do` with flags and args", func(t *testing.T) {
			err := cli.Execute(context.Background(), a, []string{"do", "--force", "-o", "build", "src"})
			Expect(t, err, Be[error](nil))

			t.Run("Flags should be parsed correct", func(t *testing.T) {
				Expect(t, do.Force, Be(true))
				Expect(t, do.Output, Be("build"))
				Expect(t, do.Src, Equal([]string{"src"}))
			})
		})
	})
}
