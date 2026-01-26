package handlers

import (
	"context"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/internal/rest/resp"
)

type certificateService interface {
	DownloadCertificate(ctx context.Context, req *dtos.CertificateDownloadReq) (*dtos.CertificateDownloadResp, error)
	RevokeCertificate(ctx context.Context, req *dtos.CertificateRevokeReq) (*dtos.CertificateRevokeResp, error)
}

type CertificatesHandler struct {
	service certificateService
}

func NewCertificatesHandler(router *router.RouterGroup[*core.RequestEvent], service certificateService) {
	handler := &CertificatesHandler{
		service: service,
	}

	group := router.Group("/certificates")
	group.POST("/{certificateId}/download", handler.downloadCertificate)
	group.POST("/{certificateId}/revoke", handler.revokeCertificate)

	group.POST("/{certificateId}/archive", handler.downloadCertificate) // 兼容旧版
}

func (handler *CertificatesHandler) downloadCertificate(e *core.RequestEvent) error {
	req := &dtos.CertificateDownloadReq{}
	req.CertificateId = e.Request.PathValue("certificateId")
	if err := e.BindBody(req); err != nil {
		return resp.Err(e, err)
	}

	res, err := handler.service.DownloadCertificate(e.Request.Context(), req)
	if err != nil {
		return resp.Err(e, err)
	}

	return resp.Ok(e, res)
}

func (handler *CertificatesHandler) revokeCertificate(e *core.RequestEvent) error {
	req := &dtos.CertificateRevokeReq{}
	req.CertificateId = e.Request.PathValue("certificateId")
	if err := e.BindBody(req); err != nil {
		return resp.Err(e, err)
	}

	res, err := handler.service.RevokeCertificate(e.Request.Context(), req)
	if err != nil {
		return resp.Err(e, err)
	}

	return resp.Ok(e, res)
}
