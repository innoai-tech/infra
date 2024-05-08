package cron

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-courier/logr"
	"github.com/innoai-tech/infra/pkg/configuration"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

type IntervalSchedule struct {
	Interval time.Duration
}

func (i IntervalSchedule) Next(t time.Time) time.Time {
	return t.Add(i.Interval)
}

type Job struct {
	Cron string `flag:",omitempty"`

	schedule cron.Schedule
	timer    *time.Timer

	name   string
	action func(ctx context.Context)
}

func (j *Job) SetDefaults() {
	if j.Cron == "" {
		// 每周一
		// "https://crontab.guru/#0_0_*_*_1"
		j.Cron = "0 0 * * 1"
	}
}

func (j *Job) ApplyAction(name string, action func(ctx context.Context)) {
	j.name = name
	j.action = action
}

func (j *Job) Init(ctx context.Context) error {
	schedule, err := cron.ParseStandard(j.Cron)
	if err != nil {
		return errors.Wrapf(err, "parse cron failed: %s", j.Cron)
	}
	j.schedule = schedule
	return nil
}

var _ configuration.Server = (*Job)(nil)

func (j *Job) Serve(ctx context.Context) error {
	ci := configuration.ContextInjectorFromContext(ctx)

	logr.FromContext(ctx).WithValues(
		slog.String("name", j.name),
		slog.String("cron", j.Cron),
	).Info("waiting")

	j.timer = time.NewTimer(5 * time.Second)

	for {
		now := time.Now()

		j.timer.Reset(j.schedule.Next(now).Sub(now))

		select {
		case <-ctx.Done():
			return ctx.Err()
		case now = <-j.timer.C:
			if j.action != nil {
				go func() {
					j.action(ci.InjectContext(context.Background()))
				}()
			}
		}
	}
}

func (j *Job) Shutdown(context.Context) error {
	if j.timer != nil {
		j.timer.Stop()
	}
	return nil
}
