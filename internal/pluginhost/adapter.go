package pluginhost

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge/dns01"

	"github.com/certimate-go/certimate/internal/certacme/certifiers"
	"github.com/certimate-go/certimate/internal/certmgmt/deployers"
	"github.com/certimate-go/certimate/internal/certmgmt/purgers"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/notify/notifiers"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/plugin"
)

func processWithFinalization(ctx context.Context, manager *plugin.Manager, dp *plugin.DiscoveredPlugin, req *plugin.Request, logger *slog.Logger) (*plugin.Result, error) {
	session, err := manager.StartSession(ctx, dp, logger)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	res, processErr := session.Process(ctx, req, logger)
	var finalErr error
	if req.ExecutionID != "" {
		finalErr = session.PostProcess(ctx, req.ExecutionID, logger)
	}
	if processErr != nil {
		if finalErr != nil {
			return nil, fmt.Errorf("%w; finalization also failed: %v", processErr, finalErr)
		}
		return nil, processErr
	}
	return res, finalErr
}

type pluginDeployerFactory struct {
	manager      *plugin.Manager
	dp           *plugin.DiscoveredPlugin
	providerType string
	logLevel     string
	logger       *slog.Logger
	mu           sync.Mutex
}

func (f *pluginDeployerFactory) new(options *deployers.ProviderFactoryOptions) (core.Deployer, error) {
	return &pluginDeployer{
		factory:  f,
		access:   options.ProviderAccessConfig,
		extended: options.ProviderExtendedConfig,
		logger:   f.logger,
	}, nil
}

type pluginDeployer struct {
	factory  *pluginDeployerFactory
	access   map[string]any
	extended map[string]any
	logger   *slog.Logger
}

func (d *pluginDeployer) SetLogger(logger *slog.Logger) {
	if logger != nil {
		d.logger = logger
	}
}

func (d *pluginDeployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*core.DeployerDeployResult, error) {
	defer runtime.KeepAlive(d.factory)
	d.factory.mu.Lock()
	defer d.factory.mu.Unlock()

	if d.logger != nil {
		d.logger.Info("deploying via plugin", slog.String("provider", d.factory.providerType))
	}

	req := &plugin.Request{
		Capability:         plugin.CapabilityDeploy,
		LogLevel:           d.factory.logLevel,
		AccessConfigJSON:   marshalJSON(d.access),
		ExtendedConfigJSON: marshalJSON(d.extended),
		Deploy:             &plugin.DeployPayload{CertificatePEM: certPEM, PrivateKeyPEM: privkeyPEM},
	}

	res, err := processWithFinalization(ctx, d.factory.manager, d.factory.dp, req, d.logger)
	if err != nil {
		return nil, err
	}

	result := &core.DeployerDeployResult{}
	if res != nil && res.Deploy != nil && res.Deploy.ExtendedDataJSON != "" {
		var extended map[string]any
		if err := json.Unmarshal([]byte(res.Deploy.ExtendedDataJSON), &extended); err == nil {
			result.ExtendedData = extended
		}
	}
	return result, nil
}

type pluginPurgerFactory struct {
	manager      *plugin.Manager
	dp           *plugin.DiscoveredPlugin
	providerType string
	logLevel     string
	logger       *slog.Logger
	mu           sync.Mutex
}

func (f *pluginPurgerFactory) new(options *purgers.ProviderFactoryOptions) (core.Purger, error) {
	return &pluginPurger{
		factory:  f,
		access:   options.ProviderAccessConfig,
		extended: options.ProviderExtendedConfig,
		logger:   f.logger,
	}, nil
}

type pluginPurger struct {
	factory  *pluginPurgerFactory
	access   map[string]any
	extended map[string]any
	logger   *slog.Logger
}

func (p *pluginPurger) SetLogger(logger *slog.Logger) {
	if logger != nil {
		p.logger = logger
	}
}

func (p *pluginPurger) Purge(ctx context.Context, expiry time.Duration) (*core.PurgerPurgeResult, error) {
	defer runtime.KeepAlive(p.factory)
	p.factory.mu.Lock()
	defer p.factory.mu.Unlock()

	if expiry < 0 {
		return nil, fmt.Errorf("purge expiry must not be negative, got %s", expiry)
	}

	if p.logger != nil {
		p.logger.Info("purging via plugin", slog.String("provider", p.factory.providerType), slog.Duration("expiry", expiry))
	}

	req := &plugin.Request{
		Capability:         plugin.CapabilityPurge,
		LogLevel:           p.factory.logLevel,
		AccessConfigJSON:   marshalJSON(p.access),
		ExtendedConfigJSON: marshalJSON(p.extended),
		Purge:              &plugin.PurgePayload{ExpirySeconds: int64(expiry / time.Second)},
	}

	res, err := processWithFinalization(ctx, p.factory.manager, p.factory.dp, req, p.logger)
	if err != nil {
		return nil, err
	}

	result := &core.PurgerPurgeResult{}
	if res != nil && res.Purge != nil {
		result.Count = res.Purge.Count
		if res.Purge.ExtendedDataJSON != "" {
			var extended map[string]any
			if err := json.Unmarshal([]byte(res.Purge.ExtendedDataJSON), &extended); err == nil {
				result.ExtendedData = extended
			}
		}
	}
	return result, nil
}

type pluginDnsChallengerFactory struct {
	manager  *plugin.Manager
	dp       *plugin.DiscoveredPlugin
	logLevel string
	logger   *slog.Logger
}

func newPluginDnsChallengerFactory(manager *plugin.Manager, dp *plugin.DiscoveredPlugin, logger *slog.Logger) *pluginDnsChallengerFactory {
	return &pluginDnsChallengerFactory{
		manager: manager, dp: dp,
		logLevel: manager.Config().LogLevel, logger: logger.With(slog.String("plugin", dp.Manifest.ProviderType)),
	}
}

func RegisterDNSPlugin(dp *plugin.DiscoveredPlugin, logger *slog.Logger) error {
	if dp.Manifest == nil || dp.Manifest.Capability != plugin.CapabilityDns {
		return fmt.Errorf("invalid dns-01 plugin manifest")
	}
	manager := plugin.NewManager(plugin.PluginConfig{}, logger)
	factory := newPluginDnsChallengerFactory(manager, dp, logger)
	return certifiers.ACMEDns01Registries.Register(domain.ACMEDns01ProviderType(dp.Manifest.ProviderType), factory.new, dnsProviderAliases(dp.Manifest)...)
}

func dnsProviderAliases(manifest *plugin.Manifest) []domain.ACMEDns01ProviderType {
	aliases := make([]domain.ACMEDns01ProviderType, len(manifest.ProviderAliases))
	for i, alias := range manifest.ProviderAliases {
		aliases[i] = domain.ACMEDns01ProviderType(alias)
	}
	return aliases
}

func (f *pluginDnsChallengerFactory) new(options *certifiers.ProviderFactoryOptions) (core.ACMEChallenger, error) {
	ctx, cancel := context.WithTimeout(context.Background(), f.manager.Config().StartTimeout)
	defer cancel()
	session, err := f.manager.StartSession(ctx, f.dp, nil)
	if err != nil {
		return nil, err
	}
	metadata, err := session.GetMetadata(ctx)
	if err != nil {
		session.Close()
		return nil, err
	}
	timeout, interval, err := dnsRuntimeTiming(metadata.RuntimeConfigJSON)
	if err != nil {
		session.Close()
		return nil, err
	}
	extended := make(map[string]any, len(options.ProviderExtendedConfig)+5)
	maps.Copy(extended, options.ProviderExtendedConfig)
	extended["nameservers"] = slices.Clone(options.DnsNameservers)
	extended["disableFollowCNAME"] = options.DisableFollowCNAME
	if _, specified := extended["ttl"]; !specified && options.DnsTTL > 0 {
		extended["ttl"] = options.DnsTTL
	}
	challenger := &pluginDnsChallenger{
		session:         session,
		access:          options.ProviderAccessConfig,
		extended:        extended,
		logLevel:        f.logLevel,
		logger:          f.logger,
		timeout:         timeout,
		pollingInterval: interval,
	}
	runtime.AddCleanup(challenger, func(s *plugin.Session) { s.Close() }, session)
	return challenger, nil
}

type pluginDnsChallenger struct {
	session         *plugin.Session
	access          map[string]any
	extended        map[string]any
	logLevel        string
	logger          *slog.Logger
	timeout         time.Duration
	pollingInterval time.Duration
}

func (c *pluginDnsChallenger) Timeout() (time.Duration, time.Duration) {
	return c.timeout, c.pollingInterval
}

func dnsRuntimeTiming(raw []byte) (timeout, interval time.Duration, err error) {
	var config struct {
		TimeoutSeconds  int64 `json:"propagation_timeout_seconds"`
		IntervalSeconds int64 `json:"propagation_polling_interval_seconds"`
	}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &config); err != nil {
			return 0, 0, fmt.Errorf("decode DNS plugin runtime config: %w", err)
		}
	}
	const maxSeconds = int64((1<<63 - 1) / time.Second)
	if config.TimeoutSeconds < 0 || config.TimeoutSeconds > maxSeconds || config.IntervalSeconds < 0 || config.IntervalSeconds > maxSeconds {
		return 0, 0, fmt.Errorf("DNS plugin propagation timing must be valid nonnegative seconds")
	}
	timeout, interval = dns01.DefaultPropagationTimeout, dns01.DefaultPollingInterval
	if config.TimeoutSeconds > 0 {
		timeout = time.Duration(config.TimeoutSeconds) * time.Second
	}
	if config.IntervalSeconds > 0 {
		interval = time.Duration(config.IntervalSeconds) * time.Second
	}
	return timeout, interval, nil
}

func (c *pluginDnsChallenger) Present(ctx context.Context, name, token, keyAuth string) error {
	defer runtime.KeepAlive(c)
	if c.logger != nil {
		c.logger.Info("presenting dns-01 challenge via plugin", slog.String("domain", name))
	}
	info := dns01.GetChallengeInfo(ctx, strings.TrimPrefix(name, "*."), keyAuth)
	extended := maps.Clone(c.extended)
	extended["recordName"] = info.EffectiveFQDN
	extended["recordValue"] = info.Value
	req := &plugin.Request{
		Capability:         plugin.CapabilityDns,
		LogLevel:           c.logLevel,
		AccessConfigJSON:   marshalJSON(c.access),
		ExtendedConfigJSON: marshalJSON(extended),
		Dns:                &plugin.DnsPayload{Domain: name, Token: token, KeyAuth: keyAuth},
	}
	_, err := c.session.Process(ctx, req, c.logger)
	return err
}

func (c *pluginDnsChallenger) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	defer runtime.KeepAlive(c)
	err := c.session.PostProcessDnsKey(ctx, plugin.DnsKey{Domain: domain, Token: token, KeyAuth: keyAuth}, c.logger)
	if plugin.IsExecutionNotFound(err) {
		return nil
	}
	return err
}

func (c *pluginDnsChallenger) Close() {
	c.session.Close()
}

type pluginNotifierFactory struct {
	manager      *plugin.Manager
	dp           *plugin.DiscoveredPlugin
	providerType string
	logLevel     string
	logger       *slog.Logger
	mu           sync.Mutex
}

func (f *pluginNotifierFactory) new(options *notifiers.ProviderFactoryOptions) (core.Notifier, error) {
	return &pluginNotifier{
		factory:  f,
		access:   options.ProviderAccessConfig,
		extended: options.ProviderExtendedConfig,
		logger:   f.logger,
	}, nil
}

type pluginNotifier struct {
	factory  *pluginNotifierFactory
	access   map[string]any
	extended map[string]any
	logger   *slog.Logger
}

func (n *pluginNotifier) SetLogger(logger *slog.Logger) {
	if logger != nil {
		n.logger = logger
	}
}

func (n *pluginNotifier) Notify(ctx context.Context, subject, message string) (*core.NotifierNotifyResult, error) {
	defer runtime.KeepAlive(n.factory)
	n.factory.mu.Lock()
	defer n.factory.mu.Unlock()

	if n.logger != nil {
		n.logger.Info("notifying via plugin", slog.String("provider", n.factory.providerType))
	}

	req := &plugin.Request{
		Capability:         plugin.CapabilityNotify,
		LogLevel:           n.factory.logLevel,
		AccessConfigJSON:   marshalJSON(n.access),
		ExtendedConfigJSON: marshalJSON(n.extended),
		Notify:             &plugin.NotifyPayload{Subject: subject, Message: message},
	}

	res, err := processWithFinalization(ctx, n.factory.manager, n.factory.dp, req, n.logger)
	if err != nil {
		return nil, err
	}

	result := &core.NotifierNotifyResult{}
	if res != nil && res.Notify != nil && res.Notify.ExtendedDataJSON != "" {
		var extended map[string]any
		if err := json.Unmarshal([]byte(res.Notify.ExtendedDataJSON), &extended); err == nil {
			result.ExtendedData = extended
		}
	}
	return result, nil
}

func marshalJSON(m map[string]any) string {
	if len(m) == 0 {
		return "{}"
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(b)
}
