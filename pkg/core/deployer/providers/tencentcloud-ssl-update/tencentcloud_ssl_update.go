package tencentcloudsslupdate

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/tencentcloud-ssl"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-ssl-update/internal"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type DeployerConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云接口端点。
	Endpoint string `json:"endpoint,omitempty"`
	// 原证书 ID。
	CertificateId string `json:"certificateId"`
	// 是否替换原有证书（即保持原证书 ID 不变）。
	IsReplaced bool `json:"isReplaced,omitempty"`
	// 云产品类型数组。
	ResourceProducts []string `json:"resourceProducts"`
	// 云产品地域数组。
	ResourceRegions []string `json:"resourceRegions"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *internal.SslClient
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.SecretId, config.SecretKey, config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		SecretId:  config.SecretId,
		SecretKey: config.SecretKey,
		Endpoint:  config.Endpoint,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create certmgr: %w", err)
	}

	return &Deployer{
		config:     config,
		logger:     slog.Default(),
		sdkClient:  client,
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
	if d.config.CertificateId == "" {
		return nil, errors.New("config `certificateId` is required")
	}
	if len(d.config.ResourceProducts) == 0 {
		return nil, errors.New("config `resourceProducts` is required")
	}

	if d.config.IsReplaced {
		if err := d.executeUploadUpdateCertificateInstance(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}
	} else {
		if err := d.executeUpdateCertificateInstance(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) executeUpdateCertificateInstance(ctx context.Context, certPEM, privkeyPEM string) error {
	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 一键更新新旧证书资源
	// REF: https://cloud.tencent.com/document/product/400/91649
	var deployRecordId string
	if _, err := xwait.UntilWithContext(ctx, func(_ context.Context, _ int) (bool, error) {
		updateCertificateInstanceReq := tcssl.NewUpdateCertificateInstanceRequest()
		updateCertificateInstanceReq.OldCertificateId = common.StringPtr(d.config.CertificateId)
		updateCertificateInstanceReq.CertificateId = common.StringPtr(upres.CertId)
		updateCertificateInstanceReq.ResourceTypes = common.StringPtrs(d.config.ResourceProducts)
		updateCertificateInstanceReq.ResourceTypesRegions = wrapResourceProductRegions(d.config.ResourceProducts, d.config.ResourceRegions)
		updateCertificateInstanceResp, err := d.sdkClient.UpdateCertificateInstance(updateCertificateInstanceReq)
		d.logger.Debug("sdk request 'ssl.UpdateCertificateInstance'", slog.Any("request", updateCertificateInstanceReq), slog.Any("response", updateCertificateInstanceResp))
		if err != nil {
			return false, fmt.Errorf("failed to execute sdk request 'ssl.UpdateCertificateInstance': %w", err)
		}

		if updateCertificateInstanceResp.Response.DeployStatus == nil || updateCertificateInstanceResp.Response.DeployRecordId == nil {
			return false, fmt.Errorf("unexpected deployment job status")
		} else if *updateCertificateInstanceResp.Response.DeployRecordId > 0 {
			deployRecordId = fmt.Sprintf("%d", *updateCertificateInstanceResp.Response.DeployRecordId)
			return true, nil
		}

		return false, nil
	}, time.Second*5); err != nil {
		return err
	}

	// 查询证书云资源更新记录详情，等待任务状态变更
	// REF: https://cloud.tencent.com/document/api/400/91652
	if _, err := xwait.UntilWithContext(ctx, func(_ context.Context, _ int) (bool, error) {
		describeHostUpdateRecordDetailReq := tcssl.NewDescribeHostUpdateRecordDetailRequest()
		describeHostUpdateRecordDetailReq.DeployRecordId = common.StringPtr(deployRecordId)
		describeHostUpdateRecordDetailReq.Limit = common.StringPtr("200")
		describeHostUpdateRecordDetailResp, err := d.sdkClient.DescribeHostUpdateRecordDetail(describeHostUpdateRecordDetailReq)
		d.logger.Debug("sdk request 'ssl.DescribeHostUpdateRecordDetail'", slog.Any("request", describeHostUpdateRecordDetailReq), slog.Any("response", describeHostUpdateRecordDetailResp))
		if err != nil {
			return false, fmt.Errorf("failed to execute sdk request 'ssl.DescribeHostUpdateRecordDetail': %w", err)
		}

		var pendingCount, runningCount, succeededCount, failedCount, totalCount int64
		if describeHostUpdateRecordDetailResp.Response.TotalCount == nil {
			return false, fmt.Errorf("unexpected tencentcloud deployment job status")
		} else {
			pendingCount = lo.FromPtr(describeHostUpdateRecordDetailResp.Response.PendingTotalCount)
			runningCount = lo.FromPtr(describeHostUpdateRecordDetailResp.Response.RunningTotalCount)
			succeededCount = lo.FromPtr(describeHostUpdateRecordDetailResp.Response.SuccessTotalCount)
			failedCount = lo.FromPtr(describeHostUpdateRecordDetailResp.Response.FailedTotalCount)
			totalCount = lo.FromPtr(describeHostUpdateRecordDetailResp.Response.TotalCount)

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
		return err
	}

	return nil
}

func (d *Deployer) executeUploadUpdateCertificateInstance(ctx context.Context, certPEM, privkeyPEM string) error {
	// 更新证书内容并更新关联的云资源
	// REF: https://cloud.tencent.com/document/product/400/119791
	var deployRecordId int64
	if _, err := xwait.UntilWithContext(ctx, func(_ context.Context, _ int) (bool, error) {
		uploadUpdateCertificateInstanceReq := tcssl.NewUploadUpdateCertificateInstanceRequest()
		uploadUpdateCertificateInstanceReq.OldCertificateId = common.StringPtr(d.config.CertificateId)
		uploadUpdateCertificateInstanceReq.CertificatePublicKey = common.StringPtr(certPEM)
		uploadUpdateCertificateInstanceReq.CertificatePrivateKey = common.StringPtr(privkeyPEM)
		uploadUpdateCertificateInstanceReq.ResourceTypes = common.StringPtrs(d.config.ResourceProducts)
		uploadUpdateCertificateInstanceReq.ResourceTypesRegions = wrapResourceProductRegions(d.config.ResourceProducts, d.config.ResourceRegions)
		uploadUpdateCertificateInstanceResp, err := d.sdkClient.UploadUpdateCertificateInstance(uploadUpdateCertificateInstanceReq)
		d.logger.Debug("sdk request 'ssl.UploadUpdateCertificateInstance'", slog.Any("request", uploadUpdateCertificateInstanceReq), slog.Any("response", uploadUpdateCertificateInstanceResp))
		if err != nil {
			return false, fmt.Errorf("failed to execute sdk request 'ssl.UploadUpdateCertificateInstance': %w", err)
		}

		if uploadUpdateCertificateInstanceResp.Response.DeployStatus == nil {
			return false, fmt.Errorf("unexpected deployment job status")
		} else if *uploadUpdateCertificateInstanceResp.Response.DeployStatus == 1 {
			deployRecordId = int64(*uploadUpdateCertificateInstanceResp.Response.DeployRecordId)
			return true, nil
		}

		return false, nil
	}, time.Second*5); err != nil {
		return err
	}

	// 查询证书云资源更新记录详情，等待任务状态变更
	// REF: https://cloud.tencent.com/document/product/400/120056
	if _, err := xwait.UntilWithContext(ctx, func(_ context.Context, _ int) (bool, error) {
		describeHostUploadUpdateRecordDetailReq := tcssl.NewDescribeHostUploadUpdateRecordDetailRequest()
		describeHostUploadUpdateRecordDetailReq.DeployRecordId = common.Int64Ptr(deployRecordId)
		describeHostUploadUpdateRecordDetailReq.Limit = common.Int64Ptr(200)
		describeHostUploadUpdateRecordDetailResp, err := d.sdkClient.DescribeHostUploadUpdateRecordDetail(describeHostUploadUpdateRecordDetailReq)
		d.logger.Debug("sdk request 'ssl.DescribeHostUploadUpdateRecordDetail'", slog.Any("request", describeHostUploadUpdateRecordDetailReq), slog.Any("response", describeHostUploadUpdateRecordDetailResp))
		if err != nil {
			return false, fmt.Errorf("failed to execute sdk request 'ssl.DescribeHostUploadUpdateRecordDetail': %w", err)
		}

		var runningCount, succeededCount, failedCount, totalCount int64
		if describeHostUploadUpdateRecordDetailResp.Response.DeployRecordDetail == nil {
			return false, fmt.Errorf("unexpected tencentcloud deployment job status")
		} else {
			for _, record := range describeHostUploadUpdateRecordDetailResp.Response.DeployRecordDetail {
				runningCount += lo.FromPtr(record.RunningTotalCount)
				succeededCount += lo.FromPtr(record.SuccessTotalCount)
				failedCount += lo.FromPtr(record.FailedTotalCount)
				totalCount += lo.FromPtr(record.TotalCount)
			}

			if succeededCount+failedCount == totalCount {
				if failedCount > 0 {
					return false, fmt.Errorf("tencentcloud deployment job failed (succeeded: %d, failed: %d, total: %d)", succeededCount, failedCount, totalCount)
				}
				return true, nil
			}
		}

		d.logger.Info(fmt.Sprintf("waiting for tencentcloud deployment job completion (running: %d, succeeded: %d, failed: %d, total: %d) ...", runningCount, succeededCount, failedCount, totalCount))
		return false, nil
	}, time.Second*5); err != nil {
		return err
	}

	return nil
}

func createSDKClient(secretId, secretKey, endpoint string) (*internal.SslClient, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := internal.NewSslClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func wrapResourceProductRegions(resourceProducts, resourceRegions []string) []*tcssl.ResourceTypeRegions {
	if len(resourceProducts) == 0 || len(resourceRegions) == 0 {
		return nil
	}

	// 仅以下云产品类型支持地域
	resourceProductsRequireRegion := []string{"apigateway", "clb", "cos", "tcb", "tke", "tse", "waf"}

	temp := make([]*tcssl.ResourceTypeRegions, 0)
	for _, resourceProduct := range resourceProducts {
		if slices.Contains(resourceProductsRequireRegion, resourceProduct) {
			temp = append(temp, &tcssl.ResourceTypeRegions{
				ResourceType: common.StringPtr(resourceProduct),
				Regions:      common.StringPtrs(resourceRegions),
			})
		}
	}

	return temp
}
