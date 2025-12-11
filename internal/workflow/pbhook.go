package workflow

import (
	"context"
	"fmt"

	"github.com/pocketbase/pocketbase/core"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/repository"
)

func Register() {
	pb := app.GetApp()
	pb.OnRecordCreateRequest(domain.CollectionNameWorkflow).BindFunc(func(e *core.RecordRequestEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		if err := onWorkflowRecordCreateOrUpdate(e.Request.Context(), e.Record); err != nil {
			app.GetLogger().Error(err.Error())
			return err
		}

		return nil
	})
	pb.OnRecordUpdateRequest(domain.CollectionNameWorkflow).BindFunc(func(e *core.RecordRequestEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		if err := onWorkflowRecordCreateOrUpdate(e.Request.Context(), e.Record); err != nil {
			app.GetLogger().Error(err.Error())
			return err
		}

		return nil
	})
	pb.OnRecordDeleteRequest(domain.CollectionNameWorkflow).BindFunc(func(e *core.RecordRequestEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		if err := onWorkflowRecordDelete(e.Request.Context(), e.Record); err != nil {
			app.GetLogger().Error(err.Error())
			return err
		}

		return nil
	})
}

func onWorkflowRecordCreateOrUpdate(_ context.Context, record *core.Record) error {
	scheduler := app.GetScheduler()

	// 向数据库插入/更新时，同时更新定时任务
	enabled := record.GetBool("enabled")
	trigger := record.GetString("trigger")
	triggerCron := record.GetString("triggerCron")

	// 如果非定时触发或未启用，移除定时任务
	if !enabled || trigger != string(domain.WorkflowTriggerTypeScheduled) {
		scheduler.Remove(fmt.Sprintf("workflow#%s", record.Id))
		return nil
	}

	// 反之，重新添加定时任务
	workflowSrv := NewWorkflowService(repository.NewWorkflowRepository(), repository.NewWorkflowRunRepository(), repository.NewSettingsRepository())
	if err := addWorkflowJob(workflowSrv, record.Id, triggerCron); err != nil {
		return err
	}

	return nil
}

func onWorkflowRecordDelete(_ context.Context, record *core.Record) error {
	scheduler := app.GetScheduler()

	// 从数据库删除时，同时移除定时任务
	jobId := fmt.Sprintf("workflow#%s", record.Id)
	scheduler.Remove(jobId)

	return nil
}
