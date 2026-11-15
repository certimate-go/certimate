package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	Capability             string          `json:"capability"`
	SchemaRevision         string          `json:"schemaRevision"`
	AccessField            string          `json:"accessField"`
	DeployField            string          `json:"deployField"`
	BootstrapError         string          `json:"bootstrapError"`
	NotifyError            string          `json:"notifyError"`
	PresentError           string          `json:"presentError"`
	CleanUpError           string          `json:"cleanUpError"`
	PurgeError             string          `json:"purgeError"`
	MetadataDelayMS        int             `json:"metadataDelayMs"`
	MetadataURL            string          `json:"metadataUrl"`
	StartupLog             string          `json:"startupLog"`
	AccessSchema           json.RawMessage `json:"accessSchema"`
	DeploySchema           json.RawMessage `json:"deploySchema"`
}

type fixturePlugin struct {
	plugin.NoopFinalizer
	config fixtureConfig
}

func (d *fixturePlugin) metadata(ctx context.Context) (*plugin.Metadata, error) {
	config := d.config
	if config.MetadataURL != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, config.MetadataURL, nil)
		if err != nil {
			return nil, err
		}
		response, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
		if err != nil {
			return nil, err
		}
		response.Body.Close()
	}
	if config.MetadataDelayMS > 0 {
		select {
		case <-time.After(time.Duration(config.MetadataDelayMS) * time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if config.BootstrapError != "" {
		return nil, fmt.Errorf("%s", config.BootstrapError)
	}
	accessType := config.AccessProviderType
	if config.OmitMetadataAccessType {
		accessType = ""
	}
	return &plugin.Metadata{
		ProviderType:         config.ProviderType,
		AccessProviderType:   accessType,
		Capability:           config.Capability,
		ProtocolVersion:      plugin.ProtocolVersion,
		Category:             "other",
		DisplayNameKey:       "plugin.e2e.name",
		AccessDisplayNameKey: "plugin.e2e.access",
	}, nil
}

func (d *fixturePlugin) configSchema() *plugin.ConfigSchema {
	config := d.config
	access := config.AccessSchema
	if len(access) == 0 {
		access = schema(config.AccessProviderType, "access", []map[string]any{
			{"name": "url", "valueType": "text", "labelKey": "plugin.e2e.url", "required": true},
			{"name": config.AccessField, "valueType": "text", "labelKey": "plugin.e2e.token", "secret": true},
		})
	}
	if config.OmitAccessSchema {
		access = nil
	}
	provider := config.DeploySchema
	if len(provider) == 0 {
		provider = schema(config.ProviderType, "deploy", []map[string]any{
			{"name": "count", "valueType": "number", "labelKey": "plugin.e2e.count", "default": 1},
			{"name": config.DeployField, "valueType": "text", "labelKey": "plugin.e2e.message"},
		})
	}
	return &plugin.ConfigSchema{
		AccessSchemaJSON: access,
		ProviderSchemaJSON: provider,
		I18n: map[string]map[string]string{
			"en": {"plugin.e2e.name": "E2E Webhook", "plugin.e2e.access": "E2E Access", "plugin.e2e.url": "Webhook URL", "plugin.e2e.token": "Token " + config.SchemaRevision, "plugin.e2e.count": "Count", "plugin.e2e.message": "Message " + config.SchemaRevision},
			"zh": {"plugin.e2e.name": "E2E Webhook", "plugin.e2e.access": "E2E Access", "plugin.e2e.url": "Webhook URL", "plugin.e2e.token": "Token " + config.SchemaRevision, "plugin.e2e.count": "Count", "plugin.e2e.message": "Message " + config.SchemaRevision},
		},
	}
}

func (d *fixturePlugin) GetMetadata(ctx context.Context) (*plugin.Metadata, error) {
	return d.metadata(ctx)
}

func (d *fixturePlugin) GetConfigSchema(context.Context) (*plugin.ConfigSchema, error) {
	return d.configSchema(), nil
}

func schema(provider, category string, columns []map[string]any) []byte {
	data, err := json.Marshal(map[string]any{"schemaVersion": "form/v1", "provider": provider, "category": category, "schema": map[string]any{"columns": columns}})
	if err != nil {
		panic(err)
	}
	return data
}

func (d *fixturePlugin) Process(ctx context.Context, req *plugin.Request, logger *slog.Logger) (*plugin.Result, error) {
	switch req.Capability {
	case plugin.CapabilityDeploy:
		return d.processDeploy(ctx, req, logger)
	case plugin.CapabilityDns:
		return d.processDns(ctx, req)
	case plugin.CapabilityNotify:
		return d.processNotify(ctx, req)
	case plugin.CapabilityPurge:
		return d.processPurge(ctx, req, logger)
	}
	return nil, fmt.Errorf("unsupported capability %q", req.Capability)
}

func (d *fixturePlugin) processDeploy(ctx context.Context, req *plugin.Request, logger *slog.Logger) (*plugin.Result, error) {
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
		"cert": req.Deploy.CertificatePEM, "key": req.Deploy.PrivateKeyPEM,
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
	return &plugin.Result{Deploy: &plugin.DeployResult{ExtendedDataJSON: `{"deployed":true}`}}, nil
}

func (d *fixturePlugin) processDns(ctx context.Context, req *plugin.Request) (*plugin.Result, error) {
	if d.config.PresentError != "" {
		return nil, fmt.Errorf("%s", d.config.PresentError)
	}
	if err := postFixtureJSON(ctx, d.config, req.AccessConfigJSON, map[string]any{
		"action": "present", "domain": req.Dns.Domain, "token": req.Dns.Token, "keyAuth": req.Dns.KeyAuth,
		"pid": os.Getpid(), "version": d.config.SchemaRevision,
	}); err != nil {
		return nil, err
	}
	return &plugin.Result{Dns: &plugin.DnsResult{}}, nil
}

func (d *fixturePlugin) processNotify(ctx context.Context, req *plugin.Request) (*plugin.Result, error) {
	if d.config.NotifyError != "" {
		return nil, fmt.Errorf("%s", d.config.NotifyError)
	}
	if err := postFixtureJSON(ctx, d.config, req.AccessConfigJSON, map[string]any{
		"subject": req.Notify.Subject, "message": req.Notify.Message,
		"pid": os.Getpid(), "version": d.config.SchemaRevision,
	}); err != nil {
		return nil, err
	}
	return &plugin.Result{Notify: &plugin.NotifyResult{ExtendedDataJSON: `{"notified":true}`}}, nil
}

func (d *fixturePlugin) processPurge(ctx context.Context, req *plugin.Request, logger *slog.Logger) (*plugin.Result, error) {
	if d.config.PurgeError != "" {
		return nil, fmt.Errorf("%s", d.config.PurgeError)
	}
	body, err := postFixtureJSONResponse(ctx, d.config, req.AccessConfigJSON, map[string]any{
		"action": "purge", "expirySeconds": req.Purge.ExpirySeconds,
		"pid": os.Getpid(), "version": d.config.SchemaRevision,
		"config": json.RawMessage(req.ExtendedConfigJSON),
	})
	if err != nil {
		return nil, err
	}
	var vendor struct {
		Count int `json:"count"`
	}
	count := 0
	if len(body) > 0 {
		if err := json.Unmarshal(body, &vendor); err != nil {
			return nil, err
		}
		count = vendor.Count
	}
	if logger != nil {
		logger.Info("purge fixture responded", slog.Int("count", count))
	}
	return &plugin.Result{Purge: &plugin.PurgeResult{Count: count, ExtendedDataJSON: `{"purged":true}`}}, nil
}

func (d *fixturePlugin) PostProcess(ctx context.Context, req *plugin.Request, logger *slog.Logger) error {
	if req.Capability != plugin.CapabilityDns {
		return plugin.NoopFinalizer{}.PostProcess(ctx, req, logger)
	}
	if d.config.CleanUpError != "" {
		return fmt.Errorf("%s", d.config.CleanUpError)
	}
	return postFixtureJSON(ctx, d.config, req.AccessConfigJSON, map[string]any{
		"action": "cleanup", "domain": req.Dns.Domain, "token": req.Dns.Token, "keyAuth": req.Dns.KeyAuth,
		"pid": os.Getpid(), "version": d.config.SchemaRevision,
	})
}

func postFixtureJSON(ctx context.Context, config fixtureConfig, accessJSON string, payload map[string]any) error {
	_, err := postFixtureJSONResponse(ctx, config, accessJSON, payload)
	return err
}

func postFixtureJSONResponse(ctx context.Context, config fixtureConfig, accessJSON string, payload map[string]any) ([]byte, error) {
	var access struct {
		URL   string `json:"url"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal([]byte(accessJSON), &access); err != nil {
		return nil, err
	}
	if access.URL == "" {
		return nil, nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, access.URL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if access.Token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+access.Token)
	}
	response, err := (&http.Client{Timeout: 10 * time.Second}).Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("webhook status %d", response.StatusCode)
	}
	return body, nil
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
	if config.Capability == "" {
		config.Capability = plugin.CapabilityDeploy
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
	githubplugin.Serve(plugin.ServeConfig(&fixturePlugin{config: config}))
}
