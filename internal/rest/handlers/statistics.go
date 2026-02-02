package handlers

import (
	"context"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/rest/resp"
)

type statisticsService interface {
	Get(ctx context.Context) (*domain.Statistics, error)
}

type StatisticsHandler struct {
	service statisticsService
}

func NewStatisticsHandler(router *router.RouterGroup[*core.RequestEvent], service statisticsService) {
	handler := &StatisticsHandler{
		service: service,
	}

	router.GET("/statistics", handler.get)

	router.GET("/statistics/get", handler.get) // 兼容旧版
}

func (handler *StatisticsHandler) get(e *core.RequestEvent) error {
	res, err := handler.service.Get(e.Request.Context())
	if err != nil {
		return resp.Err(e, err)
	}

	return resp.Ok(e, res)
}
