//go:build fakeplugin

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	githubplugin "github.com/hashicorp/go-plugin"

	"github.com/certimate-go/certimate/pkg/plugin"
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func statePath() string {
	return envOr("FAKEPLUGIN_STATE_FILE", filepath.Join(os.TempDir(), "fakeplugin-state.json"))
}

func appendState(event string) {
	f, err := os.OpenFile(statePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s\n", event)
}

func metadata() *plugin.Metadata {
	accessType := os.Getenv("FAKEPLUGIN_ACCESS_TYPE")
	if accessType == "" {
		accessType = os.Getenv("FAKEPLUGIN_PROVIDER_TYPE")
	}
	capability := envOr("FAKEPLUGIN_CAPABILITY", plugin.CapabilityDeploy)
	return &plugin.Metadata{
		ProviderType:         envOr("FAKEPLUGIN_PROVIDER_TYPE", "fake"),
		AccessProviderType:   accessType,
		Capability:           capability,
		ProtocolVersion:      plugin.ProtocolVersion,
		Category:             "other",
		DisplayNameKey:       "plugin.fake.name",
		AccessDisplayNameKey: "plugin.fake.name",
	}
}

func omitAccessSchema() bool {
	return os.Getenv("FAKEPLUGIN_OMIT_ACCESS_SCHEMA") == "1"
}

func configSchema() *plugin.ConfigSchema {
	var accessSchema []byte
	if !omitAccessSchema() {
		accessSchema = []byte(`{"schemaVersion":"form/v1","provider":"fake","category":"access","schema":{"columns":[{"name":"apiKey","valueType":"string","required":true,"secret":true}]}}`)
	}
	return &plugin.ConfigSchema{
		AccessSchemaJSON: accessSchema,
		ProviderSchemaJSON: []byte(`{"schemaVersion":"form/v1","provider":"fake","category":"deploy","schema":{"columns":[{"name":"endpoint","valueType":"string","required":true}]}}`),
		I18n: map[string]map[string]string{
			"en": {"plugin.fake.name": "Fake Plugin"},
		},
	}
}

type fakePlugin struct {
	behavior string
}

func (f *fakePlugin) GetMetadata(context.Context) (*plugin.Metadata, error) {
	return metadata(), nil
}

func (f *fakePlugin) GetConfigSchema(context.Context) (*plugin.ConfigSchema, error) {
	return configSchema(), nil
}

func (f *fakePlugin) Process(ctx context.Context, req *plugin.Request, logger *slog.Logger) (*plugin.Result, error) {
	switch req.Capability {
	case plugin.CapabilityDeploy:
		return f.processDeploy(ctx, req, logger)
	case plugin.CapabilityDns:
		return f.processDns(ctx, req)
	case plugin.CapabilityNotify:
		return f.processNotify(ctx, req)
	case plugin.CapabilityPurge:
		return f.processPurge(ctx, req, logger)
	}
	return nil, fmt.Errorf("unsupported capability %q", req.Capability)
}

func (f *fakePlugin) processDeploy(_ context.Context, req *plugin.Request, logger *slog.Logger) (*plugin.Result, error) {
	fmt.Fprintf(os.Stderr, "fakeplugin deploy behavior=%s access=%s\n", f.behavior, req.AccessConfigJSON)
	if logger != nil {
		logger.Info("fakeplugin deploy starting", slog.String("behavior", f.behavior))
	}
	switch f.behavior {
	case "crash":
		fmt.Fprintln(os.Stderr, "fakeplugin crashing now")
		os.Exit(2)
	case "configerror":
		return nil, fmt.Errorf("bad config: missing field")
	case "error-echo":
		return nil, fmt.Errorf("remote rejected access %s and key %s", req.AccessConfigJSON, req.Deploy.PrivateKeyPEM)
	}
	appendState("deploy")
	return &plugin.Result{Deploy: &plugin.DeployResult{ExtendedDataJSON: `{"deployed":true}`}}, nil
}

func (f *fakePlugin) processDns(ctx context.Context, req *plugin.Request) (*plugin.Result, error) {
	fmt.Fprintf(os.Stderr, "fakeplugin present pid=%d domain=%s\n", os.Getpid(), req.Dns.Domain)
	switch f.behavior {
	case "present-crash":
		appendState("present")
		fmt.Fprintln(os.Stderr, "fakeplugin exiting shortly after present")
		time.AfterFunc(150*time.Millisecond, func() { os.Exit(2) })
	case "present-error":
		if req.Dns.Domain == "failed.example.test" {
			appendState("present:error")
			return nil, fmt.Errorf("fake present failure")
		}
	case "present-block":
		appendState("present:block:" + req.Dns.Domain)
		<-ctx.Done()
		return nil, ctx.Err()
	}
	appendState("present")
	return &plugin.Result{Dns: &plugin.DnsResult{}}, nil
}

func (f *fakePlugin) processNotify(_ context.Context, req *plugin.Request) (*plugin.Result, error) {
	fmt.Fprintf(os.Stderr, "fakeplugin notify subject=%s\n", req.Notify.Subject)
	switch f.behavior {
	case "notify-error":
		return nil, fmt.Errorf("fake notify failure")
	}
	appendState("notify:" + req.Notify.Subject + "|" + req.Notify.Message)
	return &plugin.Result{Notify: &plugin.NotifyResult{}}, nil
}

func (f *fakePlugin) processPurge(ctx context.Context, req *plugin.Request, logger *slog.Logger) (*plugin.Result, error) {
	fmt.Fprintf(os.Stderr, "fakeplugin purge behavior=%s expiry=%ds\n", f.behavior, req.Purge.ExpirySeconds)
	if logger != nil {
		logger.Info("fakeplugin purge starting", slog.String("behavior", f.behavior), slog.Int64("expirySeconds", req.Purge.ExpirySeconds))
	}
	switch f.behavior {
	case "purge-crash":
		fmt.Fprintln(os.Stderr, "fakeplugin crashing now")
		os.Exit(2)
	case "purge-error":
		return nil, fmt.Errorf("fake purge failure")
	case "purge-error-echo":
		return nil, fmt.Errorf("remote rejected access %s", req.AccessConfigJSON)
	case "purge-block":
		appendState("purge:block")
		<-ctx.Done()
		return nil, ctx.Err()
	}
	appendState(fmt.Sprintf("purge:%d", req.Purge.ExpirySeconds))
	count := 0
	if n, err := strconv.Atoi(envOr("FAKEPLUGIN_PURGE_COUNT", "0")); err == nil && n >= 0 {
		count = n
	}
	return &plugin.Result{Purge: &plugin.PurgeResult{Count: count, ExtendedDataJSON: `{"purged":true}`}}, nil
}

func (f *fakePlugin) PostProcess(ctx context.Context, req *plugin.Request, logger *slog.Logger) error {
	if req.Capability != plugin.CapabilityDns {
		return plugin.NoopFinalizer{}.PostProcess(ctx, req, logger)
	}
	fmt.Fprintf(os.Stderr, "fakeplugin cleanup pid=%d domain=%s\n", os.Getpid(), req.Dns.Domain)
	switch f.behavior {
	case "cleanup-error":
		if req.Dns.Domain == "cleanup-fail.example.test" {
			appendState("cleanup:error")
			return fmt.Errorf("fake cleanup failure")
		}
	case "cleanup-block":
		appendState("cleanup:block:" + req.Dns.Domain)
		<-ctx.Done()
		return ctx.Err()
	}
	appendState("cleanup")
	return nil
}

func main() {
	behavior := envOr("FAKEPLUGIN_BEHAVIOR", "ok")
	githubplugin.Serve(plugin.ServeConfig(&fakePlugin{behavior: behavior}))
}
