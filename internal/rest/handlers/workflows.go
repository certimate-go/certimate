package handlers

import (
	"context"
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/internal/rest/resp"
)

type workflowService interface {
	GetStatistics(ctx context.Context) (*dtos.WorkflowStatisticsResp, error)
	StartRun(ctx context.Context, req *dtos.WorkflowStartRunReq) (*dtos.WorkflowStartRunResp, error)
	CancelRun(ctx context.Context, req *dtos.WorkflowCancelRunReq) (*dtos.WorkflowCancelRunResp, error)
	Shutdown(ctx context.Context)
}

type WorkflowsHandler struct {
	service workflowService
}

func NewWorkflowsHandler(router *router.RouterGroup[*core.RequestEvent], service workflowService) {
	handler := &WorkflowsHandler{
		service: service,
	}

	group := router.Group("/workflows")
	group.GET("/stats", handler.getStatistics)
	group.POST("/{workflowId}/runs", handler.startRun)
	group.POST("/{workflowId}/runs/{runId}/cancel", handler.cancelRun)
}

func (handler *WorkflowsHandler) getStatistics(e *core.RequestEvent) error {
	res, err := handler.service.GetStatistics(e.Request.Context())
	if err != nil {
		return resp.Err(e, err)
	}

	return resp.Ok(e, res)
}

func (handler *WorkflowsHandler) startRun(e *core.RequestEvent) error {
	req := &dtos.WorkflowStartRunReq{}
	req.WorkflowId = e.Request.PathValue("workflowId")
	if err := e.BindBody(req); err != nil {
		return resp.Err(e, err)
	}
	if req.RunTrigger != domain.WorkflowTriggerTypeManual {
		return resp.Err(e, fmt.Errorf("invalid parameters: the value of 'trigger' must be 'manual'"))
	}

	res, err := handler.service.StartRun(e.Request.Context(), req)
	if err != nil {
		return resp.Err(e, err)
	}

	return resp.Ok(e, res)
}

func (handler *WorkflowsHandler) cancelRun(e *core.RequestEvent) error {
	req := &dtos.WorkflowCancelRunReq{}
	req.WorkflowId = e.Request.PathValue("workflowId")
	req.RunId = e.Request.PathValue("runId")

	res, err := handler.service.CancelRun(e.Request.Context(), req)
	if err != nil {
		return resp.Err(e, err)
	}

	return resp.Ok(e, res)
}
