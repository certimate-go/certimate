package gcorecdn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/G-Core/gcorelabscdn-go/gcore"
	"github.com/G-Core/gcorelabscdn-go/gcore/provider"
	"github.com/G-Core/gcorelabscdn-go/resources"
	"github.com/G-Core/gcorelabscdn-go/sslcerts"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/gcore-cdn"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	gcoresdk "github.com/certimate-go/certimate/pkg/sdk3rd/gcore"
)

type DeployerConfig struct {
	// G-Core API Token。
	ApiToken string `json:"apiToken"`
	// CDN 资源 ID。
	ResourceId int64 `json:"resourceId"`
	// 证书 ID。
	// 选填。零值时表示新建证书；否则表示更新证书。
	CertificateId int64 `json:"certificateId,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClients *wSDKClients
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

type wSDKClients struct {
	Resources *resources.Service
	SSLCerts  *sslcerts.Service
}

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	clients, err := createSDKClients(config.ApiToken)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		ApiToken: config.ApiToken,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create certmgr: %w", err)
	}

	return &Deployer{
		config:     config,
		logger:     slog.Default(),
		sdkClients: clients,
		sdkCertmgr: pcertmgr,
	}, nil
}

func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}

	d.sdkCertmgr.SetLogger(logger)
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	if d.config.ResourceId == 0 {
		return nil, errors.New("config `resourceId` is required")
	}

	// 如果原证书 ID 为空，则创建证书；否则更新证书。
	var cloudCertId int64
	if d.config.CertificateId == 0 {
		// 上传证书
		upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to upload certificate file: %w", err)
		} else {
			d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
		}

		cloudCertId, _ = strconv.ParseInt(upres.CertId, 10, 64)
	} else {
		// 获取证书
		// REF: https://api.gcore.com/docs/cdn#tag/SSL-certificates/paths/~1cdn~1sslData~1%7Bssl_id%7D/get
		getCertificateDetailResp, err := d.sdkClients.SSLCerts.Get(ctx, d.config.CertificateId)
		d.logger.Debug("sdk request 'sslcerts.Get'", slog.Int64("sslId", d.config.CertificateId), slog.Any("response", getCertificateDetailResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'sslcerts.Get': %w", err)
		}

		// 更新证书
		// REF: https://api.gcore.com/docs/cdn#tag/SSL-certificates/paths/~1cdn~1sslData~1%7Bssl_id%7D/get
		changeCertificateReq := &sslcerts.UpdateRequest{
			Name:           getCertificateDetailResp.Name,
			Cert:           certPEM,
			PrivateKey:     privkeyPEM,
			ValidateRootCA: false,
		}
		changeCertificateResp, err := d.sdkClients.SSLCerts.Update(ctx, getCertificateDetailResp.ID, changeCertificateReq)
		d.logger.Debug("sdk request 'sslcerts.Update'", slog.Int64("sslId", getCertificateDetailResp.ID), slog.Any("request", changeCertificateReq), slog.Any("response", changeCertificateResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'sslcerts.Update': %w", err)
		}

		cloudCertId = changeCertificateResp.ID
	}

	// 获取 CDN 资源详情
	// REF: https://api.gcore.com/docs/cdn#tag/CDN-resources/paths/~1cdn~1resources~1%7Bresource_id%7D/get
	getResourceResp, err := d.sdkClients.Resources.Get(ctx, d.config.ResourceId)
	d.logger.Debug("sdk request 'resources.Get'", slog.Any("resourceId", d.config.ResourceId), slog.Any("response", getResourceResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'resources.Get': %w", err)
	}

	// 更新 CDN 资源详情
	// REF: https://api.gcore.com/docs/cdn#tag/CDN-resources/operation/change_cdn_resource
	updateResourceReq := &resources.UpdateRequest{
		Description:        getResourceResp.Description,
		Active:             getResourceResp.Active,
		OriginGroup:        int(getResourceResp.OriginGroup),
		OriginProtocol:     getResourceResp.OriginProtocol,
		SecondaryHostnames: getResourceResp.SecondaryHostnames,
		SSlEnabled:         true,
		SSLData:            int(cloudCertId),
		ProxySSLEnabled:    getResourceResp.ProxySSLEnabled,
		Options:            &gcore.Options{},
	}
	if getResourceResp.ProxySSLCA != 0 {
		updateResourceReq.ProxySSLCA = &getResourceResp.ProxySSLCA
	}
	if getResourceResp.ProxySSLData != 0 {
		updateResourceReq.ProxySSLData = &getResourceResp.ProxySSLData
	}
	updateResourceResp, err := d.sdkClients.Resources.Update(ctx, d.config.ResourceId, updateResourceReq)
	d.logger.Debug("sdk request 'resources.Update'", slog.Int64("resourceId", d.config.ResourceId), slog.Any("request", updateResourceReq), slog.Any("response", updateResourceResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'resources.Update': %w", err)
	}

	return &deployer.DeployResult{}, nil
}

func createSDKClients(apiToken string) (*wSDKClients, error) {
	if apiToken == "" {
		return nil, errors.New("gcore: invalid api token")
	}

	requester := provider.NewClient(
		gcoresdk.BASE_URL,
		provider.WithSigner(gcoresdk.NewAuthRequestSigner(apiToken)),
	)
	resourcesSrv := resources.NewService(requester)
	sslCertsSrv := sslcerts.NewService(requester)
	return &wSDKClients{
		Resources: resourcesSrv,
		SSLCerts:  sslCertsSrv,
	}, nil
}
