package synologydsm

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	synology "github.com/certimate-go/certimate/pkg/sdk3rd/synology"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type DeployerConfig struct {
	// Server hostname or IP address.
	Hostname string `json:"hostname"`
	// Server port.
	// Default is 5000 for HTTP, 5001 for HTTPS.
	Port int32 `json:"port,omitempty"`
	// Connection scheme (http or https).
	// Default is "http".
	Scheme string `json:"scheme,omitempty"`
	// Synology DSM admin username.
	Username string `json:"username"`
	// Synology DSM admin password.
	Password string `json:"password"`
	// 2FA TOTP secret key (optional).
	// If provided, OTP code will be auto-generated at login time.
	TotpSecret string `json:"totpSecret,omitempty"`
	// Allow insecure HTTPS connections (skip TLS verification).
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// Certificate ID to update (optional).
	// If not provided and CertificateName is not provided, a new certificate will be created.
	CertificateID string `json:"certificateId,omitempty"`
	// Certificate description/name to find and update (optional).
	// If provided, will search for existing certificate with this description.
	CertificateName string `json:"certificateName,omitempty"`
	// Set as default certificate.
	IsDefault bool `json:"isDefault,omitempty"`
}

type Deployer struct {
	config *DeployerConfig
	logger *slog.Logger
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	if config.Hostname == "" {
		return nil, fmt.Errorf("hostname is required")
	}

	if config.Username == "" {
		return nil, fmt.Errorf("username is required")
	}

	if config.Password == "" {
		return nil, fmt.Errorf("password is required")
	}

	return &Deployer{
		config: config,
		logger: slog.Default(),
	}, nil
}

func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.Default()
	} else {
		d.logger = logger
	}
}

func (d *Deployer) Deploy(ctx context.Context, certPEM string, privkeyPEM string) (*deployer.DeployResult, error) {
	// Create Synology client
	client := synology.NewClient(
		d.config.Hostname,
		int(d.config.Port),
		d.config.Scheme,
		d.config.AllowInsecureConnections,
	)

	// Generate OTP code from TOTP secret if provided
	var otpCode string
	if d.config.TotpSecret != "" {
		code, err := synology.GenerateTOTPCode(d.config.TotpSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to generate TOTP code: %w", err)
		}
		otpCode = code
		d.logger.Info("generated TOTP code for 2FA")
	}

	// Login
	d.logger.Info("logging in to Synology DSM", slog.String("hostname", d.config.Hostname))
	if err := client.Login(d.config.Username, d.config.Password, otpCode); err != nil {
		return nil, fmt.Errorf("failed to login to Synology DSM: %w", err)
	}
	defer func() {
		if err := client.Logout(); err != nil {
			d.logger.Warn("failed to logout from Synology DSM", slog.Any("error", err))
		}
	}()

	// Determine certificate ID
	certID := d.config.CertificateID

	// If CertificateName is provided, search for existing certificate
	if certID == "" && d.config.CertificateName != "" {
		d.logger.Info("searching for certificate by name", slog.String("name", d.config.CertificateName))
		certs, err := client.ListCertificates()
		if err != nil {
			return nil, fmt.Errorf("failed to list certificates: %w", err)
		}

		for _, cert := range certs {
			if cert.Description == d.config.CertificateName {
				certID = cert.ID
				d.logger.Info("found existing certificate", slog.String("id", certID), slog.String("desc", cert.Description))
				break
			}
		}

		if certID == "" {
			d.logger.Info("certificate not found, will create new one", slog.String("name", d.config.CertificateName))
		}
	}

	// Extract server certificate and intermediate certificate from the full chain
	// Uses the standard certimate utility function to ensure correct extraction
	serverCertPEM, intermediateCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certificates from PEM: %w", err)
	}

	// Log certificate info for debugging
	d.logger.Info("certificate data for import",
		slog.Int("serverCertLen", len(serverCertPEM)),
		slog.Int("intermediateCertLen", len(intermediateCertPEM)),
		slog.Int("keyLen", len(privkeyPEM)),
	)

	// Import certificate
	d.logger.Info("importing certificate to Synology DSM",
		slog.String("certId", certID),
		slog.String("description", d.config.CertificateName),
		slog.Bool("isDefault", d.config.IsDefault),
	)

	if err := client.ImportCertificate(
		serverCertPEM, // Only the server certificate
		privkeyPEM,
		intermediateCertPEM, // Intermediate/CA certificates
		certID,
		d.config.CertificateName,
		d.config.IsDefault,
	); err != nil {
		return nil, fmt.Errorf("failed to import certificate: %w", err)
	}

	d.logger.Info("certificate imported successfully")

	// If setting as default, also apply certificate to all services
	if d.config.IsDefault {
		d.logger.Info("applying certificate to all services")

		// Get the new certificate ID by searching for it
		certs, err := client.ListCertificates()
		if err != nil {
			d.logger.Warn("failed to list certificates after import", slog.Any("error", err))
		} else {
			var newCertID string
			for _, cert := range certs {
				if cert.IsDefault {
					newCertID = cert.ID
					break
				}
			}

			if newCertID != "" {
				d.logger.Info("found new default certificate", slog.String("id", newCertID))
				if err := client.SetCertificateForAllServices(newCertID, ""); err != nil {
					d.logger.Warn("failed to set certificate for all services", slog.Any("error", err))
				} else {
					d.logger.Info("certificate applied to all services successfully")
				}
			}
		}
	}

	return &deployer.DeployResult{}, nil
}
