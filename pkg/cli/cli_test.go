package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"

	"github.com/innoai-tech/infra/pkg/cli/internal"
)

type runtimeDocValue struct{}

func (runtimeDocValue) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) == 0 {
		return []string{"root desc"}, true
	}
	if names[0] == "Name" {
		return []string{"name desc"}, true
	}
	return nil, false
}

type nestedFlags struct {
	Name string `flag:",omitzero"`
}

type testChild struct {
	C `name:"child" component:"server" envprefix:"CUSTOM_"`
	runtimeDocValue
	nestedFlags
}

type CollectArgs struct {
	Input string `arg:"INPUT"`
}

type CollectFlags struct {
	Mode  string `flag:",omitzero" alias:"m"`
	Force bool   `flag:",omitzero"`
}

type collectCommand struct {
	C
	runtimeDocValue
	CollectArgs
	CollectFlags
}

type executable struct {
	C
	parsed   []string
	executed bool
}

type ExecSingleton struct {
	inited bool
	ran    bool
	Value  string `flag:",omitzero"`
}

func (s *ExecSingleton) Init(ctx context.Context) error {
	s.inited = true
	return nil
}

func (s *ExecSingleton) Run(ctx context.Context) error {
	s.ran = true
	return nil
}

type execCommand struct {
	C `name:"run"`
	ExecSingleton
}

type ListConfigOptions struct {
	Mode   string `flag:",omitzero"`
	Secret string `flag:",omitzero,secret"`
}

func (o *ListConfigOptions) InjectContext(ctx context.Context) context.Context {
	return ctx
}

type listConfigCommand struct {
	C `name:"inspect"`
	ListConfigOptions
}

type DumpComponentServer struct {
	Addr string `flag:",omitzero,expose=http"`
}

func (s *DumpComponentServer) InjectContext(ctx context.Context) context.Context {
	return ctx
}

type dumpComponentCommand struct {
	C `name:"serve" component:"server,kind=StatefulSet"`
	DumpComponentServer
}

func (e *executable) ParseArgs(args []string) {
	e.parsed = args
}

func (e *executable) ExecuteContext(ctx context.Context) error {
	e.executed = true
	return nil
}

func TestNewAppAndAddTo(t *testing.T) {
	t.Parallel()

	app := NewApp("demo", "1.2.3", WithImageNamespace("ghcr.io/demo")).(*app)
	child := AddTo(app, &testChild{})

	app.ParseArgs([]string{"child"})

	Then(
		t, "创建应用并挂载子命令",
		Expect(app.a.ImageNamespace, Equal("ghcr.io/demo")),
		Expect(len(app.C.subcommands), Equal(1)),
		Expect(app.C.subcommands[0], Equal(Command(child))),
	)
}

func TestExecuteUsesParserAndExecutor(t *testing.T) {
	t.Parallel()

	var x executable

	Then(
		t, "Execute 会先解析参数再执行上下文",
		ExpectDo(func() error {
			return Execute(context.Background(), &x, []string{"a", "b"})
		}),
	)
}

func TestExecuteUsesParserAndExecutorResult(t *testing.T) {
	t.Parallel()

	var x executable

	Must(t, func() error {
		return Execute(context.Background(), &x, []string{"a", "b"})
	})

	Then(
		t, "参数和执行状态会被记录",
		Expect(x.parsed, Equal([]string{"a", "b"})),
		Expect(x.executed, Equal(true)),
	)
}

func TestParseArgsBuildsCommandMetadata(t *testing.T) {
	t.Parallel()

	app := NewApp("demo", "1.0.0").(*app)
	child := AddTo(app, &testChild{})

	app.ParseArgs([]string{"child"})

	Then(
		t, "ParseArgs 会构建 cobra 命令和 flag 元数据",
		Expect(app.root != nil, Equal(true)),
		Expect(child.info.Name, Equal("child")),
		Expect(child.info.Component.Name, Equal("server")),
		Expect(child.info.Desc, Equal("root desc")),
		Expect(app.root.Commands(), Be(cmp.Len[[]*cobra.Command](1))),
	)
}

func TestCollectFlagsFromConfigurator(t *testing.T) {
	t.Parallel()

	c := &C{}
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	cmd := &collectCommand{}

	collectFlagsFromConfigurator(c, flags, reflect.ValueOf(cmd).Elem(), "", "APP_", "")

	Then(
		t, "收集参数与 flags 元数据",
		Expect(c.args, Be(cmp.Len[internal.Args](1))),
		Expect(c.flagVars, Be(cmp.Len[[]*internal.FlagVar](2))),
		Expect(c.flagVars[0].Name, Equal("mode")),
		Expect(c.flagVars[0].Alias, Equal("m")),
		Expect(c.flagVars[0].EnvVar, Equal("APP_MODE")),
		Expect(flags.Lookup("mode") != nil, Equal(true)),
		Expect(flags.Lookup("force") != nil, Equal(true)),
	)
}

func TestExecuteCommandLifecycle(t *testing.T) {
	t.Setenv("DEMO_VALUE", "from-env")

	app := NewApp("demo", "1.0.0").(*app)
	cmd := AddTo(app, &execCommand{})

	Must(t, func() error {
		return Execute(context.Background(), app, []string{"run"})
	})

	Then(
		t, "执行命令时会加载环境变量并运行 singleton 生命周期",
		Expect(cmd.ExecSingleton.Value, Equal("from-env")),
		Expect(cmd.ExecSingleton.inited, Equal(true)),
		Expect(cmd.ExecSingleton.ran, Equal(true)),
	)
}

func TestExecuteCommandEnvOverridesFlagValue(t *testing.T) {
	t.Setenv("DEMO_VALUE", "from-env")

	app := NewApp("demo", "1.0.0").(*app)
	cmd := AddTo(app, &execCommand{})

	Must(t, func() error {
		return Execute(context.Background(), app, []string{"run", "--value", "from-flag"})
	})

	Then(
		t, "环境变量会覆盖命令行 flag 值",
		Expect(cmd.ExecSingleton.Value, Equal("from-env")),
	)
}

func TestListConfigurationOutput(t *testing.T) {
	t.Setenv("DEMO_MODE", "from-env")
	t.Setenv("DEMO_SECRET", "token")

	app := NewApp("demo", "1.0.0").(*app)
	_ = AddTo(app, &listConfigCommand{})

	output := captureStdout(t, func() {
		Must(t, func() error {
			return Execute(context.Background(), app, []string{"inspect", "--list-configuration"})
		})
	})

	lines := parseInfoLines(output)

	Then(
		t, "list-configuration 输出使用结构化 env=value 形式，secret 会被遮罩",
		Expect(lines["DEMO_MODE"], Equal("from-env")),
		Expect(lines["DEMO_SECRET"], Equal("-----")),
	)
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	origin := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}

	os.Stdout = w
	defer func() {
		os.Stdout = origin
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	var b bytes.Buffer
	if _, err := io.Copy(&b, r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	return b.String()
}

func parseInfoLines(output string) map[string]string {
	lines := map[string]string{}

	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		key, value, ok := strings.Cut(line, " = ")
		if !ok {
			continue
		}
		lines[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}

	return lines
}
