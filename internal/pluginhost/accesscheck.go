package pluginhost

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/internal/accesscheck"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/internal/providerschema"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/plugin"
)

type accessCheckState struct {
	supportsCheck     bool
	manager           *plugin.Manager
	dp                *plugin.DiscoveredPlugin
	accessType        string
	ownerProviderType string
	logger            *slog.Logger
}

type pluginAccessChecker struct {
	manager      *plugin.Manager
	dp           *plugin.DiscoveredPlugin
	accessType   string
	providerType string
	logger       *slog.Logger
}

var _ accesscheck.Checker = (*pluginAccessChecker)(nil)

func (c *pluginAccessChecker) ProviderType() string {
	return c.accessType
}

func (c *pluginAccessChecker) Scope() []string {
	return []string{core.AccessCheckItemConnect, core.AccessCheckItemAuth}
}

func (c *pluginAccessChecker) Check(ctx context.Context, config map[string]any) (*core.AccessCheckResult, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", core.ErrAccessCheckConfigInvalid, err)
	}
	session, err := c.manager.StartSession(ctx, c.dp)
	if err != nil {
		return accessCheckPluginFailure(err), nil
	}
	defer session.Close()
	res, err := session.CheckAccess(ctx, &plugin.AccessCheckRequest{AccessConfigJSON: string(data)})
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("plugin access check failed", slog.Any("error", err))
		}
		return accessCheckCallFailure(err), nil
	}
	return accessCheckResultToResp(res), nil
}

func accessCheckResultToResp(res *plugin.AccessCheckResult) *core.AccessCheckResult {
	items := make([]core.AccessCheckItem, 0, len(res.Items))
	for _, item := range res.Items {
		items = append(items, core.AccessCheckItem{Id: item.ID, Status: item.Status})
	}
	diagnostics := res.Diagnostics
	if diagnostics == nil {
		diagnostics = []string{}
	}
	return &core.AccessCheckResult{
		Status:      res.Status,
		ReasonCode:  res.ReasonCode,
		Items:       items,
		Diagnostics: diagnostics,
	}
}

func accessCheckPluginFailure(err error) *core.AccessCheckResult {
	return &core.AccessCheckResult{
		Status:      core.AccessCheckStatusUnavailable,
		ReasonCode:  string(dtos.AccessTestReasonCodePluginUnavailable),
		Diagnostics: []string{err.Error()},
	}
}

func accessCheckCallFailure(err error) *core.AccessCheckResult {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return &core.AccessCheckResult{Status: core.AccessCheckStatusInconclusive, ReasonCode: string(dtos.AccessTestReasonCodeTimeout), Items: []core.AccessCheckItem{}, Diagnostics: []string{}}
	case errors.Is(err, context.Canceled):
		return &core.AccessCheckResult{Status: core.AccessCheckStatusInconclusive, ReasonCode: string(dtos.AccessTestReasonCodeCanceled), Items: []core.AccessCheckItem{}, Diagnostics: []string{}}
	default:
		return accessCheckPluginFailure(err)
	}
}

func reconcileAccessCheckRegistration(p *registeredPlugin) {
	state := p.accessCheck
	if state == nil {
		return
	}
	if !p.ownsAccess {
		if state.supportsCheck && state.logger != nil {
			state.logger.Warn("plugin declares access check capability but does not own its access type; no check registration",
				slog.String("provider", state.ownerProviderType),
				slog.String("accessType", state.accessType))
		}
		return
	}
	registry := accesscheck.DefaultRegistry()
	if !state.supportsCheck {
		if err := registry.Register(&accesscheck.Registration{
			Source:            accesscheck.SourcePlugin,
			OwnerProviderType: state.ownerProviderType,
			UnavailableReason: dtos.AccessTestReasonCodeNoChecker,
		}); err != nil && state.logger != nil {
			state.logger.Warn("access check registration failed", slog.Any("error", err))
		}
		return
	}
	if err := registry.Register(&accesscheck.Registration{
		Source:            accesscheck.SourcePlugin,
		OwnerProviderType: state.ownerProviderType,
		Checker: &pluginAccessChecker{
			manager:      state.manager,
			dp:           state.dp,
			accessType:   state.accessType,
			providerType: state.ownerProviderType,
			logger:       state.logger,
		},
	}); err != nil && state.logger != nil {
		state.logger.Warn("access check registration failed", slog.Any("error", err))
	}
}

func dropAccessCheckRegistration(p *registeredPlugin) {
	if p.accessCheck == nil || !p.ownsAccess {
		return
	}
	accesscheck.DefaultRegistry().Unregister(p.accessCheck.accessType)
}

func registerStaleAccessOwner(catalog *Catalog, accessType, providerType string, logger *slog.Logger) {
	if accessType == "" || providerType == "" || isCoreOwnedAccessType(accessType) {
		return
	}
	if catalog.hasAccessPlugin(accessType) {
		return
	}
	registry := accesscheck.DefaultRegistry()
	if existing, ok := registry.Get(accessType); ok && existing.Source == accesscheck.SourceCore {
		return
	}
	if err := registry.Register(&accesscheck.Registration{
		Source:            accesscheck.SourcePlugin,
		OwnerProviderType: providerType,
		UnavailableReason: dtos.AccessTestReasonCodePluginProtocolOld,
	}); err != nil && logger != nil {
		logger.Debug("stale access owner registration skipped", slog.String("provider", providerType), slog.Any("error", err))
	}
}

func registerIncompatibleAccessOwner(catalog *Catalog, dp *plugin.DiscoveredPlugin, err error, logger *slog.Logger) {
	var incompatible *plugin.ErrPluginIncompatible
	if !errors.As(err, &incompatible) {
		return
	}
	if incompatible.Have != 0 && incompatible.Have > incompatible.Want {
		return
	}
	registerStaleAccessOwner(catalog, dp.Manifest.AccessProviderType, dp.Manifest.ProviderType, logger)
}

func registerStaleDiscoveryFailure(catalog *Catalog, failure plugin.DiscoveryFailure, logger *slog.Logger) {
	var incompatible *plugin.ErrPluginIncompatible
	if !errors.As(failure.Err, &incompatible) || incompatible.Have == 0 || incompatible.Have >= incompatible.Want {
		return
	}
	registerStaleAccessOwner(catalog, failure.AccessProviderType, failure.ProviderType, logger)
}

func (c *Catalog) hasAccessPlugin(accessType string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, p := range c.plugins {
		if p.accessType == accessType {
			return true
		}
	}
	return false
}

func (c *Catalog) knownAccessProvider(providerType string) bool {
	if domain.IsBuiltInAccessProviderType(domain.AccessProviderType(providerType)) {
		return true
	}
	if schema, err := providerschema.Registries.Get(providerType); err == nil && schema.Category == providerschema.CategoryAccess {
		return true
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, p := range c.plugins {
		if p.accessType == providerType || p.source.Manifest.ProviderType == providerType {
			return true
		}
	}
	return false
}

func wireAccessCheckKnownProviders(catalog *Catalog) {
	accesscheck.DefaultRegistry().SetKnownProviderResolver(catalog.knownAccessProvider)
}
