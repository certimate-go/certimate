package handlers

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/pluginhost"
	"github.com/certimate-go/certimate/internal/rest/resp"
)

type pluginCatalogService interface {
	Entries() []pluginhost.CatalogEntry
}

type PluginCatalogHandler struct {
	service pluginCatalogService
}

func NewPluginCatalogHandler(router *router.RouterGroup[*core.RequestEvent], service pluginCatalogService) {
	handler := &PluginCatalogHandler{service: service}
	router.GET("/plugin-catalog", handler.list)
}

func (handler *PluginCatalogHandler) list(e *core.RequestEvent) error {
	return resp.Ok(e, handler.service.Entries())
}
