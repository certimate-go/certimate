package certificate

import (
	"context"

	"github.com/pocketbase/dbx"

	"github.com/certimate-go/certimate/internal/domain"
)

type acmeAccountRepository interface {
	GetByAcctUrl(ctx context.Context, acctUrl string) (*domain.ACMEAccount, error)
}

type certificateRepository interface {
	GetById(ctx context.Context, id string) (*domain.Certificate, error)
	Save(ctx context.Context, certificate *domain.Certificate) (*domain.Certificate, error)
	DeleteWithExprs(ctx context.Context, exprs ...dbx.Expression) (int, error)
}

type settingsRepository interface {
	GetByName(ctx context.Context, name string) (*domain.Settings, error)
}
