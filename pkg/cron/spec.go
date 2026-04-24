package cron

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/robfig/cron/v3"
)

// Spec 表示一个标准 cron 表达式。
type Spec string

// Times 基于 spec 按触发时间持续产出 time.Time。
func (spec Spec) Times(ctx context.Context) iter.Seq[time.Time] {
	s, _ := cron.ParseStandard(string(spec))
	if s == nil {
		return func(yield func(time.Time) bool) {
		}
	}
	return Times(ctx, s)
}

// IsZero 返回 spec 是否为空。
func (spec Spec) IsZero() bool {
	return spec == ""
}

// Schedule 将 spec 解析为 cron 调度器。
//
// `@never` 会被视为显式禁用，因此返回 nil。
func (spec Spec) Schedule() cron.Schedule {
	if spec == "@never" {
		return nil
	}
	s, _ := cron.ParseStandard(string(spec))
	return s
}

// UnmarshalText 校验并加载文本形式的 cron 表达式。
func (spec *Spec) UnmarshalText(text []byte) error {
	s := string(text)

	switch s {
	case "@never":
		*spec = Spec(s)
		return nil
	default:
		_, err := cron.ParseStandard(s)
		if err != nil {
			return fmt.Errorf("invalid cron spec: %s: %w", s, err)
		}
		*spec = Spec(s)
	}

	return nil
}

// Times 基于给定 schedule 按触发时间持续产出 time.Time。
func Times(ctx context.Context, schedule cron.Schedule) iter.Seq[time.Time] {
	if schedule == nil {
		return func(yield func(time.Time) bool) {
		}
	}

	next := func() time.Duration {
		now := time.Now()
		return schedule.Next(now).Sub(now)
	}

	return func(yield func(time.Time) bool) {
		timer := time.NewTimer(next())
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-timer.C:
				if !ok || !yield(v) {
					return
				}
			}

			timer.Reset(next())
		}
	}
}
