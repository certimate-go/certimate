package gcp

import (
	"encoding/json"
	"fmt"
)

func GetProjectIDFromServiceAccountKey(serviceAccountKey string) (string, error) {
	saKey := []byte(serviceAccountKey)

	var saKeyJSON struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(saKey, &saKeyJSON); err != nil || saKeyJSON.ProjectID == "" {
		return "", fmt.Errorf("invalid service account key")
	}

	return saKeyJSON.ProjectID, nil
}
