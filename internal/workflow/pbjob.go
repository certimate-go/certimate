package workflow

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pocketbase/pocketbase/tools/cron"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/domain/dtos"
)

func addWorkflowJob(workflowSrv *WorkflowService, workflowId string, triggerCron string) error {
	scheduler := app.GetScheduler()

	jobId := fmt.Sprintf("workflow#%s", workflowId)
	job, _ := lo.Find(scheduler.Jobs(), func(j *cron.Job) bool { return j.Id() == jobId })
	if job != nil && job.Expression() == triggerCron {
		return nil
	}

	err := scheduler.Add(jobId, triggerCron, func() {
		app.GetLogger().Info(fmt.Sprintf("workflow #%s is triggered ...", workflowId))

		_, err := workflowSrv.StartRun(context.Background(), &dtos.WorkflowStartRunReq{
			WorkflowId: workflowId,
			RunTrigger: domain.WorkflowTriggerTypeScheduled,
		})
		if err != nil {
			app.GetLogger().Warn(fmt.Sprintf("failed to start scheduled run for workflow #%s", workflowId), slog.Any("error", err))
		}
	})
	if err != nil {
		app.GetLogger().Error(fmt.Sprintf("failed to register cron job for workflow #%s", workflowId), slog.Any("error", err))
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	app.GetLogger().Info(fmt.Sprintf("registered cron job for workflow #%s", workflowId), slog.String("cron", triggerCron))
	return nil
}
