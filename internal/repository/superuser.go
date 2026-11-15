package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"

	"ganigeorgiev.com/pocketbase/dbx"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

var ErrAlreadyInitialized = errors.New("administrator already configured")

type SuperuserRepository struct{ app core.App }

func NewSuperuserRepository(app core.App) *SuperuserRepository {
	return &SuperuserRepository{app: app}
}

func (r *SuperuserRepository) Initialized() (bool, error) {
	count, err := r.app.CountRecords(core.CollectionNameSuperusers, dbx.Not(dbx.HashExp{"email": core.DefaultInstallerEmail}))
	return count > 0, err
}

func (r *SuperuserRepository) PrepareInstaller() error {
	return r.app.RunInTransaction(func(txApp core.App) error {
		initialized, err := NewSuperuserRepository(txApp).Initialized()
		if err != nil {
			return err
		}
		installer, err := txApp.FindAuthRecordByEmail(core.CollectionNameSuperusers, core.DefaultInstallerEmail)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if initialized {
			return txApp.Delete(installer)
		}
		installer.SetPassword(rand.Text())
		installer.RefreshTokenKey()
		return txApp.Save(installer)
	})
}

func (r *SuperuserRepository) CreateFirst(ctx context.Context, email, password string, authorize func() error) error {
	return r.app.RunInTransaction(func(txApp core.App) error {
		initialized, err := NewSuperuserRepository(txApp).Initialized()
		if err != nil {
			return err
		}
		if initialized {
			return ErrAlreadyInitialized
		}
		if authorize != nil {
			if err := authorize(); err != nil {
				return err
			}
		}
		collection, err := txApp.FindCollectionByNameOrId(core.CollectionNameSuperusers)
		if err != nil {
			return err
		}
		record := core.NewRecord(collection)
		record.SetEmail(email)
		record.SetPassword(password)
		if err := txApp.SaveWithContext(ctx, record); err != nil {
			return err
		}
		installer, err := txApp.FindAuthRecordByEmail(collection, core.DefaultInstallerEmail)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		return txApp.DeleteWithContext(ctx, installer)
	})
}
