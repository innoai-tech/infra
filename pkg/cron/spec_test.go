package cron

import (
	"context"
	"testing"
	"time"

	robfigcron "github.com/robfig/cron/v3"

	. "github.com/octohelm/x/testing/v2"
)

type fixedSchedule struct {
	after time.Duration
}

func (s fixedSchedule) Next(t time.Time) time.Time {
	return t.Add(s.after)
}

func TestSpecBasics(t *testing.T) {
	t.Parallel()

	var never Spec
	Must(t, func() error {
		return never.UnmarshalText([]byte("@never"))
	})

	var valid Spec
	Must(t, func() error {
		return valid.UnmarshalText([]byte("*/5 * * * *"))
	})

	var invalid Spec
	err := invalid.UnmarshalText([]byte("invalid"))

	Then(t, "Spec 支持零值、@never 和标准 cron 表达式",
		Expect(Spec("").IsZero(), Equal(true)),
		Expect(never, Equal(Spec("@never"))),
		Expect(never.Schedule(), Equal(robfigcron.Schedule(nil))),
		Expect(valid.IsZero(), Equal(false)),
		Expect(valid.Schedule() == nil, Equal(false)),
		Expect(err == nil, Equal(false)),
	)
}

func TestSpecTimesInvalidSpec(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	count := 0
	for range Spec("invalid").Times(ctx) {
		count++
	}

	Then(t, "非法 spec 不会产生任何触发时间",
		Expect(count, Equal(0)),
	)
}

func TestTimesNilSchedule(t *testing.T) {
	t.Parallel()

	count := 0
	for range Times(context.Background(), nil) {
		count++
	}

	Then(t, "nil schedule 直接返回空序列",
		Expect(count, Equal(0)),
	)
}

func TestTimesYieldStopsSequence(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	count := 0
	for range Times(ctx, fixedSchedule{after: time.Millisecond}) {
		count++
		break
	}

	Then(t, "yield 返回 false 时停止继续调度",
		Expect(count, Equal(1)),
	)
}

func TestTimesStopsOnContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	count := 0
	for range Times(ctx, fixedSchedule{after: time.Millisecond}) {
		count++
		cancel()
	}

	Then(t, "上下文取消后停止继续调度",
		Expect(count, Equal(1)),
	)
}

func TestRuntimeDoc(t *testing.T) {
	t.Parallel()

	var spec Spec
	doc, ok := (&spec).RuntimeDoc()
	prefixed, prefixedOK := runtimeDoc(&spec, "prefix: ")
	_, missingOK := runtimeDoc(struct{}{}, "prefix: ")

	Then(t, "生成的运行时文档 helper 可工作",
		Expect(ok, Equal(true)),
		Expect(doc, Equal([]string{})),
		Expect(prefixedOK, Equal(true)),
		Expect(prefixed, Equal([]string{})),
		Expect(missingOK, Equal(false)),
	)
}
