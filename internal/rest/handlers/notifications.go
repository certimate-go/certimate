package handlers

import (
	"context"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/internal/rest/resp"
)

type notifyService interface {
	TestPush(ctx context.Context, req *dtos.NotifyTestPushReq) (*dtos.NotifyTestPushResp, error)
}

type NotificationsHandler struct {
	service notifyService
}

func NewNotificationsHandler(router *router.RouterGroup[*core.RequestEvent], service notifyService) {
	handler := &NotificationsHandler{
		service: service,
	}

	group := router.Group("/notifications")
	group.POST("/test", handler.test)
}

func (handler *NotificationsHandler) test(e *core.RequestEvent) error {
	req := &dtos.NotifyTestPushReq{}
	if err := e.BindBody(req); err != nil {
		return resp.Err(e, err)
	}

	res, err := handler.service.TestPush(e.Request.Context(), req)
	if err != nil {
		return resp.Err(e, err)
	}

	return resp.Ok(e, res)
}
