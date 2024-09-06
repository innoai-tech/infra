package cron

import (
	"context"
	"github.com/pkg/errors"
	"log/slog"
	"sync"
	"time"

	"github.com/go-courier/logr"
	"github.com/innoai-tech/infra/pkg/configuration"
	"github.com/robfig/cron/v3"
)

type IntervalSchedule struct {
	Interval time.Duration
}

func (i IntervalSchedule) Next(t time.Time) time.Time {
	return t.Add(i.Interval)
}

type Job struct {
	// cron job 配置
	// 支持 标准格式
	// 也支持 @every {duration} 等语义化格式
	Cron string `flag:",omitempty"`

	schedule cron.Schedule
	timer    *time.Timer
	done     chan struct{}
	once     sync.Once

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

func (j *Job) ApplySchedule(s cron.Schedule) {
	j.schedule = s
}

func (j *Job) ApplyAction(name string, action func(ctx context.Context)) {
	j.name = name
	j.action = action
}

func (j *Job) Init(ctx context.Context) error {
	if j.schedule == nil {
		schedule, err := cron.ParseStandard(j.Cron)
		if err != nil {
			return errors.Wrapf(err, "parse cron failed: %s", j.Cron)
		}
		j.schedule = schedule
	}
	return nil
}

var _ configuration.Server = (*Job)(nil)

func (j *Job) Serve(ctx context.Context) error {
	ci := configuration.ContextInjectorFromContext(ctx)

	l := logr.FromContext(ctx).WithValues(
		slog.String("name", j.name),
		slog.String("cron", j.Cron),
	)

	l.Info("waiting")

	j.timer = time.NewTimer(5 * time.Second)

	j.done = make(chan struct{})

	for {
		now := time.Now()

		j.timer.Reset(j.schedule.Next(now).Sub(now))

		select {
		case <-j.done:
			return nil
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

func (j *Job) Shutdown(ctx context.Context) error {
	j.once.Do(func() {
		close(j.done)

		if j.timer != nil {
			j.timer.Stop()
		}
	})
	return nil
}
