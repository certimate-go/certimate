package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	githubplugin "github.com/hashicorp/go-plugin"

	"github.com/certimate-go/certimate/pkg/plugin"
)

type fixtureConfig struct {
	OmitMetadataAccessType bool            `json:"omitMetadataAccessType"`
	OmitAccessSchema       bool            `json:"omitAccessSchema"`
	ProviderType           string          `json:"providerType"`
	AccessProviderType     string          `json:"accessProviderType"`
	SchemaRevision         string          `json:"schemaRevision"`
	AccessField            string          `json:"accessField"`
	DeployField            string          `json:"deployField"`
	BootstrapError         string          `json:"bootstrapError"`
	MetadataDelayMS        int             `json:"metadataDelayMs"`
	MetadataURL            string          `json:"metadataUrl"`
	StartupLog             string          `json:"startupLog"`
	AccessSchema           json.RawMessage `json:"accessSchema"`
	DeploySchema           json.RawMessage `json:"deploySchema"`
}

type fixtureDeployer struct {
	config fixtureConfig
}

func (d *fixtureDeployer) GetMetadata(ctx context.Context) (*plugin.Metadata, error) {
	if d.config.MetadataURL != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.config.MetadataURL, nil)
		if err != nil {
			return nil, err
		}
		response, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
		if err != nil {
			return nil, err
		}
		response.Body.Close()
	}
	if d.config.MetadataDelayMS > 0 {
		select {
		case <-time.After(time.Duration(d.config.MetadataDelayMS) * time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if d.config.BootstrapError != "" {
		return nil, fmt.Errorf("%s", d.config.BootstrapError)
	}
	accessType := d.config.AccessProviderType
	if d.config.OmitMetadataAccessType {
		accessType = ""
	}
	return &plugin.Metadata{
		ProviderType:         d.config.ProviderType,
		AccessProviderType:   accessType,
		ProtocolVersion:      plugin.ProtocolVersion,
		DeployCategory:       "other",
		DeployDisplayNameKey: "plugin.e2e.name",
		AccessDisplayNameKey: "plugin.e2e.access",
	}, nil
}

func (d *fixtureDeployer) GetConfigSchema(context.Context) (*plugin.ConfigSchema, error) {
	access := d.config.AccessSchema
	if len(access) == 0 {
		access = schema(d.config.AccessProviderType, "access", []map[string]any{
			{"name": "url", "valueType": "text", "labelKey": "plugin.e2e.url", "required": true},
			{"name": d.config.AccessField, "valueType": "text", "labelKey": "plugin.e2e.token", "secret": true},
		})
	}
	if d.config.OmitAccessSchema {
		access = nil
	}
	deploy := d.config.DeploySchema
	if len(deploy) == 0 {
		deploy = schema(d.config.ProviderType, "deploy", []map[string]any{
			{"name": "count", "valueType": "number", "labelKey": "plugin.e2e.count", "default": 1},
			{"name": d.config.DeployField, "valueType": "text", "labelKey": "plugin.e2e.message"},
		})
	}
	return &plugin.ConfigSchema{
		AccessSchemaJSON: access,
		DeploySchemaJSON: deploy,
		I18n: map[string]map[string]string{
			"en": {"plugin.e2e.name": "E2E Webhook", "plugin.e2e.access": "E2E Access", "plugin.e2e.url": "Webhook URL", "plugin.e2e.token": "Token " + d.config.SchemaRevision, "plugin.e2e.count": "Count", "plugin.e2e.message": "Message " + d.config.SchemaRevision},
			"zh": {"plugin.e2e.name": "E2E Webhook", "plugin.e2e.access": "E2E Access", "plugin.e2e.url": "Webhook URL", "plugin.e2e.token": "Token " + d.config.SchemaRevision, "plugin.e2e.count": "Count", "plugin.e2e.message": "Message " + d.config.SchemaRevision},
		},
	}, nil
}

func schema(provider, category string, columns []map[string]any) []byte {
	data, err := json.Marshal(map[string]any{"schemaVersion": "form/v1", "provider": provider, "category": category, "schema": map[string]any{"columns": columns}})
	if err != nil {
		panic(err)
	}
	return data
}

func (d *fixtureDeployer) Deploy(ctx context.Context, req *plugin.DeployRequest, logger *slog.Logger) (*plugin.DeployResult, error) {
	var access struct {
		URL     string `json:"url"`
		Method  string `json:"method"`
		Headers string `json:"headers"`
		Token   string `json:"token"`
	}
	if err := json.Unmarshal([]byte(req.AccessConfigJSON), &access); err != nil {
		return nil, err
	}
	var extended struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal([]byte(req.ExtendedConfigJSON), &extended); err != nil {
		return nil, err
	}
	data, err := json.Marshal(map[string]any{
		"cert": req.CertificatePEM, "key": req.PrivateKeyPEM,
		"access": json.RawMessage(req.AccessConfigJSON), "config": json.RawMessage(req.ExtendedConfigJSON),
		"version": d.config.SchemaRevision,
	})
	if err != nil {
		return nil, err
	}
	method := access.Method
	if method == "" {
		method = http.MethodPost
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, access.URL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if access.Token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+access.Token)
	}
	for _, line := range strings.Split(access.Headers, "\n") {
		if key, value, ok := strings.Cut(line, ":"); ok {
			httpReq.Header.Set(strings.TrimSpace(key), strings.TrimSpace(value))
		}
	}
	response, err := (&http.Client{Timeout: 10 * time.Second}).Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("webhook status %d", response.StatusCode)
	}
	logger.Info("webhook responded")
	return &plugin.DeployResult{ExtendedDataJSON: `{"deployed":true}`}, nil
}

func main() {
	data, err := os.ReadFile(filepath.Join(filepath.Dir(os.Args[0]), "fixture.json"))
	if err != nil {
		panic(err)
	}
	var config fixtureConfig
	if err := json.Unmarshal(data, &config); err != nil {
		panic(err)
	}
	if config.AccessProviderType == "" {
		config.AccessProviderType = config.ProviderType
	}
	if config.AccessField == "" {
		config.AccessField = "token"
	}
	if config.DeployField == "" {
		config.DeployField = "message"
	}
	if config.StartupLog != "" {
		f, err := os.OpenFile(config.StartupLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			panic(err)
		}
		if _, err := fmt.Fprintln(f, config.SchemaRevision); err != nil {
			panic(err)
		}
		if err := f.Close(); err != nil {
			panic(err)
		}
	}
	githubplugin.Serve(&githubplugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig,
		Plugins: map[string]githubplugin.Plugin{
			plugin.PluginName: &plugin.DeployerGRPCPlugin{Impl: &fixtureDeployer{config: config}},
		},
		GRPCServer: githubplugin.DefaultGRPCServer,
	})
}
