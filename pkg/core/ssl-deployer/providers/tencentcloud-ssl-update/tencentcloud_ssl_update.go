package tencentcloudsslupdate

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"

	"github.com/certimate-go/certimate/pkg/core"
	sslmgrsp "github.com/certimate-go/certimate/pkg/core/ssl-manager/providers/tencentcloud-ssl"
)

type SSLDeployerProviderConfig struct {
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
	// 云资源类型数组。
	ResourceTypes []string `json:"resourceTypes"`
	// 云资源地域数组。
	ResourceRegions []string `json:"resourceRegions"`
}

type SSLDeployerProvider struct {
	config     *SSLDeployerProviderConfig
	logger     *slog.Logger
	sdkClient  *tcssl.Client
	sslManager core.SSLManager
}

var _ core.SSLDeployer = (*SSLDeployerProvider)(nil)

func NewSSLDeployerProvider(config *SSLDeployerProviderConfig) (*SSLDeployerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl deployer provider is nil")
	}

	client, err := createSDKClient(config.SecretId, config.SecretKey, config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	sslmgr, err := sslmgrsp.NewSSLManagerProvider(&sslmgrsp.SSLManagerProviderConfig{
		SecretId:  config.SecretId,
		SecretKey: config.SecretKey,
		Endpoint:  config.Endpoint,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create ssl manager: %w", err)
	}

	return &SSLDeployerProvider{
		config:     config,
		logger:     slog.Default(),
		sdkClient:  client,
		sslManager: sslmgr,
	}, nil
}

func (d *SSLDeployerProvider) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}

	d.sslManager.SetLogger(logger)
}

func (d *SSLDeployerProvider) Deploy(ctx context.Context, certPEM string, privkeyPEM string) (*core.SSLDeployResult, error) {
	if d.config.CertificateId == "" {
		return nil, errors.New("config `certificateId` is required")
	}
	if len(d.config.ResourceTypes) == 0 {
		return nil, errors.New("config `resourceTypes` is required")
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

	return &core.SSLDeployResult{}, nil
}

func (d *SSLDeployerProvider) executeUpdateCertificateInstance(ctx context.Context, certPEM string, privkeyPEM string) error {
	// 上传证书
	upres, err := d.sslManager.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 一键更新新旧证书资源
	// REF: https://cloud.tencent.com/document/product/400/91649
	var deployRecordId string
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		updateCertificateInstanceReq := tcssl.NewUpdateCertificateInstanceRequest()
		updateCertificateInstanceReq.OldCertificateId = common.StringPtr(d.config.CertificateId)
		updateCertificateInstanceReq.CertificateId = common.StringPtr(upres.CertId)
		updateCertificateInstanceReq.ResourceTypes = common.StringPtrs(d.config.ResourceTypes)
		updateCertificateInstanceReq.ResourceTypesRegions = wrapResourceTypeRegions(d.config.ResourceTypes, d.config.ResourceRegions)
		updateCertificateInstanceResp, err := d.sdkClient.UpdateCertificateInstance(updateCertificateInstanceReq)
		d.logger.Debug("sdk request 'ssl.UpdateCertificateInstance'", slog.Any("request", updateCertificateInstanceReq), slog.Any("response", updateCertificateInstanceResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'ssl.UpdateCertificateInstance': %w", err)
		}

		if updateCertificateInstanceResp.Response.DeployStatus == nil {
			return errors.New("unexpected deployment job status")
		} else if *updateCertificateInstanceResp.Response.DeployStatus == 1 {
			deployRecordId = fmt.Sprintf("%d", *updateCertificateInstanceResp.Response.DeployRecordId)
			break
		}

		time.Sleep(time.Second * 5)
	}

	// 循环查询证书云资源更新记录详情，等待任务状态变更
	// REF: https://cloud.tencent.com/document/api/400/91652
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		describeHostUpdateRecordDetailReq := tcssl.NewDescribeHostUpdateRecordDetailRequest()
		describeHostUpdateRecordDetailReq.DeployRecordId = common.StringPtr(deployRecordId)
		describeHostUpdateRecordDetailResp, err := d.sdkClient.DescribeHostUpdateRecordDetail(describeHostUpdateRecordDetailReq)
		d.logger.Debug("sdk request 'ssl.DescribeHostUpdateRecordDetail'", slog.Any("request", describeHostUpdateRecordDetailReq), slog.Any("response", describeHostUpdateRecordDetailResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'ssl.DescribeHostUpdateRecordDetail': %w", err)
		}

		var runningCount, succeededCount, failedCount, totalCount int64
		if describeHostUpdateRecordDetailResp.Response.TotalCount == nil {
			return errors.New("unexpected deployment job status")
		} else {
			if describeHostUpdateRecordDetailResp.Response.RunningTotalCount != nil {
				runningCount = *describeHostUpdateRecordDetailResp.Response.RunningTotalCount
			}
			if describeHostUpdateRecordDetailResp.Response.SuccessTotalCount != nil {
				succeededCount = *describeHostUpdateRecordDetailResp.Response.SuccessTotalCount
			}
			if describeHostUpdateRecordDetailResp.Response.FailedTotalCount != nil {
				failedCount = *describeHostUpdateRecordDetailResp.Response.FailedTotalCount
			}
			if describeHostUpdateRecordDetailResp.Response.TotalCount != nil {
				totalCount = *describeHostUpdateRecordDetailResp.Response.TotalCount
			}

			if succeededCount+failedCount == totalCount {
				break
			}
		}

		d.logger.Info(fmt.Sprintf("waiting for deployment job completion (running: %d, succeeded: %d, failed: %d, total: %d) ...", runningCount, succeededCount, failedCount, totalCount))
		time.Sleep(time.Second * 5)
	}

	return nil
}

func (d *SSLDeployerProvider) executeUploadUpdateCertificateInstance(ctx context.Context, certPEM string, privkeyPEM string) error {
	// 更新证书内容并更新关联的云资源
	// REF: https://cloud.tencent.com/document/product/400/119791
	var deployRecordId int64
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		uploadUpdateCertificateInstanceReq := tcssl.NewUploadUpdateCertificateInstanceRequest()
		uploadUpdateCertificateInstanceReq.OldCertificateId = common.StringPtr(d.config.CertificateId)
		uploadUpdateCertificateInstanceReq.CertificatePublicKey = common.StringPtr(certPEM)
		uploadUpdateCertificateInstanceReq.CertificatePrivateKey = common.StringPtr(privkeyPEM)
		uploadUpdateCertificateInstanceReq.ResourceTypes = common.StringPtrs(d.config.ResourceTypes)
		uploadUpdateCertificateInstanceReq.ResourceTypesRegions = wrapResourceTypeRegions(d.config.ResourceTypes, d.config.ResourceRegions)
		uploadUpdateCertificateInstanceResp, err := d.sdkClient.UploadUpdateCertificateInstance(uploadUpdateCertificateInstanceReq)
		d.logger.Debug("sdk request 'ssl.UploadUpdateCertificateInstance'", slog.Any("request", uploadUpdateCertificateInstanceReq), slog.Any("response", uploadUpdateCertificateInstanceResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'ssl.UploadUpdateCertificateInstance': %w", err)
		}

		if uploadUpdateCertificateInstanceResp.Response.DeployStatus == nil {
			return errors.New("unexpected deployment job status")
		} else if *uploadUpdateCertificateInstanceResp.Response.DeployStatus == 1 {
			deployRecordId = int64(*uploadUpdateCertificateInstanceResp.Response.DeployRecordId)
			break
		}

		time.Sleep(time.Second * 5)
	}

	// 循环查询证书云资源更新记录详情，等待任务状态变更
	// REF: https://cloud.tencent.com/document/product/400/120056
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		describeHostUploadUpdateRecordDetailReq := tcssl.NewDescribeHostUploadUpdateRecordDetailRequest()
		describeHostUploadUpdateRecordDetailReq.DeployRecordId = common.Int64Ptr(deployRecordId)
		describeHostUploadUpdateRecordDetailReq.Limit = common.Int64Ptr(200)
		describeHostUploadUpdateRecordDetailResp, err := d.sdkClient.DescribeHostUploadUpdateRecordDetail(describeHostUploadUpdateRecordDetailReq)
		d.logger.Debug("sdk request 'ssl.DescribeHostUploadUpdateRecordDetail'", slog.Any("request", describeHostUploadUpdateRecordDetailReq), slog.Any("response", describeHostUploadUpdateRecordDetailResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'ssl.DescribeHostUploadUpdateRecordDetail': %w", err)
		}

		var runningCount, succeededCount, failedCount, totalCount int64
		if describeHostUploadUpdateRecordDetailResp.Response.DeployRecordDetail == nil {
			return errors.New("unexpected deployment job status")
		} else {
			for _, record := range describeHostUploadUpdateRecordDetailResp.Response.DeployRecordDetail {
				if record.RunningTotalCount != nil {
					runningCount = *record.RunningTotalCount
				}
				if record.SuccessTotalCount != nil {
					succeededCount = *record.SuccessTotalCount
				}
				if record.FailedTotalCount != nil {
					failedCount = *record.FailedTotalCount
				}
				if record.TotalCount != nil {
					totalCount = *record.TotalCount
				}
			}

			if succeededCount+failedCount == totalCount {
				break
			}
		}

		d.logger.Info(fmt.Sprintf("waiting for deployment job completion (running: %d, succeeded: %d, failed: %d, total: %d) ...", runningCount, succeededCount, failedCount, totalCount))
		time.Sleep(time.Second * 5)
	}

	return nil
}

func createSDKClient(secretId, secretKey, endpoint string) (*tcssl.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := tcssl.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func wrapResourceTypeRegions(resourceTypes, resourceRegions []string) []*tcssl.ResourceTypeRegions {
	if len(resourceTypes) == 0 || len(resourceRegions) == 0 {
		return nil
	}

	// 仅以下云资源类型支持地域
	resourceTypesRequireRegion := []string{"apigateway", "clb", "cos", "tcb", "tke", "tse", "waf"}

	temp := make([]*tcssl.ResourceTypeRegions, 0)
	for _, resourceType := range resourceTypes {
		if slices.Contains(resourceTypesRequireRegion, resourceType) {
			temp = append(temp, &tcssl.ResourceTypeRegions{
				ResourceType: common.StringPtr(resourceType),
				Regions:      common.StringPtrs(resourceRegions),
			})
		}
	}

	return temp
}
