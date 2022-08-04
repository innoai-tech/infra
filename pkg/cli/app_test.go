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
	Force  bool   `flag:",omitempty"`
	Output string `flag:",omitempty"`
}

func TestApp(t *testing.T) {
	t.Run("Setup cli app", func(t *testing.T) {
		a := cli.NewApp("app", "1.0.0")

		do := cli.AddTo(a, &Do{})

		t.Run("When execute `do` with flags and args", func(t *testing.T) {
			err := cli.Execute(context.Background(), a, []string{"do", "--force", "--output", "build", "src"})
			Expect(t, err, Be[error](nil))

			t.Run("Flags should be parsed correct", func(t *testing.T) {
				Expect(t, do.Force, Be(true))
				Expect(t, do.Output, Be("build"))
			})

			t.Run("Args should be set", func(t *testing.T) {
				Expect(t, do.Args, Equal([]string{"src"}))
			})
		})
	})
}
