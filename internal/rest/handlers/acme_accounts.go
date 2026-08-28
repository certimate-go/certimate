package handlers

import (
	"context"
	"strconv"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/internal/rest/resp"
)

type acmeAccountService interface {
	List(ctx context.Context, req *dtos.ACMEAccountListReq) (*dtos.ACMEAccountListResp, error)
	Import(ctx context.Context, req *dtos.ACMEAccountImportReq) (*dtos.ACMEAccountImportResp, error)
	Export(ctx context.Context, accountId string) (*dtos.ACMEAccountExportResp, error)
	Rotate(ctx context.Context, accountId string) (*dtos.ACMEAccountRotateResp, error)
}

type ACMEAccountsHandler struct {
	service acmeAccountService
}

func NewACMEAccountsHandler(router *router.RouterGroup[*core.RequestEvent], service acmeAccountService) {
	handler := &ACMEAccountsHandler{service: service}

	group := router.Group("/acme-accounts")
	group.GET("", handler.list)
	group.POST("/import", handler.importAccount)
	group.POST("/{accountId}/export", handler.exportAccount)
	group.POST("/{accountId}/rotate", handler.rotateAccount)
}

func (handler *ACMEAccountsHandler) list(e *core.RequestEvent) error {
	page, _ := strconv.Atoi(e.Request.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(e.Request.URL.Query().Get("perPage"))
	ca := e.Request.URL.Query().Get("ca")

	res, err := handler.service.List(e.Request.Context(), &dtos.ACMEAccountListReq{
		Page:    page,
		PerPage: perPage,
		CA:      ca,
	})
	if err != nil {
		return resp.Err(e, err)
	}
	return resp.Ok(e, res)
}

func (handler *ACMEAccountsHandler) importAccount(e *core.RequestEvent) error {
	req := &dtos.ACMEAccountImportReq{}
	if err := e.BindBody(req); err != nil {
		return resp.Err(e, err)
	}

	res, err := handler.service.Import(e.Request.Context(), req)
	if err != nil {
		return resp.Err(e, err)
	}
	return resp.Ok(e, res)
}

func (handler *ACMEAccountsHandler) exportAccount(e *core.RequestEvent) error {
	accountId := e.Request.PathValue("accountId")
	res, err := handler.service.Export(e.Request.Context(), accountId)
	if err != nil {
		return resp.Err(e, err)
	}
	return resp.Ok(e, res)
}

func (handler *ACMEAccountsHandler) rotateAccount(e *core.RequestEvent) error {
	accountId := e.Request.PathValue("accountId")
	res, err := handler.service.Rotate(e.Request.Context(), accountId)
	if err != nil {
		return resp.Err(e, err)
	}
	return resp.Ok(e, res)
}
