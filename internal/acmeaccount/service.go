package acmeaccount

import (
	"context"
	"fmt"

	"github.com/certimate-go/certimate/internal/certacme"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/domain/dtos"
)

type acmeAccountRepository interface {
	List(ctx context.Context, page, perPage int, ca string) ([]*domain.ACMEAccount, int64, error)
	GetById(ctx context.Context, id string) (*domain.ACMEAccount, error)
}

type Service struct {
	repo acmeAccountRepository
}

func NewService(repo acmeAccountRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, req *dtos.ACMEAccountListReq) (*dtos.ACMEAccountListResp, error) {
	page := 1
	perPage := 15
	ca := ""
	if req != nil {
		page = req.Page
		perPage = req.PerPage
		ca = req.CA
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 15
	}
	if perPage > 100 {
		perPage = 100
	}

	accounts, total, err := s.repo.List(ctx, page, perPage, ca)
	if err != nil {
		return nil, fmt.Errorf("failed to list acme accounts: %w", err)
	}

	items := make([]*dtos.ACMEAccountView, 0, len(accounts))
	for _, account := range accounts {
		items = append(items, toView(account))
	}
	return &dtos.ACMEAccountListResp{
		Items:      items,
		TotalItems: total,
		Page:       page,
		PerPage:    perPage,
	}, nil
}

func (s *Service) Import(ctx context.Context, req *dtos.ACMEAccountImportReq) (*dtos.ACMEAccountImportResp, error) {
	if req == nil {
		return nil, domain.NewError(400, "request body is empty")
	}

	account, err := certacme.ImportACMEAccount(ctx, &certacme.ImportACMEAccountOptions{
		CA:            req.CA,
		ACMEDirUrl:    req.ACMEDirUrl,
		PrivateKeyPem: req.PrivateKeyPem,
		Email:         req.Email,
	})
	if err != nil {
		return nil, err
	}

	return &dtos.ACMEAccountImportResp{Item: toView(account)}, nil
}

func (s *Service) Export(ctx context.Context, accountId string) (*dtos.ACMEAccountExportResp, error) {
	if accountId == "" {
		return nil, domain.NewError(400, "the account id is empty")
	}

	account, err := s.repo.GetById(ctx, accountId)
	if err != nil {
		if domain.IsRecordNotFoundError(err) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, fmt.Errorf("failed to get acme account: %w", err)
	}
	if account.PrivateKey == "" {
		return nil, domain.NewError(404, "private key is empty")
	}

	return &dtos.ACMEAccountExportResp{PrivateKeyPem: account.PrivateKey}, nil
}

func (s *Service) Rotate(ctx context.Context, accountId string) (*dtos.ACMEAccountRotateResp, error) {
	account, err := certacme.RotateACMEAccountKey(ctx, accountId)
	if err != nil {
		return nil, err
	}
	return &dtos.ACMEAccountRotateResp{Item: toView(account)}, nil
}

func toView(account *domain.ACMEAccount) *dtos.ACMEAccountView {
	if account == nil {
		return nil
	}
	return &dtos.ACMEAccountView{
		Id:               account.Id,
		CA:               account.CA,
		Email:            account.Email,
		ACMEDirectoryUrl: account.ACMEDirectoryUrl,
		ACMEAccountUrl:   account.ACMEAccountUrl,
		CreatedAt:        account.CreatedAt,
		UpdatedAt:        account.UpdatedAt,
	}
}
