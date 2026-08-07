package scheduler

import (
	"context"
)

type certificateService interface {
	InitSchedule(ctx context.Context) error
}

func initCertificateScheduler(service certificateService) error {
	return service.InitSchedule(context.Background())
}
