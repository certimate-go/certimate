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

type WorkflowLogRepository struct{}

func NewWorkflowLogRepository() *WorkflowLogRepository {
	return &WorkflowLogRepository{}
}

func (r *WorkflowLogRepository) ListByWorkflowRunId(ctx context.Context, workflowRunId string) ([]*domain.WorkflowLog, error) {
	records, err := app.GetApp().FindRecordsByFilter(
		domain.CollectionNameWorkflowLog,
		"runRef={:runId}",
		"timestamp",
		0, 0,
		dbx.Params{"runId": workflowRunId},
	)
	if err != nil {
		return nil, err
	}

	workflowLogs := make([]*domain.WorkflowLog, 0)
	for _, record := range records {
		workflowLog, err := r.castRecordToModel(record)
		if err != nil {
			return nil, err
		}

		workflowLogs = append(workflowLogs, workflowLog)
	}

	return workflowLogs, nil
}

func (r *WorkflowLogRepository) Save(ctx context.Context, workflowLog *domain.WorkflowLog) (*domain.WorkflowLog, error) {
	collection, err := app.GetApp().FindCollectionByNameOrId(domain.CollectionNameWorkflowLog)
	if err != nil {
		return workflowLog, err
	}

	var record *core.Record
	if workflowLog.Id == "" {
		record = core.NewRecord(collection)
	} else {
		record, err = app.GetApp().FindRecordById(collection, workflowLog.Id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return workflowLog, err
			}
			record = core.NewRecord(collection)
		}
	}

	record.Set("workflowRef", workflowLog.WorkflowId)
	record.Set("runRef", workflowLog.RunId)
	record.Set("nodeId", workflowLog.NodeId)
	record.Set("nodeName", workflowLog.NodeName)
	record.Set("timestamp", workflowLog.TimestampMilli)
	record.Set("level", workflowLog.Level)
	record.Set("message", workflowLog.Message)
	record.Set("data", workflowLog.Data)
	record.Set("created", workflowLog.CreatedAt)
	err = app.GetApp().Save(record)
	if err != nil {
		return workflowLog, err
	}

	workflowLog.Id = record.Id
	workflowLog.CreatedAt = record.GetDateTime("created").Time()
	workflowLog.UpdatedAt = record.GetDateTime("updated").Time()

	return workflowLog, nil
}

func (r *WorkflowLogRepository) castRecordToModel(record *core.Record) (*domain.WorkflowLog, error) {
	if record == nil {
		return nil, fmt.Errorf("the record is nil")
	}

	logdata := make(map[string]any)
	if err := record.UnmarshalJSONField("data", &logdata); err != nil {
		return nil, fmt.Errorf("field 'data' is malformed")
	}

	workflowLog := &domain.WorkflowLog{
		Meta: domain.Meta{
			Id:        record.Id,
			CreatedAt: record.GetDateTime("created").Time(),
			UpdatedAt: record.GetDateTime("updated").Time(),
		},
		WorkflowId:     record.GetString("workflowRef"),
		RunId:          record.GetString("runRef"),
		NodeId:         record.GetString("nodeId"),
		NodeName:       record.GetString("nodeName"),
		TimestampMilli: int64(record.GetInt("timestamp")),
		Level:          int32(record.GetInt("level")),
		Message:        record.GetString("message"),
		Data:           logdata,
	}
	return workflowLog, nil
}
