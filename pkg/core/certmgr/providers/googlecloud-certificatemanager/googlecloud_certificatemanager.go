package googlecloudcertificatemanager

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/oauth2/google"
	gcpcm "google.golang.org/api/certificatemanager/v1"
	gcpoption "google.golang.org/api/option"

	"github.com/certimate-go/certimate/pkg/core"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xgcp "github.com/certimate-go/certimate/pkg/utils/third-party/gcp"
)

type (
	Provider      = core.Certmgr
	UploadResult  = core.CertmgrUploadResult
	ReplaceResult = core.CertmgrReplaceResult
)

type CertmgrConfig struct {
	// GCP 项目 ID。
	ProjectId string `json:"projectId"`
	// GCP 服务账号密钥。
	ServiceAccountKey string `json:"serviceAccountKey"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *gcpcm.Service
}

var _ Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.ServiceAccountKey)
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

func (c *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*UploadResult, error) {
	// 生成 GCP 所需的证书参数
	gcpProject, err := xgcp.GetProjectIDFromServiceAccountKey(c.config.ServiceAccountKey)
	gcpLocation := "global"
	gcpParent := fmt.Sprintf("projects/%s/locations/%s", gcpProject, gcpLocation)
	if err != nil {
		return nil, err
	} else if gcpProject != c.config.ProjectId {
		return nil, fmt.Errorf("invalid project ID: expected '%s', got '%s'", c.config.ProjectId, gcpProject)
	}

	// 获取证书列表，避免重复上传
	// REF: https://docs.cloud.google.com/certificate-manager/docs/reference/certificate-manager/rest/v1/projects.locations.certificates/list
	listCertificatesNextPageToken := ""
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCertificatesResp, err := c.sdkClient.Projects.Locations.Certificates.
			List(gcpParent).
			Context(ctx).
			PageSize(100).
			PageToken(listCertificatesNextPageToken).
			Do()
		c.logger.Debug("sdk request 'certificatemanager.projects.locations.certificates.list'", slog.Any("response", listCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'certificatemanager.projects.locations.certificates.list': %w", err)
		}

		for _, certItem := range listCertificatesResp.Certificates {
			// 如果已存在相同证书，直接返回
			if xcert.EqualCertificatesFromPEM(certPEM, certItem.PemCertificate) {
				c.logger.Info("ssl certificate already exists")
				return &UploadResult{
					CertId:   certItem.Name,
					CertName: certItem.Description,
				}, nil
			}
		}

		if len(listCertificatesResp.Certificates) == 0 || listCertificatesResp.NextPageToken == "" {
			break
		}

		listCertificatesNextPageToken = listCertificatesResp.NextPageToken
	}

	// 生成新证书 ID 与证书名（需符合 GCP 命名规则）
	certId := fmt.Sprintf("certimate-%d", time.Now().Unix())
	certDesc := certId

	// 创建证书
	// REF: https://docs.cloud.google.com/certificate-manager/docs/reference/certificate-manager/rest/v1/projects.locations.certificates/create
	createCertificateResp, err := c.sdkClient.Projects.Locations.Certificates.
		Create(gcpParent, &gcpcm.Certificate{
			Description: certDesc,
			SelfManaged: &gcpcm.SelfManagedCertificate{
				PemCertificate: certPEM,
				PemPrivateKey:  privkeyPEM,
			},
		}).
		Context(ctx).
		CertificateId(certId).
		Do()
	c.logger.Debug("sdk request 'certificatemanager.projects.locations.certificates.create'", slog.Any("response", createCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'certificatemanager.projects.locations.certificates.create': %w", err)
	}

	return &UploadResult{
		CertId:   fmt.Sprintf("%s/certificates/%s", gcpParent, certId),
		CertName: certDesc,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*ReplaceResult, error) {
	return nil, core.ErrUnsupported
}

func createSDKClient(serviceAccountKey string) (*gcpcm.Service, error) {
	saKey := []byte(serviceAccountKey)
	saConf, err := google.JWTConfigFromJSON(saKey, gcpcm.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("unable to acquire service account config: %w", err)
	}

	ctx := context.Background()
	tokenSource := gcpoption.WithTokenSource(saConf.TokenSource(ctx))
	service, err := gcpcm.NewService(ctx, tokenSource)
	if err != nil {
		return nil, err
	}

	return service, nil
}
