package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type CertificateRepository struct{}

func NewCertificateRepository() *CertificateRepository {
	return &CertificateRepository{}
}

func (r *CertificateRepository) ListExpiringSoon(ctx context.Context) ([]*domain.Certificate, error) {
	records, err := app.GetApp().FindAllRecords(
		domain.CollectionNameCertificate,
		dbx.NewExp("validityNotAfter>DATETIME('now')"),
		dbx.NewExp("validityNotAfter<DATETIME('now', '+20 days')"),
		dbx.NewExp("deleted=null"),
	)
	if err != nil {
		return nil, err
	}

	certificates := make([]*domain.Certificate, 0)
	for _, record := range records {
		certificate, err := r.castRecordToModel(record)
		if err != nil {
			return nil, err
		}

		certificates = append(certificates, certificate)
	}

	return certificates, nil
}

func (r *CertificateRepository) GetById(ctx context.Context, id string) (*domain.Certificate, error) {
	record, err := app.GetApp().FindRecordById(domain.CollectionNameCertificate, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	if !record.GetDateTime("deleted").Time().IsZero() {
		return nil, domain.ErrRecordNotFound
	}

	return r.castRecordToModel(record)
}

func (r *CertificateRepository) GetByWorkflowIdAndNodeId(ctx context.Context, workflowId string, workflowNodeId string) (*domain.Certificate, error) {
	records, err := app.GetApp().FindRecordsByFilter(
		domain.CollectionNameCertificate,
		"workflowRef={:workflowId} && workflowNodeId={:workflowNodeId} && deleted=null",
		"-created",
		1, 0,
		dbx.Params{"workflowId": workflowId},
		dbx.Params{"workflowNodeId": workflowNodeId},
	)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, domain.ErrRecordNotFound
	}

	return r.castRecordToModel(records[0])
}

func (r *CertificateRepository) GetByWorkflowRunIdAndNodeId(ctx context.Context, workflowRunId string, workflowNodeId string) (*domain.Certificate, error) {
	records, err := app.GetApp().FindRecordsByFilter(
		domain.CollectionNameCertificate,
		"workflowRunRef={:workflowRunId} && workflowNodeId={:workflowNodeId} && deleted=null",
		"-created",
		1, 0,
		dbx.Params{"workflowRunId": workflowRunId},
		dbx.Params{"workflowNodeId": workflowNodeId},
	)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, domain.ErrRecordNotFound
	}

	return r.castRecordToModel(records[0])
}

func (r *CertificateRepository) Save(ctx context.Context, certificate *domain.Certificate) (*domain.Certificate, error) {
	collection, err := app.GetApp().FindCollectionByNameOrId(domain.CollectionNameCertificate)
	if err != nil {
		return certificate, err
	}

	var record *core.Record
	if certificate.Id == "" {
		record = core.NewRecord(collection)
	} else {
		record, err = app.GetApp().FindRecordById(collection, certificate.Id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return certificate, domain.ErrRecordNotFound
			}
			return certificate, err
		}
	}

	record.Set("source", string(certificate.Source))
	record.Set("subjectAltNames", certificate.SubjectAltNames)
	record.Set("serialNumber", certificate.SerialNumber)
	record.Set("certificate", certificate.Certificate)
	record.Set("privateKey", certificate.PrivateKey)
	record.Set("issuerOrg", certificate.IssuerOrg)
	record.Set("issuerCertificate", certificate.IssuerCertificate)
	record.Set("keyAlgorithm", string(certificate.KeyAlgorithm))
	record.Set("validityNotBefore", certificate.ValidityNotBefore)
	record.Set("validityNotAfter", certificate.ValidityNotAfter)
	record.Set("validityInterval", certificate.ValidityInterval)
	record.Set("acmeAcctUrl", certificate.ACMEAcctUrl)
	record.Set("acmeCertUrl", certificate.ACMECertUrl)
	record.Set("isRenewed", certificate.IsRenewed)
	record.Set("isRevoked", certificate.IsRevoked)
	record.Set("workflowRef", certificate.WorkflowId)
	record.Set("workflowRunRef", certificate.WorkflowRunId)
	record.Set("workflowNodeId", certificate.WorkflowNodeId)
	if err := app.GetApp().Save(record); err != nil {
		return certificate, err
	}

	certificate.Id = record.Id
	certificate.CreatedAt = record.GetDateTime("created").Time()
	certificate.UpdatedAt = record.GetDateTime("updated").Time()
	return certificate, nil
}

func (r *CertificateRepository) DeleteWhere(ctx context.Context, exprs ...dbx.Expression) (int, error) {
	records, err := app.GetApp().FindAllRecords(domain.CollectionNameCertificate, exprs...)
	if err != nil {
		return 0, nil
	}

	var ret int
	var errs []error
	for _, record := range records {
		if err := app.GetApp().Delete(record); err != nil {
			errs = append(errs, err)
		} else {
			ret++
		}
	}

	if len(errs) > 0 {
		return ret, errors.Join(errs...)
	}

	return ret, nil
}

func (r *CertificateRepository) castRecordToModel(record *core.Record) (*domain.Certificate, error) {
	if record == nil {
		return nil, errors.New("the record is nil")
	}

	certificate := &domain.Certificate{
		Meta: domain.Meta{
			Id:        record.Id,
			CreatedAt: record.GetDateTime("created").Time(),
			UpdatedAt: record.GetDateTime("updated").Time(),
		},
		Source:            domain.CertificateSourceType(record.GetString("source")),
		SubjectAltNames:   record.GetString("subjectAltNames"),
		SerialNumber:      record.GetString("serialNumber"),
		Certificate:       record.GetString("certificate"),
		PrivateKey:        record.GetString("privateKey"),
		IssuerOrg:         record.GetString("issuerOrg"),
		IssuerCertificate: record.GetString("issuerCertificate"),
		KeyAlgorithm:      domain.CertificateKeyAlgorithmType(record.GetString("keyAlgorithm")),
		ValidityNotBefore: record.GetDateTime("validityNotBefore").Time(),
		ValidityNotAfter:  record.GetDateTime("validityNotAfter").Time(),
		ValidityInterval:  int32(record.GetInt("validityInterval")),
		ACMEAcctUrl:       record.GetString("acmeAcctUrl"),
		ACMECertUrl:       record.GetString("acmeCertUrl"),
		IsRenewed:         record.GetBool("isRenewed"),
		IsRevoked:         record.GetBool("isRevoked"),
		WorkflowId:        record.GetString("workflowRef"),
		WorkflowRunId:     record.GetString("workflowRunRef"),
		WorkflowNodeId:    record.GetString("workflowNodeId"),
	}
	return certificate, nil
}
