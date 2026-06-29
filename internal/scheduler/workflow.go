package scheduler

import (
	"context"
)

type workflowService interface {
	InitSchedule(ctx context.Context) error
}

func initWorkflowScheduler(service workflowService) error {
	return service.InitSchedule(context.Background())
}
