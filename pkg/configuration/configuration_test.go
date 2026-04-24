package configuration

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/octohelm/x/cmp"
	contextx "github.com/octohelm/x/context"
	. "github.com/octohelm/x/testing/v2"
)

type testContextKey string

type testInjector struct {
	key      testContextKey
	value    string
	disabled bool
}

func (i testInjector) InjectContext(ctx context.Context) context.Context {
	return contextx.WithValue(ctx, i.key, i.value)
}

func (i testInjector) Disabled(ctx context.Context) bool {
	return i.disabled
}

type testSingleton struct {
	defaulted bool
	inited    bool
	ran       bool
	shutdown  bool
	value     string
}

func (s *testSingleton) SetDefaults() {
	s.defaulted = true
	if s.value == "" {
		s.value = "default"
	}
}

func (s *testSingleton) Init(ctx context.Context) error {
	s.inited = true
	if current, ok := CurrentInstanceFromContext(ctx); !ok || current != s {
		return errors.New("missing current instance")
	}
	return nil
}

func (s *testSingleton) InjectContext(ctx context.Context) context.Context {
	return contextx.WithValue(ctx, testContextKey("singleton"), s.value)
}

func (s *testSingleton) Run(ctx context.Context) error {
	s.ran = true
	return nil
}

func (s *testSingleton) Shutdown(ctx context.Context) error {
	s.shutdown = true
	return nil
}

type timeoutShutdown struct{}

func (timeoutShutdown) ShutdownTimeout(ctx context.Context) time.Duration {
	return 10 * time.Millisecond
}

func (timeoutShutdown) Shutdown(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

type lifecycleContainer struct {
	Named testSingleton
	testSingleton
	Ignored string
}

type ExportedSingleton struct {
	testSingleton
}

type lifecycleContainerWithAnonymous struct {
	Named ExportedSingleton
	ExportedSingleton
}

type postServeRecorder struct {
	served   bool
	postRun  bool
	shutdown bool
}

func (r *postServeRecorder) Serve(ctx context.Context) error {
	r.served = true
	return nil
}

func (r *postServeRecorder) PostServeRun(ctx context.Context) error {
	r.postRun = true
	return nil
}

func (r *postServeRecorder) Shutdown(ctx context.Context) error {
	r.shutdown = true
	return nil
}

type disabledServerRecorder struct {
	served   bool
	postRun  bool
	shutdown bool
	disabled bool
}

func (r *disabledServerRecorder) Disabled(ctx context.Context) bool {
	return r.disabled
}

func (r *disabledServerRecorder) Serve(ctx context.Context) error {
	r.served = true
	return nil
}

func (r *disabledServerRecorder) PostServeRun(ctx context.Context) error {
	r.postRun = true
	return nil
}

func (r *disabledServerRecorder) Shutdown(ctx context.Context) error {
	r.shutdown = true
	return nil
}

type disabledCleanupRecorder struct {
	ran      bool
	shutdown bool
	disabled bool
}

func (r *disabledCleanupRecorder) Run(ctx context.Context) error {
	r.ran = true
	return nil
}

func (r *disabledCleanupRecorder) Disabled(ctx context.Context) bool {
	return r.disabled
}

func (r *disabledCleanupRecorder) Shutdown(ctx context.Context) error {
	r.shutdown = true
	return nil
}

var lifecycleErr = errors.New("boom")

type initErrConfigurator struct{}

func (initErrConfigurator) Init(ctx context.Context) error {
	return lifecycleErr
}

type runErrConfigurator struct{}

func (runErrConfigurator) Run(ctx context.Context) error {
	return lifecycleErr
}

type serveErrConfigurator struct{}

func (serveErrConfigurator) Serve(ctx context.Context) error {
	return lifecycleErr
}

func (serveErrConfigurator) Shutdown(ctx context.Context) error {
	return nil
}

type shutdownErrConfigurator struct{}

func (shutdownErrConfigurator) Shutdown(ctx context.Context) error {
	return lifecycleErr
}

func TestContextHelpers(t *testing.T) {
	t.Parallel()

	ctx := Background(ContextInjectorInjectContext(context.Background(), testInjector{
		key:   testContextKey("k"),
		value: "v",
	}))

	Then(t, "background 继承当前上下文注入器",
		Expect(mustStringValue(ctx.Value(testContextKey("k"))), Equal("v")),
	)

	ctx = InjectContext(context.Background(),
		testInjector{key: testContextKey("a"), value: "1"},
		testInjector{key: testContextKey("b"), value: "2"},
	)

	Then(t, "InjectContext 依次注入上下文",
		Expect(mustStringValue(ctx.Value(testContextKey("a"))), Equal("1")),
		Expect(mustStringValue(ctx.Value(testContextKey("b"))), Equal("2")),
	)
}

func TestComposeContextInjector(t *testing.T) {
	t.Parallel()

	ci := ComposeContextInjector(
		testInjector{key: testContextKey("a"), value: "1"},
		testInjector{key: testContextKey("b"), value: "2", disabled: true},
	)

	ctx := ci.InjectContext(context.Background())

	Then(t, "组合注入器会跳过被禁用的项",
		Expect(mustStringValue(ctx.Value(testContextKey("a"))), Equal("1")),
		Expect(ctx.Value(testContextKey("b")) == nil, Equal(true)),
	)
}

func TestContextInjectorFromContextFallback(t *testing.T) {
	t.Parallel()

	ctx := ContextInjectorFromContext(context.Background()).InjectContext(context.Background())

	Then(t, "缺省注入器不会修改上下文",
		Expect(ctx == nil, Equal(false)),
	)
}

func TestInjectContextFunc(t *testing.T) {
	t.Parallel()

	ctx := InjectContext(context.Background(), InjectContextFunc(func(ctx context.Context, input string) context.Context {
		return contextx.WithValue(ctx, testContextKey("fn"), input)
	}, "ok"))

	Then(t, "函数注入器会把输入写入上下文",
		Expect(mustStringValue(ctx.Value(testContextKey("fn"))), Equal("ok")),
	)
}

func TestCurrentInstance(t *testing.T) {
	t.Parallel()

	ctx := CurrentInstanceInjectContext(context.Background(), "x")
	v, ok := CurrentInstanceFromContext(ctx)

	Then(t, "当前实例可写入并读出",
		Expect(ok, Equal(true)),
		Expect(mustStringValue(v), Equal("x")),
	)
}

func TestSingletonsFromStruct(t *testing.T) {
	t.Parallel()

	singletons := SingletonsFromStruct(&lifecycleContainerWithAnonymous{})

	names := make([]string, 0, len(singletons))
	for _, s := range singletons {
		names = append(names, s.Name)
	}

	Then(t, "提取命名与匿名 singleton 字段",
		Expect(names, Be(cmp.Len[[]string](2))),
		Expect(names, Equal([]string{"Named", ""})),
	)
}

func TestSingletonsInitAndRunOrServe(t *testing.T) {
	t.Parallel()

	named := testSingleton{}
	anonymous := testSingleton{}
	list := Singletons{
		{Name: "Named", Configurator: &named},
		{Name: "", Configurator: &anonymous},
	}

	ctx, err := list.Init(context.Background())
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	Must(t, func() error {
		return list.RunOrServe(ctx)
	})

	Then(t, "初始化会设置默认值、注入上下文并执行 runner",
		Expect(named.defaulted, Equal(true)),
		Expect(named.inited, Equal(true)),
		Expect(named.ran, Equal(true)),
		Expect(anonymous.defaulted, Equal(true)),
		Expect(mustStringValue(ContextInjectorFromContext(ctx).InjectContext(context.Background()).Value(testContextKey("singleton"))), Equal("default")),
	)
}

func TestRunOrServeCleanupShutdown(t *testing.T) {
	t.Parallel()

	s := &testSingleton{}

	Must(t, func() error {
		return RunOrServe(context.Background(), s)
	})

	Then(t, "无 server 时会在收尾阶段执行 shutdown",
		Expect(s.ran, Equal(true)),
		Expect(s.shutdown, Equal(true)),
	)
}

func TestShutdownTimeout(t *testing.T) {
	t.Parallel()

	Then(t, "Shutdown 使用自定义超时并返回上下文错误",
		ExpectDo(func() error {
			return Shutdown(context.Background(), timeoutShutdown{})
		},
			ErrorIs(context.DeadlineExceeded),
		),
	)
}

func TestRunOrServeWithServer(t *testing.T) {
	t.Parallel()

	r := &postServeRecorder{}

	Must(t, func() error {
		return RunOrServe(context.Background(), r)
	})

	Then(t, "存在 server 时会执行 serve post-serve 和 shutdown",
		Expect(r.served, Equal(true)),
		Expect(r.postRun, Equal(true)),
		Expect(r.shutdown, Equal(true)),
	)
}

func TestRunOrServeSkipsDisabledServer(t *testing.T) {
	t.Parallel()

	r := &disabledServerRecorder{disabled: true}

	Must(t, func() error {
		return RunOrServe(context.Background(), r)
	})

	Then(t, "禁用的 server 不会进入 serve post-serve 或 shutdown",
		Expect(r.served, Equal(false)),
		Expect(r.postRun, Equal(false)),
		Expect(r.shutdown, Equal(false)),
	)
}

func TestRunOrServeSkipsDisabledCleanup(t *testing.T) {
	t.Parallel()

	r := &disabledCleanupRecorder{disabled: true}

	Must(t, func() error {
		return RunOrServe(context.Background(), r)
	})

	Then(t, "禁用的 cleanup 项会运行 runner 但跳过 shutdown",
		Expect(r.ran, Equal(true)),
		Expect(r.shutdown, Equal(false)),
	)
}

func TestSingletonsConfiguratorsAndRuntimeDoc(t *testing.T) {
	t.Parallel()

	list := Singletons{
		{Name: "A", Configurator: "a"},
		{Name: "B", Configurator: "b"},
	}

	configurators := make([]any, 0, 1)
	for configurator := range list.Configurators() {
		configurators = append(configurators, configurator)
		break
	}

	singletonDoc, singletonOK := (&Singleton{}).RuntimeDoc()
	nameDoc, nameOK := (&Singleton{}).RuntimeDoc("Name")
	configuratorDoc, configuratorOK := (&Singleton{}).RuntimeDoc("Configurator")
	_, singletonMissingOK := (&Singleton{}).RuntimeDoc("missing")
	listDoc, listOK := (&Singletons{}).RuntimeDoc()
	prefixedDoc, prefixedOK := runtimeDoc(&Singleton{}, "prefix: ")
	_, helperMissingOK := runtimeDoc(struct{}{}, "prefix: ")

	Then(t, "Configurators 支持提前停止，RuntimeDoc 暴露生成文档",
		Expect(configurators, Equal([]any{"a"})),
		Expect(singletonOK, Equal(true)),
		Expect(singletonDoc, Equal([]string{})),
		Expect(nameOK, Equal(true)),
		Expect(nameDoc, Equal([]string{})),
		Expect(configuratorOK, Equal(true)),
		Expect(configuratorDoc, Equal([]string{})),
		Expect(singletonMissingOK, Equal(false)),
		Expect(listOK, Equal(true)),
		Expect(listDoc, Equal([]string{})),
		Expect(prefixedOK, Equal(true)),
		Expect(prefixedDoc, Equal([]string{})),
		Expect(helperMissingOK, Equal(false)),
	)
}

func TestLifecycleErrorsIncludeStageAndType(t *testing.T) {
	t.Parallel()

	initErr := Init(context.Background(), initErrConfigurator{})
	runErr := RunOrServe(context.Background(), runErrConfigurator{})
	serveErr := RunOrServe(context.Background(), serveErrConfigurator{})
	shutdownErr := Shutdown(context.Background(), shutdownErrConfigurator{})

	Then(t, "生命周期错误会携带阶段和类型信息，同时保留原始错误",
		Expect(strings.Contains(initErr.Error(), "init configuration.initErrConfigurator"), Equal(true)),
		ExpectDo(func() error { return initErr }, ErrorIs(lifecycleErr)),
		Expect(strings.Contains(runErr.Error(), "run configuration.runErrConfigurator"), Equal(true)),
		ExpectDo(func() error { return runErr }, ErrorIs(lifecycleErr)),
		Expect(strings.Contains(serveErr.Error(), "serve configuration.serveErrConfigurator"), Equal(true)),
		ExpectDo(func() error { return serveErr }, ErrorIs(lifecycleErr)),
		Expect(strings.Contains(shutdownErr.Error(), "shutdown configuration.shutdownErrConfigurator"), Equal(true)),
		ExpectDo(func() error { return shutdownErr }, ErrorIs(lifecycleErr)),
	)
}

func mustStringValue(v any) string {
	s, _ := v.(string)
	return s
}
