package handlers

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/internal/rest/resp"
	"github.com/certimate-go/certimate/internal/setup"
)

func NewSetupHandler(router *router.RouterGroup[*core.RequestEvent], service *setup.Service) {
	group := router.Group("/setup")
	group.BindFunc(func(e *core.RequestEvent) error {
		e.Response.Header().Set("Cache-Control", "no-store")
		return e.Next()
	})
	group.GET("/status", func(e *core.RequestEvent) error {
		status, err := service.Status()
		if err != nil {
			return resp.Err(e, err)
		}
		return resp.Ok(e, status)
	})
	group.POST("", func(e *core.RequestEvent) error {
		status, err := service.Status()
		if err != nil {
			return resp.Err(e, err)
		}
		if status.Initialized {
			return resp.Err(e, domain.NewError(409, "Administrator already configured. Please log in."))
		}
		invalid := domain.NewError(400, "Expected a valid JSON setup request of at most 4096 bytes.")
		contentType, _, err := mime.ParseMediaType(e.Request.Header.Get("Content-Type"))
		if err != nil || contentType != "application/json" {
			return resp.Err(e, invalid)
		}
		e.Request.Body = http.MaxBytesReader(e.Response, e.Request.Body, 4096)
		decoder := json.NewDecoder(e.Request.Body)
		decoder.DisallowUnknownFields()
		req := &dtos.SetupReq{}
		if err := decoder.Decode(req); err != nil {
			return resp.Err(e, invalid)
		}
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			return resp.Err(e, invalid)
		}
		if err := service.Create(e.Request.Context(), req); err != nil {
			return resp.Err(e, err)
		}
		return resp.Ok(e, nil)
	})
}
