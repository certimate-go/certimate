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
	for _, dp := range discovered {
		if err := registerOne(ctx, manager, dp, catalog, logger, ""); err != nil {
			logger.Error("plugin registration failed",
				slog.String("provider", dp.Manifest.ProviderType),
				slog.Any("error", err))
			errs = append(errs, fmt.Errorf("plugin %q: %w", dp.Manifest.ProviderType, err))
		}
	}
	return catalog, errs
}

func registerOne(ctx context.Context, manager *plugin.Manager, dp *plugin.DiscoveredPlugin, catalog *Catalog, logger *slog.Logger, fingerprint string) error {
	prepared, err := preparePlugin(ctx, manager, dp, logger, fingerprint)
	if err != nil {
		return err
	}
	if err := catalog.publish(prepared); err != nil {
		os.RemoveAll(prepared.runtimeDir)
		return err
	}
	return nil
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
	deployType, accessType := meta.ProviderType, meta.AccessProviderType
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
	if err := addEnvelope(deployType, schema.DeploySchemaJSON, providerschema.CategoryDeploy); err != nil {
		return nil, fmt.Errorf("deploy schema: %w", err)
	}
	builtinAccess, _ := providerschema.Registries.Get(accessType)
	if len(schema.AccessSchemaJSON) > 0 && (builtinAccess == nil || builtinAccess.Category != providerschema.CategoryAccess) {
		if err := addEnvelope(accessType, schema.AccessSchemaJSON, providerschema.CategoryAccess); err != nil {
			return nil, fmt.Errorf("access schema: %w", err)
		}
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
		usages = []string{"hosting"}
	}
	prepared := &registeredPlugin{
		source: dp, fingerprint: fingerprint, runtimeDir: runtimeDir, envelopes: envelopes,
		factory: &pluginDeployerFactory{manager: manager, dp: runtimeDP, providerType: deployType, logLevel: manager.Config().LogLevel, logger: logger.With(slog.String("plugin", deployType))},
		entry:   &CatalogEntry{Source: SourcePlugin, ProviderType: deployType, AccessProviderType: accessType, DeployCategory: meta.DeployCategory, DisplayNameKey: meta.DeployDisplayNameKey, AccessDisplayNameKey: meta.AccessDisplayNameKey, Icon: readIcon(dp, logger), I18n: schema.I18n, AccessUsages: usages, Priority: dp.Manifest.Priority, Description: dp.Manifest.Description},
	}
	accepted = true
	return prepared, nil
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
