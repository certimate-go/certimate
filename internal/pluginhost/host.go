package pluginhost

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/certimate-go/certimate/internal/certmgmt/deployers"
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
	for _, dp := range discovered {
		if err := registerOne(ctx, manager, dp, catalog, logger); err != nil {
			logger.Error("plugin registration failed",
				slog.String("provider", dp.Manifest.ProviderType),
				slog.Any("error", err))
			errs = append(errs, fmt.Errorf("plugin %q: %w", dp.Manifest.ProviderType, err))
		}
	}
	return catalog, errs
}

func registerOne(ctx context.Context, manager *plugin.Manager, dp *plugin.DiscoveredPlugin, catalog *Catalog, logger *slog.Logger) error {
	meta, schema, err := manager.Bootstrap(ctx, dp)
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	if meta.ProtocolVersion != plugin.ProtocolVersion {
		return &plugin.ErrPluginIncompatible{
			ProviderType: meta.ProviderType,
			Have:         meta.ProtocolVersion,
			Want:         plugin.ProtocolVersion,
		}
	}

	deployType := meta.ProviderType
	accessType := meta.AccessProviderType
	if accessType == "" {
		accessType = deployType
	}

	factory := &pluginDeployerFactory{
		manager:      manager,
		dp:           dp,
		providerType: deployType,
		logLevel:     manager.Config().LogLevel,
		logger:       logger.With(slog.String("plugin", deployType)),
	}

	if err := deployers.Registries.Register(domain.DeploymentProviderType(deployType), factory.new); err != nil {
		logger.Warn("deployer registry register skipped",
			slog.String("provider", deployType),
			slog.Any("error", err))
	}

	registerAccessEnvelope(deployType, accessType, schema.AccessSchemaJSON, logger)
	registerEnvelope(deployType, schema.DeploySchemaJSON, providerschema.CategoryDeploy, logger)

	usages := dp.Manifest.Usages
	if len(usages) == 0 {
		usages = []string{"hosting"}
	}

	catalog.Add(&CatalogEntry{
		Source:               SourcePlugin,
		ProviderType:         deployType,
		AccessProviderType:   accessType,
		DeployCategory:       meta.DeployCategory,
		DisplayNameKey:       meta.DeployDisplayNameKey,
		AccessDisplayNameKey: meta.AccessDisplayNameKey,
		Icon:                 readIcon(dp, logger),
		I18n:                 schema.I18n,
		AccessUsages:         usages,
		Priority:             dp.Manifest.Priority,
		Description:          dp.Manifest.Description,
	})

	return nil
}

func registerAccessEnvelope(deployType, accessType string, raw []byte, logger *slog.Logger) {
	if len(raw) == 0 {
		logger.Info("plugin reuses an existing access type; no access schema registered",
			slog.String("provider", deployType),
			slog.String("accessType", accessType))
		return
	}
	if accessType == deployType {
		logger.Warn("plugin access type equals deploy type; skipping access schema to avoid collision",
			slog.String("provider", deployType))
		return
	}
	if providerschema.Registries.Has(accessType) {
		logger.Info("plugin reuses a built-in access type; keeping built-in access schema",
			slog.String("provider", deployType),
			slog.String("accessType", accessType))
		return
	}
	if env, ok := providerschema.Envelopes.GetEnvelope(accessType); ok && env != nil {
		logger.Info("plugin shares an access type already registered by another plugin; keeping existing schema",
			slog.String("provider", deployType),
			slog.String("accessType", accessType))
		_ = env
		return
	}
	registerEnvelope(accessType, raw, providerschema.CategoryAccess, logger)
}

func registerEnvelope(providerType string, raw []byte, category providerschema.Category, logger *slog.Logger) {
	env, err := parseEnvelope(raw)
	if err != nil {
		logger.Warn("plugin schema envelope parse failed",
			slog.String("provider", providerType),
			slog.Any("error", err))
		return
	}
	if env.Provider == "" {
		env.Provider = providerType
	}
	env.Category = category
	if err := providerschema.Envelopes.RegisterEnvelope(providerType, env); err != nil {
		logger.Warn("envelope register skipped",
			slog.String("provider", providerType),
			slog.Any("error", err))
	}
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
