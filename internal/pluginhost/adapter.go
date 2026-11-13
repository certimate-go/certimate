package pluginhost

import (
	"context"
	"encoding/json"
	"log/slog"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"

	"github.com/certimate-go/certimate/internal/certacme/certifiers"
	"github.com/certimate-go/certimate/internal/certmgmt/deployers"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/notify/notifiers"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/plugin"
)

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

	req := &plugin.DeployRequest{
		LogLevel:           d.factory.logLevel,
		AccessConfigJSON:   marshalJSON(d.access),
		ExtendedConfigJSON: marshalJSON(d.extended),
		CertificatePEM:     certPEM,
		PrivateKeyPEM:      privkeyPEM,
	}

	res, err := d.factory.manager.Deploy(ctx, d.factory.dp, req, d.logger)
	if err != nil {
		return nil, err
	}

	result := &core.DeployerDeployResult{}
	if res != nil && res.ExtendedDataJSON != "" {
		var extended map[string]any
		if err := json.Unmarshal([]byte(res.ExtendedDataJSON), &extended); err == nil {
			result.ExtendedData = extended
		}
	}
	return result, nil
}

type pluginDnsChallengerFactory struct {
	manager      *plugin.Manager
	dp           *plugin.DiscoveredPlugin
	providerType string
	logLevel     string
	logger       *slog.Logger
}

func (f *pluginDnsChallengerFactory) new(options *certifiers.ProviderFactoryOptions) (core.ACMEChallenger, error) {
	held, err := f.manager.StartDnsProvider(context.Background(), f.dp)
	if err != nil {
		return nil, err
	}
	challenger := &pluginDnsChallenger{
		held:               held,
		providerType:       f.providerType,
		nameservers:        slices.Clone(options.DnsNameservers),
		access:             options.ProviderAccessConfig,
		extended:           options.ProviderExtendedConfig,
		logLevel:           f.logLevel,
		propagationTimeout: options.DnsPropagationTimeout,
		ttl:                options.DnsTTL,
		logger:             f.logger,
	}
	runtime.AddCleanup(challenger, func(h *plugin.HeldDnsProvider) { h.Close() }, held)
	return challenger, nil
}

type pluginDnsChallenger struct {
	held               *plugin.HeldDnsProvider
	providerType       string
	nameservers        []string
	access             map[string]any
	extended           map[string]any
	logLevel           string
	propagationTimeout int
	ttl                int
	logger             *slog.Logger
}

func (c *pluginDnsChallenger) Present(ctx context.Context, name, token, keyAuth string) error {
	defer runtime.KeepAlive(c)
	if c.logger != nil {
		c.logger.Info("presenting dns-01 challenge via plugin", slog.String("domain", name))
	}
	extended := c.extended
	if c.providerType == string(domain.ACMEDns01ProviderTypeCMCCCloudDNS) {
		info := dns01.GetChallengeInfo(ctx, strings.TrimPrefix(name, "*."), keyAuth)
		extended = map[string]any{
			"recordName":  info.EffectiveFQDN,
			"recordValue": info.Value,
			"nameservers": c.nameservers,
			"ttl":         c.ttl,
		}
	}
	req := &plugin.DnsRequest{
		LogLevel:           c.logLevel,
		AccessConfigJSON:   marshalJSON(c.access),
		ExtendedConfigJSON: marshalJSON(extended),
		Domain:             name,
		Token:              token,
		KeyAuth:            keyAuth,
	}
	return c.held.Present(ctx, req)
}

func (c *pluginDnsChallenger) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	defer runtime.KeepAlive(c)
	req := &plugin.DnsRequest{
		LogLevel:           c.logLevel,
		AccessConfigJSON:   marshalJSON(c.access),
		ExtendedConfigJSON: marshalJSON(c.extended),
		Domain:             domain,
		Token:              token,
		KeyAuth:            keyAuth,
	}
	return c.held.CleanUp(ctx, req)
}

func (c *pluginDnsChallenger) Close() {
	c.held.Close()
}

func (c *pluginDnsChallenger) Timeout() (time.Duration, time.Duration) {
	timeout := dns01.DefaultPropagationTimeout
	interval := dns01.DefaultPollingInterval
	if c.providerType == string(domain.ACMEDns01ProviderTypeCMCCCloudDNS) {
		timeout = env.GetOrDefaultSecond("CMCCCLOUD_PROPAGATION_TIMEOUT", 10*time.Minute)
		interval = env.GetOrDefaultSecond("CMCCCLOUD_POLLING_INTERVAL", interval)
	} else if c.ttl > 0 {
		interval = time.Duration(c.ttl) * time.Second
	}
	if c.propagationTimeout > 0 {
		timeout = time.Duration(c.propagationTimeout) * time.Second
	}
	return timeout, interval
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

	req := &plugin.NotifyRequest{
		LogLevel:           n.factory.logLevel,
		AccessConfigJSON:   marshalJSON(n.access),
		ExtendedConfigJSON: marshalJSON(n.extended),
		Subject:            subject,
		Message:            message,
	}

	res, err := n.factory.manager.Notify(ctx, n.factory.dp, req)
	if err != nil {
		return nil, err
	}

	result := &core.NotifierNotifyResult{}
	if res != nil && res.ExtendedDataJSON != "" {
		var extended map[string]any
		if err := json.Unmarshal([]byte(res.ExtendedDataJSON), &extended); err == nil {
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
