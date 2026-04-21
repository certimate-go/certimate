package tencentcloudcos

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/tencentcloud-ssl"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-cos/internal"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type DeployerConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云地域。
	Region string `json:"region"`
	// 存储桶名。
	Bucket string `json:"bucket"`
	// 自定义域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *wSDKClients
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

type wSDKClients struct {
	SSL *internal.SslClient
}

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	clients, err := createSDKClients(config.SecretId, config.SecretKey, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		SecretId:  config.SecretId,
		SecretKey: config.SecretKey,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create certmgr: %w", err)
	}

	return &Deployer{
		config:     config,
		logger:     slog.Default(),
		sdkClient:  clients,
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM, issuerCertificatePEM string) (*deployer.DeployResult, error) {
	if d.config.Bucket == "" {
		return nil, errors.New("config `bucket` is required")
	}
	if d.config.Domain == "" {
		return nil, errors.New("config `domain` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 避免多次部署，否则会报错 https://github.com/certimate-go/certimate/issues/897#issuecomment-3182904098
	if bind, _ := d.checkIsBind(ctx, upres.CertId); bind {
		d.logger.Info("ssl certificate already deployed")
		return &deployer.DeployResult{}, nil
	}

	// 证书部署到 COS 实例
	// REF: https://cloud.tencent.com/document/api/400/91667
	deployCertificateInstanceReq := tcssl.NewDeployCertificateInstanceRequest()
	deployCertificateInstanceReq.CertificateId = common.StringPtr(upres.CertId)
	deployCertificateInstanceReq.ResourceType = common.StringPtr("cos")
	deployCertificateInstanceReq.Status = common.Int64Ptr(1)
	deployCertificateInstanceReq.InstanceIdList = common.StringPtrs([]string{fmt.Sprintf("%s|%s|%s", d.config.Region, d.config.Bucket, d.config.Domain)})
	deployCertificateInstanceResp, err := d.sdkClient.SSL.DeployCertificateInstance(deployCertificateInstanceReq)
	d.logger.Debug("sdk request 'ssl.DeployCertificateInstance'", slog.Any("request", deployCertificateInstanceReq), slog.Any("response", deployCertificateInstanceResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'ssl.DeployCertificateInstance': %w", err)
	}

	// 获取部署任务详情，等待任务状态变更
	// REF: https://cloud.tencent.com/document/api/400/91658
	if _, err := xwait.UntilWithContext(ctx, func(_ context.Context, _ int) (bool, error) {
		describeHostDeployRecordDetailReq := tcssl.NewDescribeHostDeployRecordDetailRequest()
		describeHostDeployRecordDetailReq.DeployRecordId = common.StringPtr(fmt.Sprintf("%d", *deployCertificateInstanceResp.Response.DeployRecordId))
		describeHostDeployRecordDetailResp, err := d.sdkClient.SSL.DescribeHostDeployRecordDetail(describeHostDeployRecordDetailReq)
		d.logger.Debug("sdk request 'ssl.DescribeHostDeployRecordDetail'", slog.Any("request", describeHostDeployRecordDetailReq), slog.Any("response", describeHostDeployRecordDetailResp))
		if err != nil {
			return false, fmt.Errorf("failed to execute sdk request 'ssl.DescribeHostDeployRecordDetail': %w", err)
		}

		var pendingCount, runningCount, succeededCount, failedCount, totalCount int64
		if describeHostDeployRecordDetailResp.Response.TotalCount == nil {
			return false, fmt.Errorf("unexpected tencentcloud deployment job status")
		} else {
			pendingCount = lo.FromPtr(describeHostDeployRecordDetailResp.Response.PendingTotalCount)
			runningCount = lo.FromPtr(describeHostDeployRecordDetailResp.Response.RunningTotalCount)
			succeededCount = lo.FromPtr(describeHostDeployRecordDetailResp.Response.SuccessTotalCount)
			failedCount = lo.FromPtr(describeHostDeployRecordDetailResp.Response.FailedTotalCount)
			totalCount = lo.FromPtr(describeHostDeployRecordDetailResp.Response.TotalCount)

			if succeededCount+failedCount == totalCount {
				if failedCount > 0 {
					return false, fmt.Errorf("tencentcloud deployment job failed (succeeded: %d, failed: %d, total: %d)", succeededCount, failedCount, totalCount)
				}
				return true, nil
			}
		}

		d.logger.Info(fmt.Sprintf("waiting for tencentcloud deployment job completion (pending: %d, running: %d, succeeded: %d, failed: %d, total: %d) ...", pendingCount, runningCount, succeededCount, failedCount, totalCount))
		return false, nil
	}, time.Second*5); err != nil {
		return nil, err
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) checkIsBind(ctx context.Context, cloudCertId string) (bool, error) {
	// 查询证书 COS 云资源部署实例列表
	// REF: https://cloud.tencent.com/document/api/400/91661
	describeHostCosInstanceListLimit := 100
	describeHostCosInstanceListOffset := 0
	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		describeHostCosInstanceListReq := tcssl.NewDescribeHostCosInstanceListRequest()
		describeHostCosInstanceListReq.OldCertificateId = common.StringPtr(cloudCertId)
		describeHostCosInstanceListReq.ResourceType = common.StringPtr("cos")
		describeHostCosInstanceListReq.IsCache = common.Uint64Ptr(0)
		describeHostCosInstanceListReq.Offset = common.Int64Ptr(int64(describeHostCosInstanceListOffset))
		describeHostCosInstanceListReq.Limit = common.Int64Ptr(int64(describeHostCosInstanceListLimit))
		describeHostCosInstanceListResp, err := d.sdkClient.SSL.DescribeHostCosInstanceList(describeHostCosInstanceListReq)
		d.logger.Debug("sdk request 'ssl.DescribeHostCosInstanceList'", slog.Any("request", describeHostCosInstanceListReq), slog.Any("response", describeHostCosInstanceListResp))
		if err != nil {
			return false, fmt.Errorf("failed to execute sdk request 'ssl.DescribeHostCosInstanceList': %w", err)
		}

		if describeHostCosInstanceListResp.Response == nil {
			break
		}

		for _, instance := range describeHostCosInstanceListResp.Response.InstanceList {
			if lo.FromPtr(instance.Bucket) != d.config.Bucket {
				continue
			}
			if lo.FromPtr(instance.Domain) != d.config.Domain {
				continue
			}
			if lo.FromPtr(instance.Status) != "ENABLED" {
				continue
			}
			return true, nil
		}

		if len(describeHostCosInstanceListResp.Response.InstanceList) < describeHostCosInstanceListLimit {
			break
		}

		describeHostCosInstanceListOffset += describeHostCosInstanceListLimit
	}

	return false, nil
}

func createSDKClients(secretId, secretKey, region string) (*wSDKClients, error) {
	credential := common.NewCredential(secretId, secretKey)
	client, err := internal.NewSslClient(credential, region, profile.NewClientProfile())
	if err != nil {
		return nil, err
	}

	return &wSDKClients{
		SSL: client,
	}, nil
}
