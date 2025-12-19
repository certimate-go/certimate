package uclouduclb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/samber/lo"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/ucloud-ulb"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	ucloudsdk "github.com/certimate-go/certimate/pkg/sdk3rd/ucloud/ulb"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type DeployerConfig struct {
	// 优刻得 API 私钥。
	PrivateKey string `json:"privateKey"`
	// 优刻得 API 公钥。
	PublicKey string `json:"publicKey"`
	// 优刻得项目 ID。
	ProjectId string `json:"projectId,omitempty"`
	// 优刻得地域。
	Region string `json:"region"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 负载均衡实例 ID。
	// 部署资源类型为 [RESOURCE_TYPE_LOADBALANCER]、[RESOURCE_TYPE_VSERVER] 时必填。
	LoadbalancerId string `json:"loadbalancerId,omitempty"`
	// 负载均衡 VServer ID。
	// 部署资源类型为 [RESOURCE_TYPE_VSERVER] 时必填。
	VServerId string `json:"vserverId,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *ucloudsdk.ULBClient
	sdkCertmgr certmgr.Provider

	sslId2PemMap   map[string]string
	sslId2PemMapMu sync.Mutex
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.PrivateKey, config.PublicKey, config.ProjectId, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		PrivateKey: config.PrivateKey,
		PublicKey:  config.PublicKey,
		ProjectId:  config.ProjectId,
		Region:     config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create certmgr: %w", err)
	}

	return &Deployer{
		config:     config,
		logger:     slog.Default(),
		sdkClient:  client,
		sdkCertmgr: pcertmgr,

		sslId2PemMap:   make(map[string]string),
		sslId2PemMapMu: sync.Mutex{},
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
	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))

		d.sslId2PemMapMu.Lock()
		d.sslId2PemMap[upres.CertId] = certPEM
		d.sslId2PemMapMu.Unlock()
	}

	// 根据部署资源类型决定部署方式
	switch d.config.ResourceType {
	case RESOURCE_TYPE_LOADBALANCER:
		if err := d.deployToLoadbalancer(ctx, upres.CertId); err != nil {
			return nil, err
		}

	case RESOURCE_TYPE_VSERVER:
		if err := d.deployToVServer(ctx, upres.CertId); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported resource type '%s'", d.config.ResourceType)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToLoadbalancer(ctx context.Context, cloudCertId string) error {
	if d.config.LoadbalancerId == "" {
		return errors.New("config `loadbalancerId` is required")
	}

	// 获取 CLB 下的 HTTPS VServer 列表
	// REF: https://docs.ucloud.cn/api/ulb-api/describe_vserver
	vserverIds := make([]string, 0)
	describeVServerOffset := 0
	describeVServerLimit := 100
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		describeVServerReq := d.sdkClient.NewDescribeVServerRequest()
		describeVServerReq.ULBId = ucloud.String(d.config.LoadbalancerId)
		describeVServerReq.Offset = ucloud.Int(describeVServerOffset)
		describeVServerReq.Limit = ucloud.Int(describeVServerLimit)
		describeVServerResp, err := d.sdkClient.DescribeVServer(describeVServerReq)
		d.logger.Debug("sdk request 'ulb.DescribeVServer'", slog.Any("request", describeVServerReq), slog.Any("response", describeVServerResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'ulb.DescribeVServer': %w", err)
		}

		for _, vserverItem := range describeVServerResp.DataSet {
			if vserverItem.Protocol == "HTTPS" {
				vserverIds = append(vserverIds, vserverItem.VServerId)
			}
		}

		if len(describeVServerResp.DataSet) < describeVServerLimit {
			break
		}

		describeVServerOffset += describeVServerLimit
	}

	// 遍历更新 VServer 证书
	if len(vserverIds) == 0 {
		d.logger.Info("no clb vservers to deploy")
	} else {
		d.logger.Info("found https vservers to deploy", slog.Any("vserverIds", vserverIds))
		var errs []error

		for _, vserverId := range vserverIds {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := d.updateVServerCertificate(ctx, d.config.LoadbalancerId, vserverId, cloudCertId); err != nil {
					errs = append(errs, err)
				}
			}
		}

		if len(errs) > 0 {
			return errors.Join(errs...)
		}
	}

	return nil
}

func (d *Deployer) deployToVServer(ctx context.Context, cloudCertId string) error {
	if d.config.LoadbalancerId == "" {
		return errors.New("config `loadbalancerId` is required")
	}
	if d.config.VServerId == "" {
		return errors.New("config `vserverId` is required")
	}

	if err := d.updateVServerCertificate(ctx, d.config.LoadbalancerId, d.config.VServerId, cloudCertId); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) updateVServerCertificate(ctx context.Context, cloudLoadbalancerId, cloudVServerId string, cloudCertId string) error {
	// 获取 CLB 下的 VServer 信息
	// REF: https://docs.ucloud.cn/api/ulb-api/describe_vserver
	describeVServerReq := d.sdkClient.NewDescribeVServerRequest()
	describeVServerReq.ULBId = ucloud.String(cloudLoadbalancerId)
	describeVServerReq.VServerId = ucloud.String(cloudVServerId)
	describeVServerReq.Limit = ucloud.Int(1)
	describeVServerResp, err := d.sdkClient.DescribeVServer(describeVServerReq)
	d.logger.Debug("sdk request 'ulb.DescribeVServer'", slog.Any("request", describeVServerReq), slog.Any("response", describeVServerResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'ulb.DescribeVServer': %w", err)
	} else if len(describeVServerResp.DataSet) == 0 {
		return fmt.Errorf("could not find vserver '%s'", cloudVServerId)
	}

	// 跳过已部署过的 VServer
	vserverInfo := describeVServerResp.DataSet[0]
	if lo.ContainsBy(vserverInfo.SSLSet, func(item ulb.ULBSSLSet) bool { return item.SSLId == cloudCertId }) {
		return nil
	}

	// 绑定 SSL 证书
	// REF: https://docs.ucloud.cn/api/ulb-api/bind_ssl
	bindSSLReq := d.sdkClient.NewBindSSLRequest()
	bindSSLReq.ULBId = ucloud.String(cloudLoadbalancerId)
	bindSSLReq.VServerId = ucloud.String(cloudVServerId)
	bindSSLReq.SSLId = ucloud.String(cloudCertId)
	bindSSLResp, err := d.sdkClient.BindSSL(bindSSLReq)
	d.logger.Debug("sdk request 'ulb.BindSSL'", slog.Any("request", bindSSLReq), slog.Any("response", bindSSLResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'ulb.BindSSL': %w", err)
	}

	// 找出需要解绑的 SSL 证书
	sslIdsToUnbind := make([]string, 0)
	for _, sslItem := range vserverInfo.SSLSet {
		if sslItem.NotAfter != 0 && int64(sslItem.NotAfter) < time.Now().Unix() {
			sslIdsToUnbind = append(sslIdsToUnbind, sslItem.SSLId) // 过期证书需要解绑
			continue
		}

		d.sslId2PemMapMu.Lock()
		certX509, err := xcert.ParseCertificateFromPEM(d.sslId2PemMap[cloudCertId])
		d.sslId2PemMapMu.Unlock()
		if err != nil {
			continue
		}

		describeSSLV2Req := d.sdkClient.NewDescribeSSLV2Request()
		describeSSLV2Req.SSLId = ucloud.String(sslItem.SSLId)
		describeSSLV2Req.Limit = ucloud.Int(1)
		describeSSLV2Resp, err := d.sdkClient.DescribeSSLV2(describeSSLV2Req)
		d.logger.Debug("sdk request 'ulb.DescribeSSLV2'", slog.Any("request", describeSSLV2Req), slog.Any("response", describeSSLV2Resp))
		if err != nil {
			continue
		} else if len(describeSSLV2Resp.DataSet) == 0 {
			continue
		}

		sslItem := describeSSLV2Resp.DataSet[0]
		if sslItem.NotAfter != 0 && int64(sslItem.NotAfter) < time.Now().Unix() {
			sslIdsToUnbind = append(sslIdsToUnbind, sslItem.SSLId) // 过期证书需要解绑
			continue
		} else if sslItem.DNSNames != "" && slices.Equal(strings.Split(sslItem.DNSNames, ","), certX509.DNSNames) {
			sslIdsToUnbind = append(sslIdsToUnbind, sslItem.SSLId) // 同域名证书需要解绑
			continue
		}
	}

	// 解绑 SSL 证书
	// REF: https://docs.ucloud.cn/api/ulb-api/unbind_ssl
	for _, sslId := range sslIdsToUnbind {
		unbindSSLReq := d.sdkClient.NewUnbindSSLRequest()
		unbindSSLReq.ULBId = ucloud.String(cloudLoadbalancerId)
		unbindSSLReq.VServerId = ucloud.String(cloudVServerId)
		unbindSSLReq.SSLId = ucloud.String(sslId)
		unbindSSLResp, err := d.sdkClient.UnbindSSL(unbindSSLReq)
		d.logger.Debug("sdk request 'ulb.UnbindSSL'", slog.Any("request", unbindSSLReq), slog.Any("response", unbindSSLResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'ulb.UnbindSSL': %w", err)
		}
	}

	return nil
}

func createSDKClient(privateKey, publicKey, projectId, region string) (*ucloudsdk.ULBClient, error) {
	if privateKey == "" {
		return nil, fmt.Errorf("ucloud: invalid private key")
	}
	if publicKey == "" {
		return nil, fmt.Errorf("ucloud: invalid public key")
	}

	cfg := ucloud.NewConfig()
	cfg.ProjectId = projectId
	cfg.Region = region

	credential := auth.NewCredential()
	credential.PrivateKey = privateKey
	credential.PublicKey = publicKey

	client := ucloudsdk.NewClient(&cfg, &credential)
	return client, nil
}
