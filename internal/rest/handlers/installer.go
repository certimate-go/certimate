package handlers

import (
	"context"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/internal/rest/resp"
)

type installerService interface {
	Status(ctx context.Context) (*dtos.InstallerStatusResp, error)
	InitializeSuperuser(ctx context.Context, req *dtos.InstallerInitializeSuperuserReq) (*dtos.InstallerInitializeSuperuserResp, error)
}

type InstallerHandler struct {
	service installerService
}

func NewInstallerHandler(router *router.RouterGroup[*core.RequestEvent], service installerService) {
	handler := &InstallerHandler{
		service: service,
	}

	group := router.Group("/installer")
	group.BindFunc(func(e *core.RequestEvent) error {
		e.Response.Header().Set("Cache-Control", "no-store")
		return e.Next()
	})

	group.GET("/status", handler.status)
	group.POST("/superuser", handler.initializeSuperuser)
}

func (handler *InstallerHandler) status(e *core.RequestEvent) error {
	res, err := handler.service.Status(e.Request.Context())
	if err != nil {
		return resp.Err(e, err)
	}

	return resp.Ok(e, res)
}

func (handler *InstallerHandler) initializeSuperuser(e *core.RequestEvent) error {
	req := &dtos.InstallerInitializeSuperuserReq{}
	if err := e.BindBody(req); err != nil {
		return resp.Err(e, err)
	}

	res, err := handler.service.InitializeSuperuser(e.Request.Context(), req)
	if err != nil {
		return resp.Err(e, err)
	}

	return resp.Ok(e, res)
}
