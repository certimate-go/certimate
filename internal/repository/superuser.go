package repository

import (
	"context"

	"ganigeorgiev.com/pocketbase/dbx"
	"ganigeorgiev.com/pocketbase/pocketbase/core"

	"github.com/certimate-go/certimate/internal/app"
)

type SuperuserRepository struct {
	pb core.App
}

func NewSuperuserRepository() *SuperuserRepository {
	return &SuperuserRepository{pb: app.GetApp()}
}

func (r *SuperuserRepository) IsEmpty(_ context.Context) (bool, error) {
	total, err := r.pb.CountRecords(core.CollectionNameSuperusers, dbx.Not(
		dbx.HashExp{"email": core.DefaultInstallerEmail},
	))
	return total == 0, err
}

func (r *SuperuserRepository) Create(ctx context.Context, email, password string) error {
	collection, err := r.pb.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		return err
	}

	record := core.NewRecord(collection)
	record.SetEmail(email)
	record.SetPassword(password)
	return r.pb.SaveWithContext(ctx, record)
}
