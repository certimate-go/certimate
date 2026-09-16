package certificate

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/pocketbase/dbx"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/certacme"
	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/internal/settings"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertificateService struct {
	acmeAccountRepo acmeAccountRepository
	certificateRepo certificateRepository
}

func NewCertificateService(acmeAccountRepo acmeAccountRepository, certificateRepo certificateRepository) *CertificateService {
	return &CertificateService{
		acmeAccountRepo: acmeAccountRepo,
		certificateRepo: certificateRepo,
	}
}

func (s *CertificateService) InitSchedule(ctx context.Context) error {
	app.GetScheduler().MustAdd("cleanupCertificateExpired", "0 0 * * *", func() {
		s.cleanupExpiredCertificates(context.Background())
	})

	return nil
}

func (s *CertificateService) DownloadCertificate(ctx context.Context, req *dtos.CertificateDownloadReq) (*dtos.CertificateDownloadResp, error) {
	certificate, err := s.certificateRepo.GetById(ctx, req.CertificateId)
	if err != nil {
		return nil, err
	}

	canonicalName := strings.Split(certificate.SubjectAltNames, ";")[0]
	canonicalName = strings.ReplaceAll(canonicalName, "*", "_")

	zipBytes, err := xcert.BuildCertificateArchive(certificate.Certificate, certificate.PrivateKey, canonicalName, xcert.CertificateArchiveOptions{
		FileFormat:   string(req.FileFormat),
		PfxPassword:  req.PfxPassword,
		PfxEncoder:   req.PfxEncoder,
		JksAlias:     req.JksAlias,
		JksKeypass:   req.JksKeypass,
		JksStorepass: req.JksStorepass,
	})
	if err != nil {
		return nil, err
	}

	resp := &dtos.CertificateDownloadResp{
		ZipBytes: zipBytes,
	}
	return resp, nil
}

func (s *CertificateService) RevokeCertificate(ctx context.Context, req *dtos.CertificateRevokeReq) (*dtos.CertificateRevokeResp, error) {
	certificate, err := s.certificateRepo.GetById(ctx, req.CertificateId)
	if err != nil {
		return nil, err
	}

	if certificate.ACMEAccountUrl == "" || certificate.ACMECertificateUrl == "" {
		return nil, fmt.Errorf("could not revoke a certificate which is not issued in Certimate")
	}
	if certificate.IsRevoked {
		return nil, fmt.Errorf("could not revoke a certificate which is already revoked")
	}

	acmeAccount, err := s.acmeAccountRepo.GetByCAAndAcctUrl(ctx, certificate.CA, certificate.ACMEAccountUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke certificate: could not find acme account: %w", err)
	}

	acmeClient, err := certacme.NewACMEClientWithAccount(acmeAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke certificate: could not initialize acme config: %w", err)
	}

	revokeReq := &certacme.RevokeCertificateRequest{
		Certificate: certificate.Certificate,
	}
	_, err = acmeClient.RevokeCertificate(ctx, revokeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke certificate: %w", err)
	}

	certificate.IsRevoked = true
	certificate, err = s.certificateRepo.Save(ctx, certificate)
	if err != nil {
		return nil, err
	}

	return &dtos.CertificateRevokeResp{}, nil
}

func (s *CertificateService) cleanupExpiredCertificates(ctx context.Context) error {
	globalSettingsForPersistence := settings.GetGlobalSettingsForPersistence()
	if globalSettingsForPersistence.CertificatesRetentionMaxDays != 0 {
		ret, err := s.certificateRepo.DeleteWithExprs(ctx,
			dbx.NewExp(fmt.Sprintf("validityNotAfter<DATETIME('now', '-%d days')", globalSettingsForPersistence.CertificatesRetentionMaxDays)),
		)
		if err != nil {
			app.GetLogger().Error("failed to delete expired certificates", slog.Any("error", err))
			return err
		}

		if ret > 0 {
			app.GetLogger().Info(fmt.Sprintf("cleanup %d expired certificates", ret))
		}
	}

	return nil
}
