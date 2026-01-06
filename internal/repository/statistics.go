package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pocketbase/dbx"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
)

type StatisticsRepository struct{}

func NewStatisticsRepository() *StatisticsRepository {
	return &StatisticsRepository{}
}

func (r *StatisticsRepository) Get(ctx context.Context) (*domain.Statistics, error) {
	statistics := &domain.Statistics{}

	// 读取设置
	var persistenceSettings *domain.SettingsContentForPersistence
	rsSettings := struct {
		Content string `db:"content"`
	}{}
	if err := app.GetDB().
		NewQuery(fmt.Sprintf("SELECT content FROM %s WHERE name = {:name}", domain.CollectionNameSettings)).
		Bind(dbx.Params{"name": domain.SettingsNamePersistence}).
		One(&rsSettings); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			persistenceSettings = (domain.SettingsContent{}).AsPersistence()
		} else {
			return nil, err
		}
	} else {
		json.Unmarshal([]byte(rsSettings.Content), &persistenceSettings)
	}

	// 统计所有证书
	rsCertTotal := struct {
		Total int `db:"total"`
	}{}
	if err := app.GetDB().
		NewQuery(fmt.Sprintf("SELECT COUNT(*) AS total FROM %s WHERE deleted = ''", domain.CollectionNameCertificate)).
		One(&rsCertTotal); err != nil {
		return nil, err
	}
	statistics.CertificateTotal = rsCertTotal.Total

	// 统计即将过期证书
	rsCertExpiringSoonTotal := struct {
		Total int `db:"total"`
	}{}
	if err := app.GetDB().
		NewQuery(fmt.Sprintf("SELECT COUNT(*) AS total FROM %s WHERE validityNotAfter <= DATETIME('now', '+%d days') AND validityNotAfter > DATETIME('now') AND isRevoked = 0 AND deleted = ''", domain.CollectionNameCertificate, persistenceSettings.CertificatesWarningDaysBeforeExpire)).
		One(&rsCertExpiringSoonTotal); err != nil {
		return nil, err
	}
	statistics.CertificateExpiringSoon = rsCertExpiringSoonTotal.Total

	// 统计已过期证书
	rsCertExpiredTotal := struct {
		Total int `db:"total"`
	}{}
	if err := app.GetDB().
		NewQuery(fmt.Sprintf("SELECT COUNT(*) AS total FROM %s WHERE validityNotAfter <= DATETIME('now') AND deleted = ''", domain.CollectionNameCertificate)).
		One(&rsCertExpiredTotal); err != nil {
		return nil, err
	}
	statistics.CertificateExpired = rsCertExpiredTotal.Total

	// 统计所有工作流
	rsWorkflowTotal := struct {
		Total int `db:"total"`
	}{}
	if err := app.GetDB().
		NewQuery(fmt.Sprintf("SELECT COUNT(*) AS total FROM %s", domain.CollectionNameWorkflow)).
		One(&rsWorkflowTotal); err != nil {
		return nil, err
	}
	statistics.WorkflowTotal = rsWorkflowTotal.Total

	// 统计已启用工作流
	rsWorkflowEnabledTotal := struct {
		Total int `db:"total"`
	}{}
	if err := app.GetDB().
		NewQuery(fmt.Sprintf("SELECT COUNT(*) AS total FROM %s WHERE enabled IS TRUE", domain.CollectionNameWorkflow)).
		One(&rsWorkflowEnabledTotal); err != nil {
		return nil, err
	}
	statistics.WorkflowEnabled = rsWorkflowEnabledTotal.Total
	statistics.WorkflowDisabled = rsWorkflowTotal.Total - rsWorkflowEnabledTotal.Total

	return statistics, nil
}
