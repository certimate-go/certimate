package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	githubplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const stderrTailLimit = 8 * 1024

type secretRedactor func(string) string

type Manager struct {
	cfg    PluginConfig
	logger *slog.Logger
}

func NewManager(cfg PluginConfig, logger *slog.Logger) *Manager {
	cfg.defaults()
	if logger == nil {
		logger = slog.Default()
	}
	return &Manager{cfg: cfg, logger: logger.With(slog.String("component", "pluginmgr"))}
}

func (m *Manager) Config() PluginConfig { return m.cfg }

func (m *Manager) Bootstrap(ctx context.Context, dp *DiscoveredPlugin) (*Metadata, *ConfigSchema, error) {
	var meta *Metadata
	var schema *ConfigSchema
	err := m.withPlugin(ctx, dp, noopRedactor, func(d Deployer) error {
		var err error
		meta, err = d.GetMetadata(ctx)
		if err != nil {
			return err
		}
		schema, err = d.GetConfigSchema(ctx)
		return err
	})
	if err != nil {
		return nil, nil, err
	}
	return meta, schema, nil
}

func (m *Manager) Deploy(ctx context.Context, dp *DiscoveredPlugin, req *DeployRequest) (*DeployResult, error) {
	var res *DeployResult
	err := m.withPlugin(ctx, dp, redactorFor(req), func(d Deployer) error {
		var err error
		res, err = d.Deploy(ctx, req)
		return err
	})
	return res, err
}

type dispensed struct {
	client   *githubplugin.Client
	deployer Deployer
	stderr   *bytes.Buffer
}

func (m *Manager) clientConfig(dp *DiscoveredPlugin, stderr *bytes.Buffer) *githubplugin.ClientConfig {
	cfg := &githubplugin.ClientConfig{
		HandshakeConfig:  HandshakeConfig,
		Plugins:          PluginSetForDeployer(),
		Cmd:              exec.Command(dp.BinaryPath),
		AllowedProtocols: []githubplugin.Protocol{githubplugin.ProtocolGRPC},
		Logger:           NewHclogLogger(m.logger),
		StartTimeout:     m.cfg.StartTimeout,
		MinPort:          m.cfg.MinPort,
		MaxPort:          m.cfg.MaxPort,
		AutoMTLS:         true,
	}
	if stderr != nil {
		cfg.Stderr = stderr
	}
	return cfg
}

func (m *Manager) dispense(dp *DiscoveredPlugin) (*dispensed, error) {
	stderr := &bytes.Buffer{}
	client := githubplugin.NewClient(m.clientConfig(dp, stderr))
	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, mapStartError(dp, err)
	}
	raw, err := rpcClient.Dispense(PluginName)
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("plugin: dispense %q: %w", dp.Manifest.ProviderType, err)
	}
	d, ok := raw.(Deployer)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf("plugin: dispensed %q is not a Deployer (%T)", dp.Manifest.ProviderType, raw)
	}
	return &dispensed{client: client, deployer: d, stderr: stderr}, nil
}

func (m *Manager) withPlugin(ctx context.Context, dp *DiscoveredPlugin, redact secretRedactor, fn func(Deployer) error) error {
	d, err := m.dispense(dp)
	if err != nil {
		return err
	}
	defer d.client.Kill()

	done := make(chan error, 1)
	go func() { done <- fn(d.deployer) }()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return m.wrapError(dp, d, redact, err)
	}
}

func (m *Manager) wrapError(dp *DiscoveredPlugin, d *dispensed, redact secretRedactor, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	tail := stderrTail(d.stderr, redact)
	if d.client.Exited() {
		return &ErrPluginCrashed{ProviderType: dp.Manifest.ProviderType, StderrTail: tail, Inner: err}
	}
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Unavailable, codes.Canceled:
			return &ErrPluginCrashed{ProviderType: dp.Manifest.ProviderType, StderrTail: tail, Inner: err}
		}
		return fmt.Errorf("plugin %q: %s", dp.Manifest.ProviderType, st.Message())
	}
	return fmt.Errorf("plugin %q: %w", dp.Manifest.ProviderType, err)
}

func mapStartError(dp *DiscoveredPlugin, err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "incompatible API version"),
		strings.Contains(msg, "plugin version"),
		strings.Contains(strings.ToLower(msg), "magic cookie"),
		strings.Contains(strings.ToLower(msg), "protocol"):
		return &ErrPluginIncompatible{ProviderType: dp.Manifest.ProviderType, Reason: msg}
	}
	return err
}

func stderrTail(buf *bytes.Buffer, redact secretRedactor) string {
	if buf == nil || buf.Len() == 0 {
		return ""
	}
	data := buf.Bytes()
	if len(data) > stderrTailLimit {
		data = data[len(data)-stderrTailLimit:]
	}
	return redact(string(data))
}

func noopRedactor(s string) string { return s }

func redactorFor(req *DeployRequest) secretRedactor {
	var secrets []string
	collect := func(jsonStr string) {
		var m map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
			return
		}
		for _, v := range flattenValues(m) {
			if s, ok := v.(string); ok && len(s) >= 4 {
				secrets = append(secrets, s)
			}
		}
	}
	if req != nil {
		collect(req.AccessConfigJSON)
		collect(req.ExtendedConfigJSON)
	}
	return func(s string) string {
		for _, sec := range secrets {
			s = strings.ReplaceAll(s, sec, "[REDACTED]")
		}
		return s
	}
}

func flattenValues(v any) []any {
	var out []any
	switch t := v.(type) {
	case map[string]any:
		for _, child := range t {
			out = append(out, flattenValues(child)...)
		}
	case []any:
		for _, child := range t {
			out = append(out, flattenValues(child)...)
		}
	default:
		out = append(out, v)
	}
	return out
}
