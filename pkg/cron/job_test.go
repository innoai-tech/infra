package cron

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestJob(t *testing.T) {
	job := &Job{}
	_ = job.Init(context.Background())
	job.schedule = IntervalSchedule{
		Interval: 50 * time.Millisecond,
	}

	t.Cleanup(func() {
		_ = job.Shutdown(context.Background())
	})

	v := int64(0)
	done := make(chan struct{})

	job.ApplyAction("test", func(ctx context.Context) {
		defer func() {
			if atomic.LoadInt64(&v) >= 5 {
				done <- struct{}{}
			}
		}()

		atomic.AddInt64(&v, 1)
		fmt.Println(atomic.LoadInt64(&v))
	})

	go func() {
		_ = job.Serve(context.Background())
	}()

	<-done
}
