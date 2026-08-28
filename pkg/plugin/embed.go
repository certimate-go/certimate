package plugin

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
)

func LoadDeploySchema(fsys embed.FS) ([]byte, error) {
	data, err := fsys.ReadFile("schema/deploy.json")
	if err != nil {
		return nil, fmt.Errorf("plugin: embedded deploy schema not found: %w", err)
	}
	return data, nil
}

func LoadAccessSchema(fsys embed.FS) ([]byte, error) {
	data, err := fsys.ReadFile("schema/access.json")
	if err != nil {
		return nil, nil
	}
	return data, nil
}

func LoadI18n(fsys embed.FS) (map[string]map[string]string, error) {
	entries, err := fs.ReadDir(fsys, "schema/i18n")
	if err != nil {
		return nil, fmt.Errorf("plugin: embedded i18n directory not found: %w", err)
	}
	result := make(map[string]map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		locale := strings.TrimSuffix(entry.Name(), ".json")
		data, err := fsys.ReadFile("schema/i18n/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("plugin: cannot read i18n file %s: %w", entry.Name(), err)
		}
		var kv map[string]string
		if err := json.Unmarshal(data, &kv); err != nil {
			return nil, fmt.Errorf("plugin: invalid i18n JSON in %s: %w", entry.Name(), err)
		}
		result[locale] = kv
	}
	return result, nil
}
