package volcenginecertcenter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	ve "github.com/volcengine/volcengine-go-sdk/volcengine"
	vesession "github.com/volcengine/volcengine-go-sdk/volcengine/session"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	vecertificateservice "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/volcengine/volcengine-go-sdk/service/certificateservice"
)

type CertmgrConfig struct {
	// 火山引擎 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 火山引擎 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 火山引擎地域。
	Region string `json:"region"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *vecertificateservice.CERTIFICATESERVICE
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, errors.New("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Certmgr{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (c *Certmgr) SetLogger(logger *slog.Logger) {
	if logger == nil {
		c.logger = slog.New(slog.DiscardHandler)
	} else {
		c.logger = logger
	}
}

func (c *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*certmgr.UploadResult, error) {
	// 上传证书
	// REF: https://www.volcengine.com/docs/6638/1365580
	importCertificateReq := &vecertificateservice.ImportCertificateInput{
		CertificateInfo: &vecertificateservice.CertificateInfoForImportCertificateInput{
			CertificateChain: ve.String(certPEM),
			PrivateKey:       ve.String(privkeyPEM),
		},
		Repeatable: ve.Bool(false),
	}
	importCertificateResp, err := c.sdkClient.ImportCertificateWithContext(ctx, importCertificateReq)
	c.logger.Debug("sdk request 'certcenter.ImportCertificate'", slog.Any("request", importCertificateReq), slog.Any("response", importCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'certcenter.ImportCertificate': %w", err)
	}

	var sslId string
	if importCertificateResp.InstanceId != nil && *importCertificateResp.InstanceId != "" {
		sslId = *importCertificateResp.InstanceId
	}
	if importCertificateResp.RepeatId != nil && *importCertificateResp.RepeatId != "" {
		sslId = *importCertificateResp.RepeatId
	}

	if sslId == "" {
		return nil, errors.New("received empty certificate id, both `InstanceId` and `RepeatId` are empty")
	}

	return &certmgr.UploadResult{
		CertId: sslId,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	return nil, certmgr.ErrUnsupported
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*vecertificateservice.CERTIFICATESERVICE, error) {
	if region == "" {
		region = "cn-beijing" // 证书中心默认区域：北京
	}

	config := ve.NewConfig().
		WithAkSk(accessKeyId, accessKeySecret).
		WithRegion(region)

	session, err := vesession.NewSession(config)
	if err != nil {
		return nil, err
	}

	client := vecertificateservice.New(session, config)
	return client, nil
}
