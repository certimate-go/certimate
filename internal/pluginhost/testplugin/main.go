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
	Capability             string          `json:"capability"`
	SchemaRevision         string          `json:"schemaRevision"`
	AccessField            string          `json:"accessField"`
	DeployField            string          `json:"deployField"`
	BootstrapError         string          `json:"bootstrapError"`
	NotifyError            string          `json:"notifyError"`
	PresentError           string          `json:"presentError"`
	MetadataDelayMS        int             `json:"metadataDelayMs"`
	MetadataURL            string          `json:"metadataUrl"`
	StartupLog             string          `json:"startupLog"`
	AccessSchema           json.RawMessage `json:"accessSchema"`
	DeploySchema           json.RawMessage `json:"deploySchema"`
}

type fixtureDeployer struct {
	config fixtureConfig
}

func fixtureMetadata(ctx context.Context, config fixtureConfig) (*plugin.Metadata, error) {
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
		ProtocolVersion:      plugin.ProtocolVersion,
		DeployCategory:       "other",
		DeployDisplayNameKey: "plugin.e2e.name",
		AccessDisplayNameKey: "plugin.e2e.access",
	}, nil
}

func fixtureConfigSchema(config fixtureConfig) *plugin.ConfigSchema {
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
	deploy := config.DeploySchema
	if len(deploy) == 0 {
		deploy = schema(config.ProviderType, "deploy", []map[string]any{
			{"name": "count", "valueType": "number", "labelKey": "plugin.e2e.count", "default": 1},
			{"name": config.DeployField, "valueType": "text", "labelKey": "plugin.e2e.message"},
		})
	}
	return &plugin.ConfigSchema{
		AccessSchemaJSON: access,
		DeploySchemaJSON: deploy,
		I18n: map[string]map[string]string{
			"en": {"plugin.e2e.name": "E2E Webhook", "plugin.e2e.access": "E2E Access", "plugin.e2e.url": "Webhook URL", "plugin.e2e.token": "Token " + config.SchemaRevision, "plugin.e2e.count": "Count", "plugin.e2e.message": "Message " + config.SchemaRevision},
			"zh": {"plugin.e2e.name": "E2E Webhook", "plugin.e2e.access": "E2E Access", "plugin.e2e.url": "Webhook URL", "plugin.e2e.token": "Token " + config.SchemaRevision, "plugin.e2e.count": "Count", "plugin.e2e.message": "Message " + config.SchemaRevision},
		},
	}
}

func (d *fixtureDeployer) GetMetadata(ctx context.Context) (*plugin.Metadata, error) {
	return fixtureMetadata(ctx, d.config)
}

func (d *fixtureDeployer) GetConfigSchema(context.Context) (*plugin.ConfigSchema, error) {
	return fixtureConfigSchema(d.config), nil
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

type fixtureDnsProvider struct {
	config fixtureConfig
}

func (d *fixtureDnsProvider) GetMetadata(ctx context.Context) (*plugin.Metadata, error) {
	return fixtureMetadata(ctx, d.config)
}

func (d *fixtureDnsProvider) GetConfigSchema(context.Context) (*plugin.ConfigSchema, error) {
	return fixtureConfigSchema(d.config), nil
}

func (d *fixtureDnsProvider) Present(ctx context.Context, req *plugin.DnsRequest) error {
	if d.config.PresentError != "" {
		return fmt.Errorf("%s", d.config.PresentError)
	}
	return postFixtureJSON(ctx, d.config, req.AccessConfigJSON, map[string]any{
		"action": "present", "domain": req.Domain, "token": req.Token, "keyAuth": req.KeyAuth,
		"pid": os.Getpid(), "version": d.config.SchemaRevision,
	})
}

func (d *fixtureDnsProvider) CleanUp(ctx context.Context, req *plugin.DnsRequest) error {
	return postFixtureJSON(ctx, d.config, req.AccessConfigJSON, map[string]any{
		"action": "cleanup", "domain": req.Domain, "token": req.Token, "keyAuth": req.KeyAuth,
		"pid": os.Getpid(), "version": d.config.SchemaRevision,
	})
}

type fixtureNotifier struct {
	config fixtureConfig
}

func (n *fixtureNotifier) GetMetadata(ctx context.Context) (*plugin.Metadata, error) {
	return fixtureMetadata(ctx, n.config)
}

func (n *fixtureNotifier) GetConfigSchema(context.Context) (*plugin.ConfigSchema, error) {
	return fixtureConfigSchema(n.config), nil
}

func (n *fixtureNotifier) Notify(ctx context.Context, req *plugin.NotifyRequest) (*plugin.NotifyResult, error) {
	if n.config.NotifyError != "" {
		return nil, fmt.Errorf("%s", n.config.NotifyError)
	}
	if err := postFixtureJSON(ctx, n.config, req.AccessConfigJSON, map[string]any{
		"subject": req.Subject, "message": req.Message,
		"pid": os.Getpid(), "version": n.config.SchemaRevision,
	}); err != nil {
		return nil, err
	}
	return &plugin.NotifyResult{ExtendedDataJSON: `{"notified":true}`}, nil
}

func postFixtureJSON(ctx context.Context, config fixtureConfig, accessJSON string, payload map[string]any) error {
	var access struct {
		URL   string `json:"url"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal([]byte(accessJSON), &access); err != nil {
		return err
	}
	if access.URL == "" {
		return nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, access.URL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if access.Token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+access.Token)
	}
	response, err := (&http.Client{Timeout: 10 * time.Second}).Do(httpReq)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("webhook status %d", response.StatusCode)
	}
	return nil
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
	var served githubplugin.Plugin
	switch config.Capability {
	case plugin.CapabilityDns:
		served = &plugin.DnsGRPCPlugin{Impl: &fixtureDnsProvider{config: config}}
	case plugin.CapabilityNotify:
		served = &plugin.NotifyGRPCPlugin{Impl: &fixtureNotifier{config: config}}
	default:
		served = &plugin.DeployerGRPCPlugin{Impl: &fixtureDeployer{config: config}}
	}
	githubplugin.Serve(&githubplugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig,
		Plugins: map[string]githubplugin.Plugin{
			plugin.ServiceNameForCapability(config.Capability): served,
		},
		GRPCServer: githubplugin.DefaultGRPCServer,
	})
}
