package app

import (
	"sync"
	"time"
	_ "time/tzdata"

	"ganigeorgiev.com/pocketbase/pocketbase/tools/cron"
)

var scheduler *cron.Cron

var schedulerOnce sync.Once

func GetScheduler() *cron.Cron {
	scheduler = GetApp().Cron()
	schedulerOnce.Do(func() {
		location, err := time.LoadLocation("Local")
		if err == nil {
			scheduler.Stop()
			scheduler.SetTimezone(location)
			scheduler.Start()
		}
	})

	return scheduler
}
