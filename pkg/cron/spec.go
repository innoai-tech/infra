package cron

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"iter"
	"time"
)

type Spec string

func (spec Spec) TimeSeq() iter.Seq[time.Time] {
	s, _ := cron.ParseStandard(string(spec))
	return TimeSeq(s)
}

func (spec Spec) Schedule() cron.Schedule {
	s, _ := cron.ParseStandard(string(spec))
	return s
}

func (spec *Spec) UnmarshalText(text []byte) error {
	s := string(text)

	switch s {
	case "@never":
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

func TimeSeq(schedule cron.Schedule) iter.Seq[time.Time] {
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
			case v := <-timer.C:
				if !yield(v) {
					return
				}
			}

			timer.Reset(next())
		}
	}
}
