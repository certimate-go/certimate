package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pocketbase/dbx"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/internal/workflow/dispatcher"
)

type workflowRepository interface {
	ListEnabledScheduled(ctx context.Context) ([]*domain.Workflow, error)
	GetById(ctx context.Context, id string) (*domain.Workflow, error)
	Save(ctx context.Context, workflow *domain.Workflow) (*domain.Workflow, error)
}

type workflowRunRepository interface {
	GetById(ctx context.Context, id string) (*domain.WorkflowRun, error)
	Save(ctx context.Context, workflowRun *domain.WorkflowRun) (*domain.WorkflowRun, error)
	DeleteWhere(ctx context.Context, exprs ...dbx.Expression) (int, error)
}

type settingsRepository interface {
	GetByName(ctx context.Context, name string) (*domain.Settings, error)
}

type WorkflowService struct {
	dispatcher *dispatcher.WorkflowDispatcher

	workflowRepo    workflowRepository
	workflowRunRepo workflowRunRepository
	settingsRepo    settingsRepository
}

func NewWorkflowService(workflowRepo workflowRepository, workflowRunRepo workflowRunRepository, settingsRepo settingsRepository) *WorkflowService {
	srv := &WorkflowService{
		dispatcher: dispatcher.GetSingletonDispatcher(),

		workflowRepo:    workflowRepo,
		workflowRunRepo: workflowRunRepo,
		settingsRepo:    settingsRepo,
	}
	return srv
}

func (s *WorkflowService) InitSchedule(ctx context.Context) error {
	// 每日清理工作流执行历史
	app.GetScheduler().MustAdd("workflowHistoryRunsCleanup", "0 0 * * *", func() {
		settings, err := s.settingsRepo.GetByName(ctx, "persistence")
		if err != nil {
			app.GetLogger().Error("failed to get persistence settings", "err", err)
			return
		}

		persistenceSettings, _ := settings.UnmarshalContentAsPersistence()
		if persistenceSettings != nil && persistenceSettings.WorkflowRunsMaxDaysRetention != 0 {
			ret, err := s.workflowRunRepo.DeleteWhere(
				context.Background(),
				dbx.NewExp(fmt.Sprintf("status!='%s'", string(domain.WorkflowRunStatusTypePending))),
				dbx.NewExp(fmt.Sprintf("status!='%s'", string(domain.WorkflowRunStatusTypeRunning))),
				dbx.NewExp(fmt.Sprintf("endedAt<DATETIME('now', '-%d days')", persistenceSettings.WorkflowRunsMaxDaysRetention)),
			)
			if err != nil {
				app.GetLogger().Error("failed to delete workflow history runs", "err", err)
			}

			if ret > 0 {
				app.GetLogger().Info(fmt.Sprintf("cleanup %d workflow history runs", ret))
			}
		}
	})

	// 工作流
	{
		workflows, err := s.workflowRepo.ListEnabledScheduled(ctx)
		if err != nil {
			return err
		}

		for _, workflow := range workflows {
			var errs []error

			err := app.GetScheduler().Add(fmt.Sprintf("workflow#%s", workflow.Id), workflow.TriggerCron, func() {
				s.StartRun(ctx, &dtos.WorkflowStartRunReq{
					WorkflowId: workflow.Id,
					RunTrigger: domain.WorkflowTriggerTypeScheduled,
				})
			})
			if err != nil {
				errs = append(errs, err)
			}

			if len(errs) > 0 {
				return errors.Join(errs...)
			}
		}
	}

	return nil
}

func (s *WorkflowService) StartRun(ctx context.Context, req *dtos.WorkflowStartRunReq) error {
	workflow, err := s.workflowRepo.GetById(ctx, req.WorkflowId)
	if err != nil {
		return err
	}

	if workflow.LastRunStatus == domain.WorkflowRunStatusTypePending || workflow.LastRunStatus == domain.WorkflowRunStatusTypeRunning {
		return errors.New("workflow is already pending or running")
	}

	run := &domain.WorkflowRun{
		WorkflowId: workflow.Id,
		Status:     domain.WorkflowRunStatusTypePending,
		Trigger:    req.RunTrigger,
		StartedAt:  time.Now(),
		Detail:     workflow.Content,
	}
	if resp, err := s.workflowRunRepo.Save(ctx, run); err != nil {
		return err
	} else {
		run = resp
	}

	s.dispatcher.Dispatch(&dispatcher.WorkflowWorkerData{
		WorkflowId:      run.WorkflowId,
		WorkflowContent: run.Detail,
		RunId:           run.Id,
	})

	return nil
}

func (s *WorkflowService) CancelRun(ctx context.Context, req *dtos.WorkflowCancelRunReq) error {
	workflow, err := s.workflowRepo.GetById(ctx, req.WorkflowId)
	if err != nil {
		return err
	}

	workflowRun, err := s.workflowRunRepo.GetById(ctx, req.RunId)
	if err != nil {
		return err
	} else if workflowRun.WorkflowId != workflow.Id {
		return errors.New("workflow run not found")
	} else if workflowRun.Status != domain.WorkflowRunStatusTypePending && workflowRun.Status != domain.WorkflowRunStatusTypeRunning {
		return errors.New("workflow run is not pending or running")
	}

	s.dispatcher.Cancel(workflowRun.Id)

	return nil
}

func (s *WorkflowService) Shutdown(ctx context.Context) {
	s.dispatcher.Shutdown()
}
