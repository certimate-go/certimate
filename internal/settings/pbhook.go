package settings

import (
	"context"

	"github.com/pocketbase/pocketbase/core"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
)

func registerSettingsRecordEvents() {
	pb := app.GetApp()
	pb.OnRecordCreateRequest(domain.CollectionNameSettings).BindFunc(func(e *core.RecordRequestEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		if err := onSettingsRecordCreateOrUpdate(e.Request.Context(), e.App, e.Record); err != nil {
			app.GetLogger().Error(err.Error())
			return err
		}

		return nil
	})
	pb.OnRecordUpdateRequest(domain.CollectionNameSettings).BindFunc(func(e *core.RecordRequestEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		if err := onSettingsRecordCreateOrUpdate(e.Request.Context(), e.App, e.Record); err != nil {
			app.GetLogger().Error(err.Error())
			return err
		}

		return nil
	})
	pb.OnRecordDeleteRequest(domain.CollectionNameSettings).BindFunc(func(e *core.RecordRequestEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		if err := onSettingsRecordDelete(e.Request.Context(), e.App, e.Record); err != nil {
			app.GetLogger().Error(err.Error())
			return err
		}

		return nil
	})
}

func onSettingsRecordCreateOrUpdate(_ context.Context, pb core.App, record *core.Record) error {
	sn := record.GetString("name")
	if sn != "" {
		content := make(domain.SettingsContent)
		record.UnmarshalJSONField("content", &content)

		pb.Store().Set(buildPbStoreKey(sn), content)
	}

	return nil
}

func onSettingsRecordDelete(_ context.Context, pb core.App, record *core.Record) error {
	sn := record.GetString("name")
	if sn != "" {
		pb.Store().Remove(buildPbStoreKey(sn))
	}

	return nil
}
