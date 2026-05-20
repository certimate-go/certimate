package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type WorkflowOutputRepository struct{}

func NewWorkflowOutputRepository() *WorkflowOutputRepository {
	return &WorkflowOutputRepository{}
}

func (r *WorkflowOutputRepository) GetByWorkflowIdAndNodeId(ctx context.Context, workflowId string, workflowNodeId string) (*domain.WorkflowOutput, error) {
	records, err := app.GetApp().FindRecordsByFilter(
		domain.CollectionNameWorkflowOutput,
		"workflowRef={:workflowId} && nodeId={:nodeId}",
		"-created",
		1, 0,
		dbx.Params{"workflowId": workflowId},
		dbx.Params{"nodeId": workflowNodeId},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}
	if len(records) == 0 {
		return nil, domain.ErrRecordNotFound
	}

	return r.castRecordToModel(records[0])
}

func (r *WorkflowOutputRepository) GetByWorkflowRunIdAndNodeId(ctx context.Context, workflowRunId string, workflowNodeId string) (*domain.WorkflowOutput, error) {
	records, err := app.GetApp().FindRecordsByFilter(
		domain.CollectionNameWorkflowOutput,
		"runRef={:workflowRunId} && nodeId={:nodeId}",
		"-created",
		1, 0,
		dbx.Params{"workflowRunId": workflowRunId},
		dbx.Params{"nodeId": workflowNodeId},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}
	if len(records) == 0 {
		return nil, domain.ErrRecordNotFound
	}

	return r.castRecordToModel(records[0])
}

func (r *WorkflowOutputRepository) Save(ctx context.Context, workflowOutput *domain.WorkflowOutput) (*domain.WorkflowOutput, error) {
	record, err := r.saveRecord(workflowOutput)
	if err != nil {
		return workflowOutput, err
	}

	workflowOutput.Id = record.Id
	workflowOutput.CreatedAt = record.GetDateTime("created").Time()
	workflowOutput.UpdatedAt = record.GetDateTime("updated").Time()
	return workflowOutput, nil
}

func (r *WorkflowOutputRepository) castRecordToModel(record *core.Record) (*domain.WorkflowOutput, error) {
	if record == nil {
		return nil, fmt.Errorf("the record is nil")
	}

	nodeConfig := make(domain.WorkflowNodeConfig)
	if err := record.UnmarshalJSONField("nodeConfig", &nodeConfig); err != nil {
		return nil, fmt.Errorf("field 'nodeConfig' is malformed")
	}

	outputs := make([]*domain.WorkflowOutputEntry, 0)
	if err := record.UnmarshalJSONField("outputs", &outputs); err != nil {
		return nil, fmt.Errorf("field 'outputs' is malformed")
	}

	workflowOutput := &domain.WorkflowOutput{
		Meta: domain.Meta{
			Id:        record.Id,
			CreatedAt: record.GetDateTime("created").Time(),
			UpdatedAt: record.GetDateTime("updated").Time(),
		},
		WorkflowId: record.GetString("workflowRef"),
		RunId:      record.GetString("runRef"),
		NodeId:     record.GetString("nodeId"),
		NodeConfig: nodeConfig,
		Outputs:    outputs,
		Succeeded:  record.GetBool("succeeded"),
	}
	return workflowOutput, nil
}

func (r *WorkflowOutputRepository) saveRecord(workflowOutput *domain.WorkflowOutput) (*core.Record, error) {
	collection, err := app.GetApp().FindCollectionByNameOrId(domain.CollectionNameWorkflowOutput)
	if err != nil {
		return nil, err
	}

	var record *core.Record
	if workflowOutput.Id == "" {
		record = core.NewRecord(collection)
	} else {
		record, err = app.GetApp().FindRecordById(collection, workflowOutput.Id)
		if err != nil {
			return record, err
		}
	}
	record.Set("workflowRef", workflowOutput.WorkflowId)
	record.Set("runRef", workflowOutput.RunId)
	record.Set("nodeId", workflowOutput.NodeId)
	record.Set("nodeConfig", workflowOutput.NodeConfig)
	record.Set("outputs", workflowOutput.Outputs)
	record.Set("succeeded", workflowOutput.Succeeded)
	if err := app.GetApp().Save(record); err != nil {
		return record, err
	}

	return record, nil
}
