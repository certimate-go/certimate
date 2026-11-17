package handlers

import (
	"context"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/internal/rest/resp"
)

type accessCheckService interface {
	TestCapabilities(providerType string) *dtos.AccessTestCapabilitiesResp
	Test(ctx context.Context, userId string, req *dtos.AccessTestReq) (*dtos.AccessTestResp, error)
}

type AccessChecksHandler struct {
	service accessCheckService
}

func NewAccessChecksHandler(router *router.RouterGroup[*core.RequestEvent], service accessCheckService) {
	handler := &AccessChecksHandler{
		service: service,
	}

	group := router.Group("/accesses")
	group.GET("/test-capabilities", handler.testCapabilities)
	group.POST("/test", handler.test)
}

func (handler *AccessChecksHandler) testCapabilities(e *core.RequestEvent) error {
	providerType := e.Request.URL.Query().Get("provider")
	if providerType == "" {
		return resp.Err(e, domain.NewError(400, "provider is required"))
	}

	return resp.Ok(e, handler.service.TestCapabilities(providerType))
}

func (handler *AccessChecksHandler) test(e *core.RequestEvent) error {
	req := &dtos.AccessTestReq{}
	if err := e.BindBody(req); err != nil {
		return resp.Err(e, domain.NewError(400, "invalid request body"))
	}
	if err := validateAccessTestReq(req); err != nil {
		return resp.Err(e, err)
	}
	if e.Auth == nil {
		return resp.Err(e, domain.NewError(403, "authentication required"))
	}

	res, err := handler.service.Test(e.Request.Context(), e.Auth.Id, req)
	if err != nil {
		return resp.Err(e, err)
	}

	return resp.Ok(e, res)
}

func validateAccessTestReq(req *dtos.AccessTestReq) error {
	switch req.Source {
	case dtos.AccessTestSourceStored:
		if req.AccessId == "" {
			return domain.NewError(400, "accessId is required")
		}
		if req.Provider != "" || req.Config != nil {
			return domain.NewError(400, "only one source may be specified")
		}

	case dtos.AccessTestSourceDraft:
		if req.AccessId != "" {
			return domain.NewError(400, "only one source may be specified")
		}
		if req.Provider == "" {
			return domain.NewError(400, "provider is required")
		}
		if req.Config == nil || len(req.Config) == 0 {
			return domain.NewError(400, "config is required")
		}

	default:
		return domain.NewError(400, "invalid source")
	}

	return nil
}
