package handlers

import (
	"context"
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/pluginhost"
	"github.com/certimate-go/certimate/internal/rest/resp"
)

type marketService interface {
	ListMarket(ctx context.Context) ([]pluginhost.MarketEntry, error)
	Install(ctx context.Context, providerType string) (*pluginhost.ReloadResult, error)
	Delete(ctx context.Context, providerType string) (*pluginhost.ReloadResult, error)
	Update(ctx context.Context, providerType string) (*pluginhost.ReloadResult, error)
}

type PluginMarketHandler struct {
	service marketService
}

func NewPluginMarketHandler(rg *router.RouterGroup[*core.RequestEvent], service marketService) {
	h := &PluginMarketHandler{service: service}
	rg.GET("/plugin/market", h.listMarket)
	rg.POST("/plugin/market/install", h.install)
	rg.DELETE("/plugin/market/{providerType}", h.delete)
	rg.POST("/plugin/market/update/{providerType}", h.updatePlugin)
}

func (h *PluginMarketHandler) listMarket(e *core.RequestEvent) error {
	entries, err := h.service.ListMarket(e.Request.Context())
	if err != nil {
		return resp.Err(e, err)
	}
	return resp.Ok(e, entries)
}

type installRequest struct {
	ProviderType string `json:"providerType"`
}

func (h *PluginMarketHandler) install(e *core.RequestEvent) error {
	var req installRequest
	if err := json.NewDecoder(e.Request.Body).Decode(&req); err != nil {
		return resp.Err(e, err)
	}
	if req.ProviderType == "" {
		return resp.Err(e, pluginhost.ErrMissingProviderType)
	}
	result, err := h.service.Install(e.Request.Context(), req.ProviderType)
	if err != nil {
		return resp.Err(e, err)
	}
	return resp.Ok(e, result)
}

func (h *PluginMarketHandler) delete(e *core.RequestEvent) error {
	providerType := e.Request.PathValue("providerType")
	if providerType == "" {
		return resp.Err(e, pluginhost.ErrMissingProviderType)
	}
	result, err := h.service.Delete(e.Request.Context(), providerType)
	if err != nil {
		return resp.Err(e, err)
	}
	return resp.Ok(e, result)
}

type updateRequest struct {
	ProviderType string `json:"providerType"`
}

func (h *PluginMarketHandler) updatePlugin(e *core.RequestEvent) error {
	var req updateRequest
	if err := json.NewDecoder(e.Request.Body).Decode(&req); err != nil {
		return resp.Err(e, err)
	}
	providerType := req.ProviderType
	if providerType == "" {
		providerType = e.Request.PathValue("providerType")
	}
	if providerType == "" {
		return resp.Err(e, pluginhost.ErrMissingProviderType)
	}
	result, err := h.service.Update(e.Request.Context(), providerType)
	if err != nil {
		return resp.Err(e, err)
	}
	return resp.Ok(e, result)
}
