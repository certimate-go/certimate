package f5bigip

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/certimate-go/certimate/pkg/core"
	f5sdk "github.com/certimate-go/certimate/pkg/sdk3rd/f5bigip"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	ServerUrl                string `json:"serverUrl"`
	Username                 string `json:"username"`
	Password                 string `json:"password"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
	DeployTarget             string `json:"deployTarget"`
	Partition                string `json:"partition,omitempty"`
	ClientSSLProfileName     string `json:"clientSSLProfileName,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *f5sdk.Client
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.AllowInsecureConnections)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Deployer{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*DeployResult, error) {
	partition := d.config.Partition
	if partition == "" {
		partition = "Common"
	}

	if err := d.sdkClient.Login(ctx, d.config.Username, d.config.Password); err != nil {
		return nil, fmt.Errorf("failed to login to F5 Big-IP: %w", err)
	}

	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	leafCertPEM, chainCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certificates from PEM: %w", err)
	}

	certName := buildCertName(certX509)
	d.logger.Info("deploying certificate to F5 Big-IP", slog.String("certName", certName), slog.String("subject", certX509.Subject.CommonName))

	if err := d.sdkClient.UploadCertificate(ctx, certName, partition, leafCertPEM); err != nil {
		return nil, fmt.Errorf("failed to upload certificate: %w", err)
	}
	d.logger.Info("certificate uploaded", slog.String("certName", certName))

	if err := d.sdkClient.UploadKey(ctx, certName, partition, privkeyPEM); err != nil {
		return nil, fmt.Errorf("failed to upload private key: %w", err)
	}
	d.logger.Info("private key uploaded", slog.String("keyName", certName))

	var chainPath string
	if chainCertPEM != "" {
		chainName := certName + "-ca"
		if err := d.sdkClient.UploadCertificate(ctx, chainName, partition, chainCertPEM); err != nil {
			return nil, fmt.Errorf("failed to upload chain certificate: %w", err)
		}
		d.logger.Info("chain certificate uploaded", slog.String("chainName", chainName))
		chainPath = fmt.Sprintf("/%s/%s", partition, chainName)
	}

	switch d.config.DeployTarget {
	case "", DEPLOY_TARGET_CERTIFICATE:
		d.logger.Info("deployment completed (certificate only)")

	case DEPLOY_TARGET_CLIENTSSL:
		if d.config.ClientSSLProfileName == "" {
			return nil, fmt.Errorf("config `clientSSLProfileName` is required when deploy target is 'clientssl'")
		}

		certPath := fmt.Sprintf("/%s/%s", partition, certName)
		keyPath := fmt.Sprintf("/%s/%s", partition, certName)

		_, err := d.sdkClient.GetClientSSLProfile(ctx, d.config.ClientSSLProfileName, partition)
		if err != nil {
			if errors.Is(err, f5sdk.ErrNotFound) {
				d.logger.Info("client-ssl profile not found, creating it", slog.String("profile", d.config.ClientSSLProfileName))
				if err := d.sdkClient.CreateClientSSLProfile(ctx, d.config.ClientSSLProfileName, partition, certPath, keyPath, chainPath); err != nil {
					return nil, fmt.Errorf("failed to create client-ssl profile: %w", err)
				}
				d.logger.Info("client-ssl profile created", slog.String("profile", d.config.ClientSSLProfileName))
			} else {
				return nil, fmt.Errorf("failed to get client-ssl profile: %w", err)
			}
		} else {
			if err := d.sdkClient.UpdateClientSSLProfile(ctx, d.config.ClientSSLProfileName, partition, certPath, keyPath, chainPath); err != nil {
				return nil, fmt.Errorf("failed to update client-ssl profile: %w", err)
			}
			d.logger.Info("client-ssl profile updated", slog.String("profile", d.config.ClientSSLProfileName))
		}

	default:
		return nil, fmt.Errorf("unsupported deploy target '%s'", d.config.DeployTarget)
	}

	return &DeployResult{}, nil
}

func buildCertName(cert *x509.Certificate) string {
	san := ""
	if len(cert.DNSNames) > 0 {
		san = cert.DNSNames[0]
	} else if cert.Subject.CommonName != "" {
		san = cert.Subject.CommonName
	} else {
		return ""
	}

	san = strings.TrimPrefix(san, "*.")
	return "certimate_" + san
}

func createSDKClient(serverUrl string, skipTlsVerify bool) (*f5sdk.Client, error) {
	client, err := f5sdk.NewClient(serverUrl)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}
