package byteplusmedialive

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	bp "github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus"
	bplive "github.com/byteplus-sdk/byteplus-sdk-golang/service/live/v20230101"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertmgrConfig struct {
	// BytePlus AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// BytePlus SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// BytePlus 项目名称。
	ProjectName string `json:"projectName,omitempty"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *bplive.Live
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
	}

	client := bplive.NewInstance()
	client.SetAccessKey(config.AccessKeyId)
	client.SetSecretKey(config.SecretAccessKey)

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
	// 查询证书列表，避免重复上传
	// REF: https://docs.byteplus.com/en/docs/byteplus-media-live/docs-listcertv2
	listCertPageNum := 1
	listCertPageSize := 10
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCertReq := &bplive.ListCertV2Body{}
		listCertReq.ProjectName = lo.EmptyableToPtr(c.config.ProjectName)
		listCertReq.PageNum = bp.Int32(int32(listCertPageNum))
		listCertReq.PageSize = bp.Int32(int32(listCertPageSize))
		listCertResp, err := c.sdkClient.ListCertV2(listCertReq)
		c.logger.Debug("sdk request 'medialive.ListCertV2'", slog.Any("request", listCertReq), slog.Any("response", listCertResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'medialive.ListCertV2': %w", err)
		}

		for _, certItem := range listCertResp.Result.CertList {
			// 查询证书详细信息
			// REF: https://docs.byteplus.com/en/docs/byteplus-media-live/docs-describecertdetailsecretv2-2023
			describeCertDetailSecretReq := &bplive.DescribeCertDetailSecretV2Body{
				ChainID: bp.String(certItem.ChainID),
			}
			describeCertDetailSecretResp, err := c.sdkClient.DescribeCertDetailSecretV2(describeCertDetailSecretReq)
			c.logger.Debug("sdk request 'medialive.DescribeCertDetailSecretV2'", slog.Any("request", describeCertDetailSecretReq), slog.Any("response", describeCertDetailSecretResp))
			if err != nil {
				continue
			}

			// 如果已存在相同证书，直接返回
			oldCertPEM := strings.Join(describeCertDetailSecretResp.Result.SSL.Chain, "\n\n")
			if xcert.EqualCertificatesFromPEM(certPEM, oldCertPEM) {
				c.logger.Info("ssl certificate already exists")
				return &certmgr.UploadResult{
					CertId:   certItem.ChainID,
					CertName: certItem.CertName,
				}, nil
			}
		}

		if len(listCertResp.Result.CertList) < listCertPageSize {
			break
		}

		listCertPageNum++
	}

	// 生成新证书名（需符合 BytePlus 命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 添加证书
	// REF: https://docs.byteplus.com/en/docs/byteplus-media-live/docs-createcert-2023
	createCertReq := &bplive.CreateCertBody{
		ProjectName: lo.EmptyableToPtr(c.config.ProjectName),
		CertName:    bp.String(certName),
		Rsa: bplive.CreateCertBodyRsa{
			Prikey: privkeyPEM,
			Pubkey: certPEM,
		},
		UseWay: "https",
	}
	createCertResp, err := c.sdkClient.CreateCert(createCertReq)
	c.logger.Debug("sdk request 'medialive.CreateCert'", slog.Any("request", createCertReq), slog.Any("response", createCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'medialive.CreateCert': %w", err)
	}

	return &certmgr.UploadResult{
		CertId:   *createCertResp.Result.ChainID,
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.ReplaceResult, error) {
	// 更新证书
	// REF: https://docs.byteplus.com/en/docs/byteplus-media-live/docs-createcert-2023
	createCertReq := &bplive.CreateCertBody{
		ProjectName: lo.EmptyableToPtr(c.config.ProjectName),
		ChainID:     bp.String(certIdOrName),
		Rsa: bplive.CreateCertBodyRsa{
			Prikey: privkeyPEM,
			Pubkey: certPEM,
		},
		UseWay: "https",
	}
	createCertResp, err := c.sdkClient.CreateCert(createCertReq)
	c.logger.Debug("sdk request 'medialive.CreateCert'", slog.Any("request", createCertReq), slog.Any("response", createCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'medialive.CreateCert': %w", err)
	}

	return &certmgr.ReplaceResult{}, nil
}
