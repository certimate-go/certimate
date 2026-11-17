package handlers

import (
	"context"
	"errors"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/internal/rest/resp"
)

type workflowTemplateService interface {
	GetById(ctx context.Context, id string) (*domain.WorkflowTemplate, error)
	List(ctx context.Context, keyword string) ([]*domain.WorkflowTemplate, error)
	Create(ctx context.Context, req *dtos.WorkflowTemplateCreateReq) (*domain.WorkflowTemplate, error)
	Update(ctx context.Context, id string, req *dtos.WorkflowTemplateUpdateReq) (*domain.WorkflowTemplate, error)
	Delete(ctx context.Context, id string) error
	Instantiate(ctx context.Context, id string, req *dtos.WorkflowTemplateInstantiateReq) (*dtos.WorkflowTemplateInstantiateResp, error)
}

type WorkflowTemplatesHandler struct {
	service workflowTemplateService
}

func NewWorkflowTemplatesHandler(router *router.RouterGroup[*core.RequestEvent], service workflowTemplateService) {
	handler := &WorkflowTemplatesHandler{
		service: service,
	}

	group := router.Group("/workflow-templates")
	group.GET("", handler.list)
	group.GET("/{templateId}", handler.get)
	group.POST("", handler.create)
	group.PUT("/{templateId}", handler.update)
	group.DELETE("/{templateId}", handler.remove)
	group.POST("/{templateId}/instantiate", handler.instantiate)
}

func (handler *WorkflowTemplatesHandler) list(e *core.RequestEvent) error {
	res, err := handler.service.List(e.Request.Context(), e.Request.URL.Query().Get("keyword"))
	if err != nil {
		return resp.Err(e, err)
	}

	return resp.Ok(e, res)
}

func (handler *WorkflowTemplatesHandler) get(e *core.RequestEvent) error {
	res, err := handler.service.GetById(e.Request.Context(), e.Request.PathValue("templateId"))
	if err != nil {
		return resp.Err(e, err)
	}

	return resp.Ok(e, res)
}

func (handler *WorkflowTemplatesHandler) create(e *core.RequestEvent) error {
	req := &dtos.WorkflowTemplateCreateReq{}
	if err := e.BindBody(req); err != nil {
		return resp.Err(e, err)
	}

	res, err := handler.service.Create(e.Request.Context(), req)
	if err != nil {
		return resp.Err(e, err)
	}

	return resp.Ok(e, res)
}

func (handler *WorkflowTemplatesHandler) update(e *core.RequestEvent) error {
	req := &dtos.WorkflowTemplateUpdateReq{}
	if err := e.BindBody(req); err != nil {
		return resp.Err(e, err)
	}

	res, err := handler.service.Update(e.Request.Context(), e.Request.PathValue("templateId"), req)
	if err != nil {
		return resp.Err(e, err)
	}

	return resp.Ok(e, res)
}

func (handler *WorkflowTemplatesHandler) remove(e *core.RequestEvent) error {
	if err := handler.service.Delete(e.Request.Context(), e.Request.PathValue("templateId")); err != nil {
		return resp.Err(e, err)
	}

	return resp.Ok(e, nil)
}

func (handler *WorkflowTemplatesHandler) instantiate(e *core.RequestEvent) error {
	req := &dtos.WorkflowTemplateInstantiateReq{}
	if err := e.BindBody(req); err != nil {
		return resp.Err(e, err)
	}

	res, err := handler.service.Instantiate(e.Request.Context(), e.Request.PathValue("templateId"), req)
	if err != nil {
		paramErrs := &dtos.WorkflowTemplateParamErrors{}
		if errors.As(err, &paramErrs) {
			return resp.ErrWithData(e, err, map[string]any{"errors": paramErrs.Items})
		}

		return resp.Err(e, err)
	}

	return resp.Ok(e, res)
}
