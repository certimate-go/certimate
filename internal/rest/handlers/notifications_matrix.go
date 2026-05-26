package handlers

import (
	"strings"

	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/internal/repository"
	"github.com/certimate-go/certimate/internal/rest/resp"
	"github.com/certimate-go/certimate/pkg/core/notifier/providers/matrix"
	"github.com/pocketbase/pocketbase/core"
)

func (handler *NotificationsHandler) verifyMatrix(e *core.RequestEvent) error {
	req := &dtos.MatrixVerifyConnectionReq{}
	if err := e.BindBody(req); err != nil {
		return resp.Err(e, err)
	}

	cfg, err := matrix.ConfigFromMap(req.Config)
	if err != nil {
		return resp.Err(e, err)
	}

	result, err := matrix.VerifyConnection(e.Request.Context(), cfg)
	if err != nil {
		return resp.Err(e, err)
	}

	steps := make([]dtos.MatrixVerifyStepDTO, len(result.Steps))
	for i, s := range result.Steps {
		steps[i] = dtos.MatrixVerifyStepDTO{
			Name:          s.Name,
			Ok:            s.Ok,
			Message:       s.Message,
			Detail:        s.Detail,
			Code:          s.Code,
			RetryAfterSec: s.RetryAfterSec,
		}
	}

	out := &dtos.MatrixVerifyConnectionResp{
		Ok:                 result.Ok,
		UserId:             result.UserId,
		SessionAccessToken: result.SessionAccessToken,
		SessionDeviceId:    result.SessionDeviceId,
		Steps:              steps,
	}

	if result.Ok && strings.TrimSpace(req.AccessId) != "" && result.SessionAccessToken != "" {
		accessRepo := repository.NewAccessRepository()
		if err := accessRepo.MergeMatrixSessionIntoAccess(
			e.Request.Context(),
			req.AccessId,
			result.SessionAccessToken,
			result.SessionDeviceId,
		); err != nil {
			return resp.Err(e, err)
		}
		out.SessionSaved = true
	}

	return resp.Ok(e, out)
}

func (handler *NotificationsHandler) sendTestMatrix(e *core.RequestEvent) error {
	req := &dtos.MatrixTestSendReq{}
	if err := e.BindBody(req); err != nil {
		return resp.Err(e, err)
	}

	cfg, err := matrix.ConfigFromMap(req.Config)
	if err != nil {
		return resp.Err(e, err)
	}

	if err := matrix.SendTestMessage(e.Request.Context(), cfg, req.Subject, req.Message); err != nil {
		return resp.Err(e, err)
	}

	out := &dtos.MatrixTestSendResp{Ok: true}
	if sess, ok := matrix.TakePendingSession(cfg); ok {
		out.SessionAccessToken = sess.AccessToken
		out.SessionDeviceId = sess.DeviceID
	}

	if strings.TrimSpace(req.AccessId) != "" && out.SessionAccessToken != "" {
		accessRepo := repository.NewAccessRepository()
		if err := accessRepo.MergeMatrixSessionIntoAccess(
			e.Request.Context(),
			req.AccessId,
			out.SessionAccessToken,
			out.SessionDeviceId,
		); err != nil {
			return resp.Err(e, err)
		}
		out.SessionSaved = true
	}

	return resp.Ok(e, out)
}
