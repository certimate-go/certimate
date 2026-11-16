package plugin

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	githubplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const stderrTailLimit = 8 * 1024

type secretRedactor func(string) string

type redactedPluginError struct {
	error
	message string
}

func (e *redactedPluginError) Error() string { return e.message }

func (e *redactedPluginError) Unwrap() error { return e.error }

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
	ctx, cancel := context.WithTimeout(ctx, m.cfg.StartTimeout)
	defer cancel()
	session, err := m.StartSession(ctx, dp)
	if err != nil {
		return nil, nil, err
	}
	defer session.Close()
	meta, err := session.GetMetadata(ctx)
	if err != nil {
		return nil, nil, err
	}
	schema, err := session.GetConfigSchema(ctx)
	if err != nil {
		return nil, nil, err
	}
	return meta, schema, nil
}

var sessionSequence atomic.Uint64

func newExecutionID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Sprintf("exec-%d", sessionSequence.Add(1))
	}
	return fmt.Sprintf("%s-%d", hex.EncodeToString(buf[:]), sessionSequence.Add(1))
}

type DnsKey struct {
	Domain  string
	Token   string
	KeyAuth string
}

type ErrExecutionNotFound struct {
	ExecutionID string
}

func (e *ErrExecutionNotFound) Error() string {
	return fmt.Sprintf("plugin: execution %q is not registered in this session", e.ExecutionID)
}

func IsExecutionNotFound(err error) bool {
	var e *ErrExecutionNotFound
	return errors.As(err, &e)
}

type SessionExecution struct {
	ID      string
	Request *Request
}

type sessionExecutionState struct {
	execution *SessionExecution
	postDone  bool
	postErr   error
}

type Session struct {
	manager *Manager
	dp      *DiscoveredPlugin
	d       *dispensed

	mu         sync.Mutex
	redactions []string
	executions map[string]*sessionExecutionState
	dnsKeys    map[DnsKey]string
	closeOnce  sync.Once
	closed     bool
}

func (m *Manager) StartSession(ctx context.Context, dp *DiscoveredPlugin) (*Session, error) {
	d, err := m.dispense(dp)
	if err != nil {
		return nil, err
	}
	s := &Session{
		manager:    m,
		dp:         dp,
		d:          d,
		executions: make(map[string]*sessionExecutionState),
		dnsKeys:    make(map[DnsKey]string),
	}
	runtime.AddCleanup(s, func(dd *dispensed) { dd.client.Kill() }, d)
	return s, nil
}

func (s *Session) GetMetadata(ctx context.Context) (*Metadata, error) {
	if err := s.checkOpen(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	redact := s.redactorLocked()
	s.mu.Unlock()
	meta, err := s.d.impl.GetMetadata(ctx)
	if err != nil {
		return nil, s.manager.wrapError(s.dp, s.d, redact, err)
	}
	return meta, nil
}

func (s *Session) Pid() int {
	if s.d == nil || s.d.client == nil {
		return 0
	}
	config := s.d.client.ReattachConfig()
	if config == nil {
		return 0
	}
	return config.Pid
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

type capabilityClient interface {
	GetMetadata(ctx context.Context) (*Metadata, error)
	GetConfigSchema(ctx context.Context) (*ConfigSchema, error)
}

var _ capabilityClient = (*pluginServiceClient)(nil)

type dispensed struct {
	client *githubplugin.Client
	impl   *pluginServiceClient
	stderr *syncBuffer
}

func (m *Manager) clientConfig(dp *DiscoveredPlugin, stderr *syncBuffer) (*githubplugin.ClientConfig, error) {
	cfg := &githubplugin.ClientConfig{
		HandshakeConfig:  HandshakeConfig,
		Plugins:          PluginSet{PluginName: &GRPCPlugin{}},
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
	return cfg, nil
}

func (m *Manager) dispense(dp *DiscoveredPlugin) (*dispensed, error) {
	stderr := &syncBuffer{}
	cfg, err := m.clientConfig(dp, stderr)
	if err != nil {
		return nil, err
	}
	client := githubplugin.NewClient(cfg)
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
	impl, ok := raw.(*pluginServiceClient)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf("plugin: dispensed %q does not implement the unified plugin service (%T)", dp.Manifest.ProviderType, raw)
	}
	return &dispensed{client: client, impl: impl, stderr: stderr}, nil
}

func (s *Session) checkOpen() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return fmt.Errorf("plugin: session for %q is already closed", s.dp.Manifest.ProviderType)
	}
	return nil
}

func (s *Session) GetConfigSchema(ctx context.Context) (*ConfigSchema, error) {
	if err := s.checkOpen(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	redact := s.redactorLocked()
	s.mu.Unlock()
	schema, err := s.d.impl.GetConfigSchema(ctx)
	if err != nil {
		return nil, s.manager.wrapError(s.dp, s.d, redact, err)
	}
	return schema, nil
}

func (s *Session) Process(ctx context.Context, req *Request, forward *slog.Logger) (*Result, error) {
	if err := s.checkOpen(); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, fmt.Errorf("the request is nil")
	}
	if req.Capability != s.dp.Manifest.Capability {
		return nil, &ErrPluginIncompatible{ProviderType: s.dp.Manifest.ProviderType, Reason: fmt.Sprintf("capability %q does not serve %q", s.dp.Manifest.Capability, req.Capability)}
	}
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("plugin %q: %w", s.dp.Manifest.ProviderType, err)
	}
	if s.manager.cfg.ProcessTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.manager.cfg.ProcessTimeout)
		defer cancel()
	}

	immutable := req.Clone()
	if immutable.ExecutionID == "" {
		immutable.ExecutionID = newExecutionID()
	}

	s.mu.Lock()
	state := &sessionExecutionState{execution: &SessionExecution{ID: immutable.ExecutionID, Request: immutable}}
	s.executions[immutable.ExecutionID] = state
	if immutable.Dns != nil {
		key := DnsKey{Domain: immutable.Dns.Domain, Token: immutable.Dns.Token, KeyAuth: immutable.Dns.KeyAuth}
		if _, exists := s.dnsKeys[key]; !exists {
			s.dnsKeys[key] = immutable.ExecutionID
		}
	}
	s.collectSecretsLocked(immutable)
	redact := s.redactorLocked()
	s.mu.Unlock()

	req.ExecutionID = immutable.ExecutionID
	sink := newForwardLogger(forward, redact, s.manager.cfg.MaxLogFrames)
	res, err := s.d.impl.Process(ctx, immutable, sink)
	if err != nil {
		return nil, s.manager.wrapError(s.dp, s.d, redact, err)
	}
	if verr := res.Validate(immutable.Capability); verr != nil {
		return nil, s.manager.wrapError(s.dp, s.d, redact, &ErrPluginProtocol{ProviderType: s.dp.Manifest.ProviderType, Reason: verr.Error()})
	}
	return res, nil
}

func (s *Session) PostProcess(ctx context.Context, executionID string, forward *slog.Logger) error {
	if err := s.checkOpen(); err != nil {
		return err
	}
	s.mu.Lock()
	state, ok := s.executions[executionID]
	if !ok {
		s.mu.Unlock()
		return &ErrExecutionNotFound{ExecutionID: executionID}
	}
	if state.postDone {
		err := state.postErr
		s.mu.Unlock()
		return err
	}
	redact := s.redactorLocked()
	s.mu.Unlock()

	ctx = context.WithoutCancel(ctx)
	if s.manager.cfg.PostProcessTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.manager.cfg.PostProcessTimeout)
		defer cancel()
	}

	sink := newForwardLogger(forward, redact, s.manager.cfg.MaxLogFrames)
	err := s.d.impl.PostProcess(ctx, state.execution.Request, sink)
	err = s.manager.wrapError(s.dp, s.d, redact, err)

	s.mu.Lock()
	if !state.postDone {
		state.postDone = true
		state.postErr = err
	}
	s.mu.Unlock()
	return err
}

func (s *Session) PostProcessDnsKey(ctx context.Context, key DnsKey, forward *slog.Logger) error {
	if err := s.checkOpen(); err != nil {
		return err
	}
	s.mu.Lock()
	executionID, ok := s.dnsKeys[key]
	s.mu.Unlock()
	if !ok {
		return &ErrExecutionNotFound{ExecutionID: fmt.Sprintf("dns:%s", key.Domain)}
	}
	return s.PostProcess(ctx, executionID, forward)
}

func (s *Session) Close() {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
	s.closeOnce.Do(func() {
		s.d.client.Kill()
	})
}

func (m *Manager) wrapError(dp *DiscoveredPlugin, d *dispensed, redact secretRedactor, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	tail := stderrTail(d.stderr, redact)
	safeErr := &redactedPluginError{error: err, message: redact(err.Error())}
	if d.client.Exited() {
		return &ErrPluginCrashed{ProviderType: dp.Manifest.ProviderType, StderrTail: tail, Inner: safeErr}
	}
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Canceled:
			return fmt.Errorf("plugin %q: %w", dp.Manifest.ProviderType, context.Canceled)
		case codes.DeadlineExceeded:
			return fmt.Errorf("plugin %q: %w", dp.Manifest.ProviderType, context.DeadlineExceeded)
		case codes.Unavailable:
			return &ErrPluginCrashed{ProviderType: dp.Manifest.ProviderType, StderrTail: tail, Inner: safeErr}
		}
		return fmt.Errorf("plugin %q: %s", dp.Manifest.ProviderType, redact(st.Message()))
	}
	return fmt.Errorf("plugin %q: %w", dp.Manifest.ProviderType, safeErr)
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

func (s *Session) collectSecretsLocked(req *Request) {
	collect := func(jsonStr string) {
		var m map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
			return
		}
		for _, v := range flattenValues(m) {
			if str, ok := v.(string); ok && len(str) >= 4 {
				s.redactions = append(s.redactions, str)
			}
		}
	}
	collect(req.AccessConfigJSON)
	collect(req.ExtendedConfigJSON)
	if req.Deploy != nil {
		collectPEM(&s.redactions, req.Deploy.CertificatePEM)
		collectPEM(&s.redactions, req.Deploy.PrivateKeyPEM)
	}
	if req.Dns != nil {
		for _, secret := range []string{req.Dns.Token, req.Dns.KeyAuth} {
			if len(secret) >= 4 {
				s.redactions = append(s.redactions, secret)
			}
		}
	}
}

func (s *Session) redactorLocked() secretRedactor {
	secrets := s.redactions
	return func(str string) string {
		for _, sec := range secrets {
			str = strings.ReplaceAll(str, sec, "[REDACTED]")
		}
		return str
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
