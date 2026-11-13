package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/pluginhost"
	"github.com/certimate-go/certimate/internal/rest/resp"
)

type marketService interface {
	ListMarketPage(ctx context.Context, query pluginhost.MarketQuery) (*pluginhost.MarketPage, error)
	Install(ctx context.Context, providerType string, sources ...string) (*pluginhost.InstallJob, error)
	Delete(ctx context.Context, providerType string, expected ...string) (*pluginhost.ReloadResult, error)
	Update(ctx context.Context, providerType string, expected ...string) (*pluginhost.ReloadResult, error)
	ListSources(context.Context) ([]pluginhost.MarketSourceStatus, error)
	AddSource(context.Context, string) (pluginhost.MarketSourceStatus, error)
	RemoveSource(context.Context, string) error
	RefreshSource(context.Context, string) (pluginhost.MarketSourceStatus, error)
	JobStatus(providerType string) (pluginhost.JobStatus, bool)
}

type PluginMarketHandler struct {
	service marketService
}

func NewPluginMarketHandler(rg *router.RouterGroup[*core.RequestEvent], service marketService) {
	h := &PluginMarketHandler{service: service}
	rg.GET("/plugin/market", h.listMarket)
	rg.GET("/plugin/market/sources", h.listSources)
	rg.POST("/plugin/market/sources", h.addSource)
	rg.POST("/plugin/market/sources/{source}/refresh", h.refreshSource)
	rg.DELETE("/plugin/market/sources/{source}", h.removeSource)
	rg.POST("/plugin/market/install", h.install)
	rg.GET("/plugin/market/install/status/{providerType}", h.installStatus)
	rg.DELETE("/plugin/market/{providerType}", h.delete)
	rg.POST("/plugin/market/update/{providerType}", h.updatePlugin)
}

func (h *PluginMarketHandler) listMarket(e *core.RequestEvent) error {
	query := e.Request.URL.Query()
	page, perPage := 1, 30
	for key, dest := range map[string]*int{"page": &page, "perPage": &perPage} {
		if query.Has(key) {
			value, err := strconv.Atoi(query.Get(key))
			if err != nil || value < 1 {
				return resp.Err(e, domain.NewError(400, "invalid "+key))
			}
			*dest = value
		}
	}
	status := query.Get("status")
	switch status {
	case "", "all", "installed", "not_installed":
	default:
		return resp.Err(e, domain.NewError(400, "invalid status"))
	}
	capability := query.Get("capability")
	switch capability {
	case "", "all", "deploy", "dns", "notify":
	default:
		return resp.Err(e, domain.NewError(400, "invalid capability"))
	}
	entries, err := h.service.ListMarketPage(e.Request.Context(), pluginhost.MarketQuery{
		Page: page, PerPage: perPage, Keyword: query.Get("keyword"), Status: status,
		Source: query.Get("source"), Capability: capability, Locale: query.Get("locale"), Refresh: query.Get("refresh") == "true",
	})
	if err != nil {
		return resp.Err(e, err)
	}
	return resp.Ok(e, entries)
}

type installRequest struct {
	ProviderType string  `json:"providerType"`
	Source       *string `json:"source"`
}

func (h *PluginMarketHandler) install(e *core.RequestEvent) error {
	var req installRequest
	if err := json.NewDecoder(e.Request.Body).Decode(&req); err != nil {
		return resp.Err(e, err)
	}
	if req.ProviderType == "" {
		return resp.Err(e, pluginhost.ErrMissingProviderType)
	}
	job, err := h.service.Install(e.Request.Context(), req.ProviderType, optionalSource(req.Source)...)
	if err != nil {
		return marketError(e, err)
	}
	return resp.Ok(e, job.Status())
}

func (h *PluginMarketHandler) installStatus(e *core.RequestEvent) error {
	providerType := e.Request.PathValue("providerType")
	if providerType == "" {
		return resp.Err(e, pluginhost.ErrMissingProviderType)
	}
	status, ok := h.service.JobStatus(providerType)
	if !ok {
		return resp.Err(e, domain.NewError(404, "market: no install job for "+providerType))
	}
	return resp.Ok(e, status)
}

func (h *PluginMarketHandler) delete(e *core.RequestEvent) error {
	providerType := e.Request.PathValue("providerType")
	if providerType == "" {
		return resp.Err(e, pluginhost.ErrMissingProviderType)
	}
	var expected []string
	if query := e.Request.URL.Query(); query.Has("source") {
		expected = []string{query.Get("source")}
	}
	result, err := h.service.Delete(e.Request.Context(), providerType, expected...)
	if err != nil {
		var dependencyError *pluginhost.AccessOwnerInUseError
		if errors.As(err, &dependencyError) {
			return e.JSON(http.StatusOK, resp.Response{Code: http.StatusConflict, Msg: dependencyError.Error(), Data: struct {
				Reason string `json:"reason"`
				*pluginhost.AccessOwnerInUseError
			}{Reason: "access_owner_in_use", AccessOwnerInUseError: dependencyError}})
		}
		return marketError(e, err)
	}
	return resp.Ok(e, result)
}

func (h *PluginMarketHandler) updatePlugin(e *core.RequestEvent) error {
	var req installRequest
	if err := json.NewDecoder(e.Request.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		return resp.Err(e, err)
	}
	providerType := e.Request.PathValue("providerType")
	if req.ProviderType != "" && req.ProviderType != providerType {
		return resp.Err(e, domain.ErrInvalidParams)
	}
	if providerType == "" {
		return resp.Err(e, pluginhost.ErrMissingProviderType)
	}
	result, err := h.service.Update(e.Request.Context(), providerType, optionalSource(req.Source)...)
	if err != nil {
		return marketError(e, err)
	}
	return resp.Ok(e, result)
}

func optionalSource(source *string) []string {
	if source == nil {
		return nil
	}
	return []string{*source}
}

func marketError(e *core.RequestEvent, err error) error {
	var reason string
	switch err {
	case pluginhost.ErrInstallInProgress:
		reason = "install_in_progress"
	case pluginhost.ErrOwnershipConflict:
		reason = "ownership_conflict"
	case pluginhost.ErrOpInProgress:
		reason = "operation_in_progress"
	case pluginhost.ErrMarketBusy:
		reason = "source_busy"
	case pluginhost.ErrDuplicateMarket:
		reason = "duplicate_source"
	default:
		return resp.Err(e, err)
	}
	return e.JSON(http.StatusOK, resp.Response{Code: http.StatusConflict, Msg: err.Error(), Data: map[string]string{"reason": reason}})
}

func (h *PluginMarketHandler) listSources(e *core.RequestEvent) error {
	sources, err := h.service.ListSources(e.Request.Context())
	if err != nil {
		return resp.Err(e, err)
	}
	return resp.Ok(e, sources)
}

func (h *PluginMarketHandler) addSource(e *core.RequestEvent) error {
	var req struct {
		Repository string `json:"repository"`
	}
	if err := e.BindBody(&req); err != nil {
		return resp.Err(e, domain.ErrInvalidParams)
	}
	source, err := h.service.AddSource(e.Request.Context(), req.Repository)
	if err != nil {
		return marketError(e, err)
	}
	return resp.Ok(e, source)
}

func (h *PluginMarketHandler) refreshSource(e *core.RequestEvent) error {
	source, err := h.service.RefreshSource(e.Request.Context(), e.Request.PathValue("source"))
	if err != nil {
		return marketError(e, err)
	}
	return resp.Ok(e, source)
}

func (h *PluginMarketHandler) removeSource(e *core.RequestEvent) error {
	if err := h.service.RemoveSource(e.Request.Context(), e.Request.PathValue("source")); err != nil {
		return marketError(e, err)
	}
	return resp.Ok(e, nil)
}
