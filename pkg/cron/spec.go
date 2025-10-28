package cron

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/robfig/cron/v3"
)

type Spec string

func (spec Spec) Times(ctx context.Context) iter.Seq[time.Time] {
	s, _ := cron.ParseStandard(string(spec))
	if s == nil {
		return func(yield func(time.Time) bool) {
		}
	}
	return Times(ctx, s)
}

func (spec Spec) IsZero() bool {
	return spec == ""
}

func (spec Spec) Schedule() cron.Schedule {
	if spec == "@never" {
		return nil
	}
	s, _ := cron.ParseStandard(string(spec))
	return s
}

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
