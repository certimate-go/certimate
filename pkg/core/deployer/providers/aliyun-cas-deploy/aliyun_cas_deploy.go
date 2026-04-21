package aliyuncasdeploy

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	alicas "github.com/alibabacloud-go/cas-20200407/v4/client"
	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-cas"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-cas-deploy/internal"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type DeployerConfig struct {
	// 阿里云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 阿里云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 阿里云资源组 ID。
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
	// 阿里云地域。
	Region string `json:"region"`
	// 云产品资源 ID 数组。
	ResourceIds []string `json:"resourceIds"`
	// 云联系人 ID 数组。
	// 零值时使用账号下第一个联系人。
	ContactIds []string `json:"contactIds"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *internal.CasClient
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		ResourceGroupId: config.ResourceGroupId,
		Region:          config.Region,
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
	if len(d.config.ResourceIds) == 0 {
		return nil, errors.New("config `resourceIds` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	contactIds := d.config.ContactIds
	if len(contactIds) == 0 {
		// 获取联系人列表
		// REF: https://help.aliyun.com/zh/ssl-certificate/developer-reference/api-cas-2020-04-07-listcontact
		listContactReq := &alicas.ListContactRequest{
			ShowSize:    tea.Int32(1),
			CurrentPage: tea.Int32(1),
		}
		listContactResp, err := d.sdkClient.ListContactWithContext(ctx, listContactReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'cas.ListContact'", slog.Any("request", listContactReq), slog.Any("response", listContactResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'cas.ListContact': %w", err)
		}

		if len(listContactResp.Body.ContactList) > 0 {
			contactIds = []string{fmt.Sprintf("%d", listContactResp.Body.ContactList[0].ContactId)}
		}
	}

	// 创建部署任务
	// REF: https://help.aliyun.com/zh/ssl-certificate/developer-reference/api-cas-2020-04-07-createdeploymentjob
	createDeploymentJobReq := &alicas.CreateDeploymentJobRequest{
		Name:        tea.String(fmt.Sprintf("certimate-%d", time.Now().UnixMilli())),
		JobType:     tea.String("user"),
		CertIds:     tea.String(upres.CertId),
		ResourceIds: tea.String(strings.Join(d.config.ResourceIds, ",")),
		ContactIds:  tea.String(strings.Join(contactIds, ",")),
	}
	createDeploymentJobResp, err := d.sdkClient.CreateDeploymentJobWithContext(ctx, createDeploymentJobReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'cas.CreateDeploymentJob'", slog.Any("request", createDeploymentJobReq), slog.Any("response", createDeploymentJobResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cas.CreateDeploymentJob': %w", err)
	}

	// 获取部署任务详情，等待任务状态变更
	// REF: https://help.aliyun.com/zh/ssl-certificate/developer-reference/api-cas-2020-04-07-describedeploymentjob
	if _, err := xwait.UntilWithContext(ctx, func(_ context.Context, _ int) (bool, error) {
		describeDeploymentJobReq := &alicas.DescribeDeploymentJobRequest{
			JobId: createDeploymentJobResp.Body.JobId,
		}
		describeDeploymentJobResp, err := d.sdkClient.DescribeDeploymentJobWithContext(ctx, describeDeploymentJobReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'cas.DescribeDeploymentJob'", slog.Any("request", describeDeploymentJobReq), slog.Any("response", describeDeploymentJobResp))
		if err != nil {
			return false, fmt.Errorf("failed to execute sdk request 'cas.DescribeDeploymentJob': %w", err)
		}

		switch tea.StringValue(describeDeploymentJobResp.Body.Status) {
		case "success", "error":
			return true, nil
		case "", "editing":
			return false, fmt.Errorf("unexpected aliyun deployment job status")
		}

		d.logger.Info("waiting for aliyun deployment job completion ...")
		return false, nil
	}, time.Second*5); err != nil {
		return nil, err
	}

	return &deployer.DeployResult{}, nil
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*internal.CasClient, error) {
	// 接入点一览 https://api.aliyun.com/product/cas
	var endpoint string
	switch region {
	case "", "cn-hangzhou":
		endpoint = "cas.aliyuncs.com"
	default:
		endpoint = fmt.Sprintf("cas.%s.aliyuncs.com", region)
	}

	config := &aliopen.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(endpoint),
	}

	client, err := internal.NewCasClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
