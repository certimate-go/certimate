package certapply

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/mail"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
)

type TencentCloudConfig struct {
	SecretId    string
	SecretKey   string
	ProjectId   int64
	DNSProvider DNSProvider
}

type TencentCloudRequest struct {
	Domains            []string
	IPAddrs            []string
	ContactEmail       string
	KeySource          string
	KeyAlgorithm       string
	CertificateId      string
	DvAuthMethod       string
	DNSPropagationWait int
}

type TencentCloudResult struct {
	CertificateId string
	Certificate   string
	PrivateKey    string
}

type TencentCloudStatusError struct {
	CertificateId string
	Status        uint64
	Message       string
}

func (e *TencentCloudStatusError) Error() string {
	return fmt.Sprintf("Tencent Cloud certificate %s cannot be issued (status %d): %s", e.CertificateId, e.Status, e.Message)
}

func (e *TencentCloudStatusError) IsTerminal() bool {
	switch e.Status {
	case 2, 3, 7, 10, 14:
		return true
	default:
		return false
	}
}

type tencentCloudSSLAPI interface {
	ApplyCertificateWithContext(context.Context, *ssl.ApplyCertificateRequest) (*ssl.ApplyCertificateResponse, error)
	DescribeCertificateDetailWithContext(context.Context, *ssl.DescribeCertificateDetailRequest) (*ssl.DescribeCertificateDetailResponse, error)
	CompleteCertificateWithContext(context.Context, *ssl.CompleteCertificateRequest) (*ssl.CompleteCertificateResponse, error)
	DownloadCertificateWithContext(context.Context, *ssl.DownloadCertificateRequest) (*ssl.DownloadCertificateResponse, error)
}

type TencentCloudClient struct {
	client       tencentCloudSSLAPI
	projectId    uint64
	logger       *slog.Logger
	pollInterval time.Duration
	dnsProvider  DNSProvider
}

func NewTencentCloudClient(config TencentCloudConfig, logger *slog.Logger) (*TencentCloudClient, error) {
	if strings.TrimSpace(config.SecretId) == "" || strings.TrimSpace(config.SecretKey) == "" {
		return nil, fmt.Errorf("Tencent Cloud SecretId and SecretKey are required")
	}
	if config.ProjectId < 0 {
		return nil, fmt.Errorf("Tencent Cloud project ID must not be negative")
	}
	clientProfile := profile.NewClientProfile()
	clientProfile.HttpProfile.ReqTimeout = 30
	client, err := ssl.NewClient(common.NewCredential(config.SecretId, config.SecretKey), "", clientProfile)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Tencent Cloud SSL client: %w", err)
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &TencentCloudClient{client: client, projectId: uint64(config.ProjectId), logger: logger, pollInterval: 5 * time.Second, dnsProvider: config.DNSProvider}, nil
}

var domainLabel = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)

func (r *TencentCloudRequest) Validate() error {
	if r.DvAuthMethod != "" && r.DvAuthMethod != "DNS_AUTO" && r.DvAuthMethod != "DNS" {
		return fmt.Errorf("unsupported Tencent Cloud validation method: %q", r.DvAuthMethod)
	}
	if r.DNSPropagationWait < 0 || r.DNSPropagationWait > 3600 {
		return fmt.Errorf("DNS propagation wait must be between 0 and 3600 seconds")
	}
	if len(r.Domains) != 1 || len(r.IPAddrs) != 0 {
		return fmt.Errorf("Tencent Cloud free certificates require exactly one domain, without IP addresses or wildcard domains")
	}
	domain := r.Domains[0]
	labels := strings.Split(domain, ".")
	if len(domain) > 253 || len(labels) < 2 || net.ParseIP(domain) != nil {
		return fmt.Errorf("invalid Tencent Cloud certificate domain: %q", domain)
	}
	for _, label := range labels {
		if !domainLabel.MatchString(label) {
			return fmt.Errorf("invalid Tencent Cloud certificate domain: %q (wildcards are not supported)", domain)
		}
	}
	if r.KeySource != "" && r.KeySource != "auto" {
		return fmt.Errorf("Tencent Cloud generates the private key; custom and reused keys are not supported")
	}
	if r.KeyAlgorithm != "RSA2048" && r.KeyAlgorithm != "EC256" {
		return fmt.Errorf("Tencent Cloud free certificates only support RSA2048 and EC256")
	}
	if r.ContactEmail != "" {
		address, err := mail.ParseAddress(r.ContactEmail)
		if err != nil || address.Address != r.ContactEmail {
			return fmt.Errorf("invalid Tencent Cloud certificate contact email")
		}
	}
	return nil
}

func (c *TencentCloudClient) Obtain(ctx context.Context, request TencentCloudRequest, onCreated func(string) error) (*TencentCloudResult, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if request.DvAuthMethod == "" {
		request.DvAuthMethod = "DNS_AUTO"
	}
	if request.DvAuthMethod == "DNS" && c.dnsProvider == nil {
		return nil, fmt.Errorf("DNS provider is required for Tencent Cloud DNS validation")
	}
	certificateId := request.CertificateId
	if certificateId == "" {
		if request.DvAuthMethod == "DNS" {
			if checker, ok := c.dnsProvider.(dnsDomainChecker); ok {
				if err := checker.checkDomain(ctx, request.Domains[0]); err != nil {
					return nil, fmt.Errorf("Tencent Cloud DNS validation preflight failed: %w", err)
				}
			}
		}
		req := ssl.NewApplyCertificateRequest()
		req.DomainName = common.StringPtr(request.Domains[0])
		req.DvAuthMethod = common.StringPtr(request.DvAuthMethod)
		req.PackageType = common.StringPtr("83")
		req.ValidityPeriod = common.StringPtr("3")
		req.ProjectId = common.Uint64Ptr(c.projectId)
		if request.DvAuthMethod == "DNS_AUTO" {
			req.DeleteDnsAutoRecord = common.BoolPtr(true)
		}
		if request.ContactEmail != "" {
			req.ContactEmail = common.StringPtr(request.ContactEmail)
		}
		if request.KeyAlgorithm == "EC256" {
			req.CsrEncryptAlgo = common.StringPtr("ECC")
			req.CsrKeyParameter = common.StringPtr("prime256v1")
		} else {
			req.CsrEncryptAlgo = common.StringPtr("RSA")
			req.CsrKeyParameter = common.StringPtr("2048")
		}
		resp, err := c.client.ApplyCertificateWithContext(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("ssl:ApplyCertificate failed: %w", err)
		}
		if resp == nil || resp.Response == nil || resp.Response.CertificateId == nil || *resp.Response.CertificateId == "" {
			return nil, fmt.Errorf("ssl:ApplyCertificate returned an empty certificate ID")
		}
		certificateId = *resp.Response.CertificateId
		c.logger.Info("Tencent Cloud free certificate application created", slog.String("certificateId", certificateId))
		if onCreated != nil {
			if err := onCreated(certificateId); err != nil {
				return nil, fmt.Errorf("failed to persist Tencent Cloud certificate order %s: %w", certificateId, err)
			}
		}
	} else {
		c.logger.Info("resuming Tencent Cloud certificate application", slog.String("certificateId", certificateId))
	}

	var cleanups []func(context.Context) error
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()
		for i := len(cleanups) - 1; i >= 0; i-- {
			if err := cleanups[i](cleanupCtx); err != nil {
				c.logger.Warn("failed to clean up Tencent Cloud validation TXT record", slog.String("certificateId", certificateId), slog.Any("error", err))
			}
		}
	}()
	presented := make(map[DNSRecord]bool)
	validationTriggered := false
	for {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("waiting for Tencent Cloud certificate %s: %w", certificateId, err)
		}
		req := ssl.NewDescribeCertificateDetailRequest()
		req.CertificateId = common.StringPtr(certificateId)
		resp, err := c.client.DescribeCertificateDetailWithContext(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("ssl:DescribeCertificateDetail failed for %s: %w", certificateId, err)
		}
		if resp == nil || resp.Response == nil || resp.Response.Status == nil {
			return nil, fmt.Errorf("ssl:DescribeCertificateDetail returned no status for %s", certificateId)
		}
		detail := resp.Response
		c.logger.Debug("Tencent Cloud certificate status", slog.String("certificateId", certificateId), slog.Uint64("status", *detail.Status))
		switch *detail.Status {
		case 1:
			downloadReq := ssl.NewDownloadCertificateRequest()
			downloadReq.CertificateId = common.StringPtr(certificateId)
			downloadResp, err := c.client.DownloadCertificateWithContext(ctx, downloadReq)
			if err != nil {
				return nil, fmt.Errorf("ssl:DownloadCertificate failed for %s: %w", certificateId, err)
			}
			if downloadResp == nil || downloadResp.Response == nil || downloadResp.Response.Content == nil {
				return nil, fmt.Errorf("ssl:DownloadCertificate returned no archive for %s", certificateId)
			}
			return tencentCloudArchiveResult(certificateId, *downloadResp.Response.Content, request)
		case 0, 4, 6, 9, 11, 13, 15:
			if request.DvAuthMethod == "DNS" {
				if detail.VerifyType != nil && *detail.VerifyType != "" && *detail.VerifyType != "DNS" {
					return nil, fmt.Errorf("Tencent Cloud order %s does not use DNS validation", certificateId)
				}
				records, err := tencentCloudDNSRecords(detail.DvAuthDetail, request.Domains[0])
				if err != nil {
					return nil, err
				}
				if len(records) == 0 {
					break
				}
				changed := false
				for _, record := range records {
					if presented[record] {
						continue
					}
					cleanup, err := c.dnsProvider.Present(ctx, record)
					if cleanup != nil {
						cleanups = append(cleanups, cleanup)
					}
					if err != nil {
						return nil, fmt.Errorf("failed to present Tencent Cloud validation TXT record: %w", err)
					}
					presented[record], changed = true, true
				}
				if changed {
					validationTriggered = false
					timer := time.NewTimer(time.Duration(request.DNSPropagationWait) * time.Second)
					select {
					case <-ctx.Done():
						timer.Stop()
						return nil, ctx.Err()
					case <-timer.C:
					}
				}
			}
			if !validationTriggered && (*detail.Status == 0 || *detail.Status == 13) {
				completeReq := ssl.NewCompleteCertificateRequest()
				completeReq.CertificateId = common.StringPtr(certificateId)
				_, err := c.client.CompleteCertificateWithContext(ctx, completeReq)
				if err != nil {
					// The order may have been issued between the status query and this call.
					if sdkErr, ok := err.(*tcerr.TencentCloudSDKError); !ok || sdkErr.GetCode() != "FailedOperation.CertificateNotFoundOrAudited" {
						return nil, fmt.Errorf("ssl:CompleteCertificate failed for %s: %w", certificateId, err)
					}
				}
				validationTriggered = true
			}
		default:
			message := ""
			if detail.StatusMsg != nil {
				message = *detail.StatusMsg
			}
			return nil, &TencentCloudStatusError{CertificateId: certificateId, Status: *detail.Status, Message: message}
		}
		timer := time.NewTimer(c.pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("waiting for Tencent Cloud certificate %s: %w", certificateId, ctx.Err())
		case <-timer.C:
		}
	}
}

func tencentCloudArchiveResult(id, content string, request TencentCloudRequest) (*TencentCloudResult, error) {
	if len(content) > 32<<20 {
		return nil, fmt.Errorf("Tencent Cloud certificate %s archive is too large", id)
	}
	data, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, fmt.Errorf("invalid Tencent Cloud certificate %s archive encoding: %w", id, err)
	}
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("invalid Tencent Cloud certificate %s ZIP archive: %w", id, err)
	}
	files := make(map[string]string)
	totalSize := 0
	for _, file := range archive.File {
		name := strings.ToLower(file.Name)
		if file.FileInfo().IsDir() || (!strings.HasSuffix(name, "_bundle.crt") && !strings.HasSuffix(name, "_bundle.pem") && !strings.HasSuffix(name, ".key")) {
			continue
		}
		if file.UncompressedSize64 > 2<<20 {
			return nil, fmt.Errorf("Tencent Cloud certificate %s archive entry is too large", id)
		}
		reader, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to read Tencent Cloud certificate %s archive: %w", id, err)
		}
		contents, err := io.ReadAll(io.LimitReader(reader, (2<<20)+1))
		reader.Close()
		totalSize += len(contents)
		if err != nil || len(contents) > 2<<20 || totalSize > 16<<20 {
			return nil, fmt.Errorf("failed to read bounded Tencent Cloud certificate %s archive entry", id)
		}
		files[file.Name] = string(contents)
	}
	for _, file := range archive.File {
		name := strings.ToLower(file.Name)
		if !strings.HasSuffix(name, "_bundle.crt") && !strings.HasSuffix(name, "_bundle.pem") {
			continue
		}
		for keyName, privateKey := range files {
			if strings.HasSuffix(strings.ToLower(keyName), ".key") && path.Dir(keyName) == path.Dir(file.Name) {
				return tencentCloudCertificateResult(id, files[file.Name], privateKey, request)
			}
		}
	}
	return nil, fmt.Errorf("Tencent Cloud certificate %s archive has no Nginx bundle and private key", id)
}

func tencentCloudCertificateResult(id, certificate, privateKey string, request TencentCloudRequest) (*TencentCloudResult, error) {
	pair, err := tls.X509KeyPair([]byte(certificate), []byte(privateKey))
	if err != nil {
		return nil, fmt.Errorf("Tencent Cloud certificate %s has invalid certificate/key PEM: %w", id, err)
	}
	if len(pair.Certificate) < 2 {
		return nil, fmt.Errorf("Tencent Cloud certificate %s bundle is missing its issuer chain", id)
	}
	chain := make([]*x509.Certificate, len(pair.Certificate))
	for i, der := range pair.Certificate {
		chain[i], err = x509.ParseCertificate(der)
		if err != nil {
			return nil, fmt.Errorf("invalid Tencent Cloud certificate %s chain: %w", id, err)
		}
		if i > 0 {
			if err := chain[i-1].CheckSignatureFrom(chain[i]); err != nil {
				return nil, fmt.Errorf("invalid Tencent Cloud certificate %s issuer chain: %w", id, err)
			}
		}
	}
	if err := chain[0].VerifyHostname(request.Domains[0]); err != nil {
		return nil, fmt.Errorf("Tencent Cloud certificate %s does not match the requested domain: %w", id, err)
	}
	if !chain[0].NotAfter.After(time.Now()) {
		return nil, fmt.Errorf("Tencent Cloud certificate %s is expired", id)
	}
	validKey := false
	switch key := pair.PrivateKey.(type) {
	case *rsa.PrivateKey:
		validKey = request.KeyAlgorithm == "RSA2048" && key.N.BitLen() == 2048
	case *ecdsa.PrivateKey:
		validKey = request.KeyAlgorithm == "EC256" && key.Curve.Params().BitSize == 256
	}
	if !validKey {
		return nil, fmt.Errorf("Tencent Cloud certificate %s does not match the requested key algorithm", id)
	}
	return &TencentCloudResult{CertificateId: id, Certificate: certificate, PrivateKey: privateKey}, nil
}
