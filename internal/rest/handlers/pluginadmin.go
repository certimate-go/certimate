package handlers

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/pluginhost"
	"github.com/certimate-go/certimate/internal/rest/resp"
)

type PluginAdminHandler struct{}

func NewPluginAdminHandler(rg *router.RouterGroup[*core.RequestEvent]) {
	h := &PluginAdminHandler{}
	rg.POST("/plugin/reload", h.reload)
}

func (h *PluginAdminHandler) reload(e *core.RequestEvent) error {
	reloader := pluginhost.GlobalReloader()
	if reloader == nil {
		return resp.Err(e, fmt.Errorf("plugin reloader not initialized"))
	}
	result := reloader.ReloadNow(e.Request.Context())
	return resp.Ok(e, result)
}
