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
	"sync"

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

func (m *Manager) Deploy(ctx context.Context, dp *DiscoveredPlugin, req *DeployRequest, forward *slog.Logger) (*DeployResult, error) {
	if m.cfg.DeployTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, m.cfg.DeployTimeout)
		defer cancel()
	}
	redact := redactorFor(req)
	sink := newForwardLogger(forward, redact, m.cfg.MaxLogFrames)
	var res *DeployResult
	err := m.withPlugin(ctx, dp, redact, func(d Deployer) error {
		var err error
		res, err = d.Deploy(ctx, req, sink)
		return err
	})
	return res, err
}

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n, err := b.buf.Write(p)
	if b.buf.Len() > 2*stderrTailLimit {
		data := b.buf.Bytes()
		copy(data, data[len(data)-stderrTailLimit:])
		b.buf.Truncate(stderrTailLimit)
	}
	return n, err
}

func (b *syncBuffer) tail(limit int) []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	data := b.buf.Bytes()
	if len(data) > limit {
		data = data[len(data)-limit:]
	}
	return bytes.Clone(data)
}

type dispensed struct {
	client   *githubplugin.Client
	deployer Deployer
	stderr   *syncBuffer
}

func (m *Manager) clientConfig(dp *DiscoveredPlugin, stderr *syncBuffer) *githubplugin.ClientConfig {
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
	stderr := &syncBuffer{}
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

func stderrTail(buf *syncBuffer, redact secretRedactor) string {
	if buf == nil {
		return ""
	}
	if data := buf.tail(stderrTailLimit); len(data) > 0 {
		return redact(string(data))
	}
	return ""
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
		collectPEM(&secrets, req.CertificatePEM)
		collectPEM(&secrets, req.PrivateKeyPEM)
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

func collectPEM(secrets *[]string, pem string) {
	pem = strings.TrimSpace(pem)
	if len(pem) >= 4 {
		*secrets = append(*secrets, pem)
	}
	for _, line := range strings.Split(pem, "\n") {
		line = strings.TrimSpace(line)
		if len(line) >= 32 && !strings.HasPrefix(line, "-----") {
			*secrets = append(*secrets, line)
		}
	}
}

type forwardHandler struct {
	mu     sync.Mutex
	inner  slog.Handler
	redact secretRedactor
	cap    int
	count  int
}

func newForwardLogger(forward *slog.Logger, redact secretRedactor, cap int) *slog.Logger {
	if forward == nil {
		return nil
	}
	return slog.New(&forwardHandler{inner: forward.Handler(), redact: redact, cap: cap})
}

func (h *forwardHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.inner.Enabled(ctx, l)
}

func (h *forwardHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	if h.cap > 0 && h.count >= h.cap {
		h.mu.Unlock()
		return nil
	}
	h.count++
	h.mu.Unlock()
	attrs := redactAttrs(r, h.redact)
	rec := slog.NewRecord(r.Time, r.Level, h.redact(r.Message), r.PC)
	rec.AddAttrs(attrs...)
	return h.inner.Handle(ctx, rec)
}

func (h *forwardHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *forwardHandler) WithGroup(_ string) slog.Handler      { return h }

func redactAttrs(r slog.Record, redact secretRedactor) []slog.Attr {
	out := make([]slog.Attr, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		a.Value = a.Value.Resolve()
		if a.Value.Kind() == slog.KindString {
			a.Value = slog.StringValue(redact(a.Value.String()))
		}
		out = append(out, a)
		return true
	})
	return out
}
