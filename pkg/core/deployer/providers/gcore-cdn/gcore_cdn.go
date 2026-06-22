package gcorecdn

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/G-Core/gcorelabscdn-go/gcore"
	"github.com/G-Core/gcorelabscdn-go/gcore/provider"
	"github.com/G-Core/gcorelabscdn-go/resources"
	"github.com/G-Core/gcorelabscdn-go/sslcerts"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/gcore-cdn"
	xgcore "github.com/certimate-go/certimate/pkg/utils/third-party/gcore"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
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
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

type wSDKClients struct {
	Resources *resources.Service
	SSLCerts  *sslcerts.Service
}

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	clients, err := createSDKClients(config.ApiToken)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*DeployResult, error) {
	if d.config.ResourceId == 0 {
		return nil, fmt.Errorf("config `resourceId` is required")
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
		cloudCertId = d.config.CertificateId

		// 更新证书
		rplres, err := d.sdkCertmgr.Replace(ctx, strconv.FormatInt(cloudCertId, 10), certPEM, privkeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to replace certificate file: %w", err)
		} else {
			d.logger.Info("ssl certificate replaced", slog.Any("result", rplres))
		}
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
		ProxySSLCA:         lo.Ternary(getResourceResp.ProxySSLCA != 0, &getResourceResp.ProxySSLCA, nil),
		ProxySSLData:       lo.Ternary(getResourceResp.ProxySSLData != 0, &getResourceResp.ProxySSLData, nil),
		Options:            &gcore.Options{},
	}
	updateResourceResp, err := d.sdkClients.Resources.Update(ctx, d.config.ResourceId, updateResourceReq)
	d.logger.Debug("sdk request 'resources.Update'", slog.Int64("params.resourceId", d.config.ResourceId), slog.Any("request", updateResourceReq), slog.Any("response", updateResourceResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'resources.Update': %w", err)
	}

	return &DeployResult{}, nil
}

func createSDKClients(apiToken string) (*wSDKClients, error) {
	if apiToken == "" {
		return nil, fmt.Errorf("gcore: invalid api token")
	}

	requester := provider.NewClient(
		xgcore.BASE_URL,
		provider.WithSigner(xgcore.NewAuthRequestSigner(apiToken)),
	)
	resourcesSrv := resources.NewService(requester)
	sslCertsSrv := sslcerts.NewService(requester)
	return &wSDKClients{
		Resources: resourcesSrv,
		SSLCerts:  sslCertsSrv,
	}, nil
}
