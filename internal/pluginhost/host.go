package pluginhost

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/providerschema"
	"github.com/certimate-go/certimate/pkg/plugin"
)

func ScanAndRegister(ctx context.Context, cfg plugin.PluginConfig, logger *slog.Logger) (*Catalog, []error) {
	if logger == nil {
		logger = slog.Default()
	}
	catalog := NewCatalog()

	if cfg.PluginDir == "" {
		return catalog, nil
	}
	if info, err := os.Stat(cfg.PluginDir); err != nil || !info.IsDir() {
		logger.Debug("plugin dir not present, skipping plugin scan", slog.String("dir", cfg.PluginDir))
		return catalog, nil
	}

	manager := plugin.NewManager(cfg, logger)

	discovered, failures := plugin.Discover(ctx, cfg)

	var errs []error
	for _, f := range failures {
		logger.Warn("plugin discovery skipped entry",
			slog.String("dir", f.Dir),
			slog.Any("error", f.Err))
		errs = append(errs, f)
	}

	var prepared []*registeredPlugin
	for _, dp := range discovered {
		p, err := preparePlugin(ctx, manager, dp, logger, "")
		if err != nil {
			logger.Error("plugin registration failed",
				slog.String("provider", dp.Manifest.ProviderType),
				slog.Any("error", err))
			errs = append(errs, fmt.Errorf("plugin %q: %w", dp.Manifest.ProviderType, err))
			continue
		}
		prepared = append(prepared, p)
	}
	ordered, resolveErrs := catalog.resolveAccessOwnership(prepared)
	for _, err := range resolveErrs {
		logger.Error("plugin access ownership conflict", slog.Any("error", err))
		errs = append(errs, err)
	}
	for _, p := range ordered {
		if err := catalog.publish(p); err != nil {
			logger.Error("plugin registration failed",
				slog.String("provider", p.source.Manifest.ProviderType),
				slog.Any("error", err))
			errs = append(errs, fmt.Errorf("plugin %q: %w", p.source.Manifest.ProviderType, err))
		}
	}
	return catalog, errs
}

func registerOne(ctx context.Context, manager *plugin.Manager, dp *plugin.DiscoveredPlugin, catalog *Catalog, logger *slog.Logger, fingerprint string) error {
	prepared, err := preparePlugin(ctx, manager, dp, logger, fingerprint)
	if err != nil {
		return err
	}
	ordered, resolveErrs := catalog.resolveAccessOwnership([]*registeredPlugin{prepared})
	if len(resolveErrs) > 0 {
		os.RemoveAll(prepared.runtimeDir)
		return resolveErrs[0]
	}
	if err := catalog.publish(ordered[0]); err != nil {
		os.RemoveAll(prepared.runtimeDir)
		return err
	}
	return nil
}

func configSchemaCategory(capability string) providerschema.Category {
	switch capability {
	case plugin.CapabilityDns:
		return providerschema.CategoryApply
	case plugin.CapabilityNotify:
		return providerschema.CategoryNotify
	default:
		return providerschema.CategoryDeploy
	}
}

func defaultAccessUsages(capability string) []string {
	switch capability {
	case plugin.CapabilityDns:
		return []string{"dns"}
	case plugin.CapabilityNotify:
		return []string{"notification"}
	default:
		return []string{"hosting"}
	}
}

func isCoreOwnedAccessType(accessType string) bool {
	if domain.IsBuiltInAccessProviderType(domain.AccessProviderType(accessType)) {
		return true
	}
	if schema, err := providerschema.Registries.Get(accessType); err == nil && schema.Category == providerschema.CategoryAccess {
		return true
	}
	return false
}

func preparePlugin(ctx context.Context, manager *plugin.Manager, dp *plugin.DiscoveredPlugin, logger *slog.Logger, fingerprint string) (*registeredPlugin, error) {
	if fingerprint == "" {
		var err error
		fingerprint, err = pluginFingerprint(dp)
		if err != nil {
			return nil, err
		}
	}
	runtimeDP, runtimeDir, err := snapshotPlugin(dp, "")
	if err != nil {
		return nil, err
	}
	accepted := false
	defer func() {
		if !accepted {
			os.RemoveAll(runtimeDir)
		}
	}()
	meta, schema, err := manager.Bootstrap(ctx, runtimeDP)
	if errors.Is(err, os.ErrPermission) {
		os.RemoveAll(runtimeDir)
		runtimeDP, runtimeDir, err = snapshotPlugin(dp, filepath.Dir(dp.Dir))
		if err == nil {
			meta, schema, err = manager.Bootstrap(ctx, runtimeDP)
		}
		if err != nil {
			return nil, fmt.Errorf("plugin runtime needs a writable executable directory; configure TMPDIR accordingly: %w", err)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("bootstrap: %w", err)
	}
	if meta == nil || schema == nil {
		return nil, fmt.Errorf("plugin returned empty metadata or schema")
	}
	if meta.ProtocolVersion != plugin.ProtocolVersion {
		return nil, &plugin.ErrPluginIncompatible{ProviderType: meta.ProviderType, Have: meta.ProtocolVersion, Want: plugin.ProtocolVersion}
	}
	if meta.AccessProviderType == "" {
		meta.AccessProviderType = meta.ProviderType
	}
	if meta.ProviderType != dp.Manifest.ProviderType || meta.AccessProviderType != dp.Manifest.AccessProviderType {
		return nil, fmt.Errorf("plugin metadata identity does not match manifest: provider %q/%q, access %q/%q", meta.ProviderType, dp.Manifest.ProviderType, meta.AccessProviderType, dp.Manifest.AccessProviderType)
	}
	capability := dp.Manifest.Capability
	providerType, accessType := meta.ProviderType, meta.AccessProviderType
	var envelopes []*providerschema.Envelope
	addEnvelope := func(name string, raw []byte, category providerschema.Category) error {
		env, err := parseEnvelope(raw)
		if err != nil {
			return err
		}
		if env.SchemaVersion != "" && env.SchemaVersion != providerschema.SchemaVersion {
			return fmt.Errorf("unsupported schema version %q", env.SchemaVersion)
		}
		env.SchemaVersion = providerschema.SchemaVersion
		env.Provider = name
		env.Category = category
		envelopes = append(envelopes, env)
		return nil
	}
	if err := addEnvelope(providerType, schema.DeploySchemaJSON, configSchemaCategory(capability)); err != nil {
		return nil, fmt.Errorf("provider schema: %w", err)
	}
	ownsAccess := false
	if len(schema.AccessSchemaJSON) > 0 {
		if isCoreOwnedAccessType(accessType) {
			return nil, fmt.Errorf("access type %q is owned by core built-ins; plugin %q must reuse it without shipping an access schema", accessType, providerType)
		}
		if err := addEnvelope(accessType, schema.AccessSchemaJSON, providerschema.CategoryAccess); err != nil {
			return nil, fmt.Errorf("access schema: %w", err)
		}
		ownsAccess = true
	}
	after, err := pluginFingerprint(dp)
	if err != nil {
		return nil, err
	}
	if fingerprint != after {
		return nil, fmt.Errorf("plugin changed while loading")
	}
	usages := dp.Manifest.Usages
	if len(usages) == 0 {
		usages = defaultAccessUsages(capability)
	}
	registered := &registeredPlugin{
		source: dp, fingerprint: fingerprint, runtimeDir: runtimeDir,
		capability: capability, accessType: accessType, ownsAccess: ownsAccess,
		envelopes: envelopes,
		entry: &CatalogEntry{
			Source: SourcePlugin, Capability: capability, ProviderType: providerType, AccessProviderType: accessType,
			OwnsAccess: ownsAccess, DeployCategory: meta.DeployCategory, DisplayNameKey: meta.DeployDisplayNameKey,
			AccessDisplayNameKey: meta.AccessDisplayNameKey, Icon: readIcon(dp, logger), I18n: schema.I18n,
			AccessUsages: usages, Priority: dp.Manifest.Priority, Description: dp.Manifest.Description,
		},
	}
	switch capability {
	case plugin.CapabilityDns:
		factory := &pluginDnsChallengerFactory{
			manager: manager, dp: runtimeDP, providerType: providerType,
			logLevel: manager.Config().LogLevel, logger: logger.With(slog.String("plugin", providerType)),
		}
		registered.factory = factory
		registered.newChallenger = factory.new
	case plugin.CapabilityNotify:
		factory := &pluginNotifierFactory{
			manager: manager, dp: runtimeDP, providerType: providerType,
			logLevel: manager.Config().LogLevel, logger: logger.With(slog.String("plugin", providerType)),
		}
		registered.factory = factory
		registered.newNotifier = factory.new
	default:
		factory := &pluginDeployerFactory{
			manager: manager, dp: runtimeDP, providerType: providerType,
			logLevel: manager.Config().LogLevel, logger: logger.With(slog.String("plugin", providerType)),
		}
		registered.factory = factory
		registered.newDeployer = factory.new
	}
	accepted = true
	return registered, nil
}

func parseEnvelope(raw []byte) (*providerschema.Envelope, error) {
	if len(raw) == 0 {
		return &providerschema.Envelope{SchemaVersion: providerschema.SchemaVersion}, nil
	}
	var env providerschema.Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

func readIcon(dp *plugin.DiscoveredPlugin, logger *slog.Logger) string {
	iconName := strings.TrimSpace(dp.Manifest.Icon)
	if iconName == "" {
		return ""
	}
	iconPath := iconName
	if !filepath.IsAbs(iconPath) {
		iconPath = filepath.Join(dp.Dir, iconName)
	}
	data, err := os.ReadFile(iconPath)
	if err != nil {
		logger.Debug("plugin icon unreadable, using placeholder",
			slog.String("provider", dp.Manifest.ProviderType),
			slog.String("icon", iconPath))
		return ""
	}
	mime, ok := mimeForExt(filepath.Ext(iconName))
	if !ok {
		logger.Debug("plugin icon format unsupported, using placeholder",
			slog.String("provider", dp.Manifest.ProviderType),
			slog.String("icon", iconPath))
		return ""
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
}

func mimeForExt(ext string) (string, bool) {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png", true
	case ".jpg", ".jpeg":
		return "image/jpeg", true
	case ".gif":
		return "image/gif", true
	case ".webp":
		return "image/webp", true
	case ".svg":
		return "image/svg+xml", true
	default:
		return "", false
	}
}
