package handlers

import (
	"context"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/rest/resp"
	"github.com/certimate-go/certimate/internal/system"
)

type environmentService interface {
	GetEnvironment(context.Context) (*system.Environment, error)
}

type SystemHandler struct {
	service environmentService
}

func NewSystemHandler(router *router.RouterGroup[*core.RequestEvent], service environmentService) {
	handler := &SystemHandler{service: service}

	group := router.Group("/system")
	group.GET("/environment", handler.getEnvironment)
}

func (handler *SystemHandler) getEnvironment(e *core.RequestEvent) error {
	env, err := handler.service.GetEnvironment(e.Request.Context())
	if err != nil {
		return resp.Err(e, err)
	}

	return resp.Ok(e, env)
}
