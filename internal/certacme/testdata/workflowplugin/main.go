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

	goplugin "github.com/hashicorp/go-plugin"

	"github.com/certimate-go/certimate/pkg/plugin"
)

type workflowPlugin struct {
	metadata plugin.Metadata
}

func main() {
	name := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
	metadata := plugin.Metadata{
		ProviderType: name, ProtocolVersion: plugin.ProtocolVersion, DisplayNameKey: "plugin." + name + ".name",
	}
	switch name {
	case "mockdns":
		metadata.Capability, metadata.AccessProviderType = plugin.CapabilityDns, "mockdns"
	case "mockdns-deploy":
		metadata.Capability, metadata.AccessProviderType = plugin.CapabilityDeploy, "mockdns"
	case "mocknotify":
		metadata.Capability, metadata.AccessProviderType = plugin.CapabilityNotify, "webhook"
	default:
		panic("unknown workflow plugin fixture: " + name)
	}
	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig,
		Plugins:         goplugin.PluginSet{plugin.PluginName: &plugin.GRPCPlugin{Impl: &workflowPlugin{metadata}}},
		GRPCServer:      goplugin.DefaultGRPCServer,
	})
}

func (p *workflowPlugin) GetMetadata(context.Context) (*plugin.Metadata, error) {
	return &p.metadata, nil
}

func (p *workflowPlugin) GetConfigSchema(context.Context) (*plugin.ConfigSchema, error) {
	schema := &plugin.ConfigSchema{ProviderSchemaJSON: []byte(`{"schemaVersion":"form/v1","schema":{"columns":[]}}`)}
	if p.metadata.Capability == plugin.CapabilityDns {
		schema.AccessSchemaJSON = []byte(`{"schemaVersion":"form/v1","schema":{"columns":[{"name":"apiUrl","valueType":"text"},{"name":"apiKey","valueType":"secret","secret":true}]}}`)
	}
	return schema, nil
}

func (p *workflowPlugin) Process(ctx context.Context, req *plugin.Request, _ *slog.Logger) (*plugin.Result, error) {
	switch p.metadata.Capability {
	case plugin.CapabilityDns:
		err := send(ctx, req, "/present", dnsPayload(req.Dns))
		return &plugin.Result{Dns: &plugin.DnsResult{}}, err
	case plugin.CapabilityDeploy:
		err := send(ctx, req, "/deploy", map[string]string{
			"certificate": req.Deploy.CertificatePEM,
			"privateKey":  req.Deploy.PrivateKeyPEM,
		})
		return &plugin.Result{Deploy: &plugin.DeployResult{}}, err
	default:
		err := send(ctx, req, "", map[string]string{"subject": req.Notify.Subject, "message": req.Notify.Message})
		return &plugin.Result{Notify: &plugin.NotifyResult{}}, err
	}
}

func (p *workflowPlugin) PostProcess(ctx context.Context, req *plugin.Request, _ *slog.Logger) error {
	if p.metadata.Capability == plugin.CapabilityDns {
		return send(ctx, req, "/cleanup", dnsPayload(req.Dns))
	}
	return nil
}

func dnsPayload(dns *plugin.DnsPayload) map[string]string {
	return map[string]string{"fqdn": dns.Domain, "token": dns.Token, "keyAuth": dns.KeyAuth}
}

func send(ctx context.Context, req *plugin.Request, path string, body map[string]string) error {
	var access struct {
		APIURL string `json:"apiUrl"`
		APIKey string `json:"apiKey"`
		URL    string `json:"url"`
	}
	if err := json.Unmarshal([]byte(req.AccessConfigJSON), &access); err != nil {
		return err
	}
	endpoint := access.URL
	if path != "" {
		endpoint = access.APIURL + path
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	if access.APIKey != "" {
		request.Header.Set("Authorization", "Bearer "+access.APIKey)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("fixture endpoint returned HTTP %d", response.StatusCode)
	}
	return nil
}
