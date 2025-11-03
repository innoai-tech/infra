package cli_test

import (
	"context"
	"testing"

	testingx "github.com/octohelm/x/testing"

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
	Src []string `arg:"INPUT"`

	Force  bool   `flag:",omitempty" alias:"f"`
	Output string `flag:",omitempty" alias:"o"`
}

func (x *X) InjectContext(ctx context.Context) context.Context {
	return ctx
}

func TestApp(t *testing.T) {
	t.Run("Setup cli app", func(t *testing.T) {
		a := cli.NewApp("app", "1.0.0")
		do := cli.AddTo(a, &Do{})

		t.Run("When execute `do` with flags and args", func(t *testing.T) {
			err := cli.Execute(context.Background(), a, []string{"do", "--force", "-o", "build", "src"})
			testingx.Expect(t, err, testingx.BeNil[error]())

			t.Run("Flags should be parsed correct", func(t *testing.T) {
				testingx.Expect(t, do.Src, testingx.Equal([]string{"src"}))
				testingx.Expect(t, do.Force, testingx.Be(true))
				testingx.Expect(t, do.Output, testingx.Be("build"))
			})
		})
	})
}
