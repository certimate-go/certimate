package domain

import "encoding/json"

const CollectionNameSettings = "settings"

type Settings struct {
	Meta
	Name    string `json:"name" db:"name"`
	Content string `json:"content" db:"content"`
}

type SettingsContentAsPersistence struct {
	WorkflowRunsMaxDaysRetention        int `json:"workflowRunsMaxDaysRetention"`
	ExpiredCertificatesMaxDaysRetention int `json:"expiredCertificatesMaxDaysRetention"`
}

func (s *Settings) UnmarshalContentAsPersistence() (*SettingsContentAsPersistence, error) {
	var content *SettingsContentAsPersistence
	if err := json.Unmarshal([]byte(s.Content), &content); err != nil {
		return nil, err
	}
	return content, nil
}
