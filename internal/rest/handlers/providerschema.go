package handlers

import (
	"context"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/providerschema"
	"github.com/certimate-go/certimate/internal/rest/resp"
)

type providerSchemaService interface {
	GetByProviderType(ctx context.Context, providerType string, categories ...providerschema.Category) (*providerschema.Envelope, error)
	List(ctx context.Context) ([]*providerschema.Envelope, error)
}

type ProviderSchemaHandler struct {
	service providerSchemaService
}

func NewProviderSchemaHandler(router *router.RouterGroup[*core.RequestEvent], service providerSchemaService) {
	handler := &ProviderSchemaHandler{
		service: service,
	}

	router.GET("/provider-schemas", handler.list)
	router.GET("/provider-schemas/{providerType}", handler.getByProviderType)
}

func (handler *ProviderSchemaHandler) getByProviderType(e *core.RequestEvent) error {
	providerType := e.Request.PathValue("providerType")

	category := providerschema.Category(e.Request.URL.Query().Get("category"))
	switch category {
	case "", providerschema.CategoryAccess, providerschema.CategoryDeploy, providerschema.CategoryApply, providerschema.CategoryNotify:
	default:
		return resp.Err(e, domain.NewError(400, "invalid schema category"))
	}
	res, err := handler.service.GetByProviderType(e.Request.Context(), providerType, category)
	if err != nil {
		return resp.Err(e, err)
	}

	return resp.Ok(e, res)
}

func (handler *ProviderSchemaHandler) list(e *core.RequestEvent) error {
	res, err := handler.service.List(e.Request.Context())
	if err != nil {
		return resp.Err(e, err)
	}

	return resp.Ok(e, res)
}
