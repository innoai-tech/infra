package cli_test

import (
	"context"
	"testing"

	"github.com/innoai-tech/infra/pkg/cli"
	. "github.com/smartystreets/goconvey/convey"
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
	Convey("Setup cli app", t, func() {
		vflags := &VerboseFlags{}

		a := cli.NewApp("app", "1.0.0", vflags)
		d := cli.Add(a, &Do{})

		Convey("When execute `do` with flags and args", func() {
			err := cli.Execute(context.Background(), a, []string{"do", "-v1", "--force", "-o", "build", "src"})
			So(err, ShouldBeNil)

			Convey("Flags should be parsed correct", func() {
				So(d.Force, ShouldEqual, true)
				So(d.Output, ShouldEqual, "build")
			})

			Convey("Args should be set", func() {
				So(d.Args, ShouldResemble, []string{"src"})
			})
		})
	})
}
