package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"ganigeorgiev.com/pocketbase/dbx"
	"ganigeorgiev.com/pocketbase/pocketbase/core"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
)

type WorkflowTemplateRepository struct {
	pb core.App
}

func NewWorkflowTemplateRepository() *WorkflowTemplateRepository {
	return &WorkflowTemplateRepository{pb: app.GetApp()}
}

func (r *WorkflowTemplateRepository) GetById(ctx context.Context, id string) (*domain.WorkflowTemplate, error) {
	record, err := r.pb.FindRecordById(domain.CollectionNameWorkflowTemplate, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	return r.castRecordToModel(record)
}

func (r *WorkflowTemplateRepository) List(ctx context.Context, keyword string) ([]*domain.WorkflowTemplate, error) {
	filter := ""
	params := dbx.Params{}
	if keyword != "" {
		filter = "(id={:keyword} || name~{:keyword})"
		params["keyword"] = keyword
	}

	records, err := r.pb.FindRecordsByFilter(
		domain.CollectionNameWorkflowTemplate,
		filter,
		"-created",
		0, 0,
		params,
	)
	if err != nil {
		return nil, err
	}

	templates := make([]*domain.WorkflowTemplate, 0)
	for _, record := range records {
		template, err := r.castRecordToModel(record)
		if err != nil {
			return nil, err
		}

		templates = append(templates, template)
	}

	return templates, nil
}

func (r *WorkflowTemplateRepository) Save(ctx context.Context, template *domain.WorkflowTemplate) (*domain.WorkflowTemplate, error) {
	collection, err := r.pb.FindCollectionByNameOrId(domain.CollectionNameWorkflowTemplate)
	if err != nil {
		return template, err
	}

	var record *core.Record
	if template.Id == "" {
		record = core.NewRecord(collection)
	} else {
		record, err = r.pb.FindRecordById(collection, template.Id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return template, domain.ErrRecordNotFound
			}
			return template, err
		}
	}

	record.Set("name", template.Name)
	record.Set("description", template.Description)
	record.Set("formatVersion", template.FormatVersion)
	record.Set("revision", template.Revision)
	record.Set("graph", template.Graph)
	record.Set("params", template.Params)
	if err := r.pb.SaveWithContext(ctx, record); err != nil {
		return template, err
	}

	template.Id = record.Id
	template.CreatedAt = record.GetDateTime("created").Time()
	template.UpdatedAt = record.GetDateTime("updated").Time()
	return template, nil
}

func (r *WorkflowTemplateRepository) Delete(ctx context.Context, id string) error {
	record, err := r.pb.FindRecordById(domain.CollectionNameWorkflowTemplate, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrRecordNotFound
		}
		return err
	}

	return r.pb.Delete(record)
}

func (r *WorkflowTemplateRepository) castRecordToModel(record *core.Record) (*domain.WorkflowTemplate, error) {
	if record == nil {
		return nil, fmt.Errorf("the record is nil")
	}

	graph := &domain.WorkflowGraph{}
	if err := record.UnmarshalJSONField("graph", graph); err != nil {
		return nil, fmt.Errorf("field 'graph' is malformed")
	}

	params := make(domain.WorkflowTemplateParams, 0)
	if err := record.UnmarshalJSONField("params", &params); err != nil {
		return nil, fmt.Errorf("field 'params' is malformed")
	}

	template := &domain.WorkflowTemplate{
		Meta: domain.Meta{
			Id:        record.Id,
			CreatedAt: record.GetDateTime("created").Time(),
			UpdatedAt: record.GetDateTime("updated").Time(),
		},
		Name:          record.GetString("name"),
		Description:   record.GetString("description"),
		FormatVersion: record.GetInt("formatVersion"),
		Revision:      record.GetInt("revision"),
		Graph:         graph,
		Params:        params,
	}
	return template, nil
}
