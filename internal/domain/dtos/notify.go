package dtos

import (
	"github.com/certimate-go/certimate/internal/domain"
)

type NotifyTestPushReq struct {
	Provider domain.NotificationProviderType `json:"provider"`
	AccessId string                          `json:"accessId"`
}

type NotifyTestPushResp struct{}
