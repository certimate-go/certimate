package pluginhost

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/certimate-go/certimate/pkg/plugin"
)

var ErrMissingProviderType = errors.New("market: providerType is required")

type MarketEntry struct {
	*plugin.MarketManifest
	Status string `json:"status"`
}

type MarketService struct {
	marketRepo string
	indexURL   string
	pluginDir  string
	cache      []MarketEntry
	cachedAt   time.Time
	cacheTTL   time.Duration
	mu         sync.RWMutex
	httpClient *http.Client
	logger     *slog.Logger
}

type MarketConfig struct {
	MarketRepo string
	IndexURL   string
	PluginDir  string
	CacheTTL   time.Duration
	Logger     *slog.Logger
}

func NewMarketService(cfg MarketConfig) *MarketService {
	if cfg.CacheTTL <= 0 {
		cfg.CacheTTL = 5 * time.Minute
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.MarketRepo == "" {
		cfg.MarketRepo = "certimate-go/plugins"
	}
	if cfg.IndexURL == "" {
		cfg.IndexURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/main/index.json", cfg.MarketRepo)
	}
	return &MarketService{
		marketRepo: cfg.MarketRepo,
		indexURL:   cfg.IndexURL,
		pluginDir:  cfg.PluginDir,
		cacheTTL:   cfg.CacheTTL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     cfg.Logger.With(slog.String("component", "market")),
	}
}

func (s *MarketService) ListMarket(ctx context.Context) ([]MarketEntry, error) {
	s.mu.RLock()
	if s.cache != nil && time.Since(s.cachedAt) < s.cacheTTL {
		entries := s.cache
		s.mu.RUnlock()
		s.logger.Debug("market listing served from cache", slog.Int("entries", len(entries)))
		return entries, nil
	}
	s.mu.RUnlock()

	return s.fetchAndCache(ctx)
}

func (s *MarketService) fetchAndCache(ctx context.Context) ([]MarketEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cache != nil && time.Since(s.cachedAt) < s.cacheTTL {
		return s.cache, nil
	}

	entries, err := s.fetchMarketListing(ctx)
	if err != nil {
		if s.cache != nil {
			s.logger.Warn("market fetch failed, returning stale cache", slog.Any("error", err))
			return s.cache, nil
		}
		return nil, err
	}

	s.cache = entries
	s.cachedAt = time.Now()
	s.logger.Info("market listing fetched", slog.Int("entries", len(entries)))
	return entries, nil
}

func (s *MarketService) fetchMarketListing(ctx context.Context) ([]MarketEntry, error) {
	url := s.indexURL
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("market: create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("market: fetch index: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, fmt.Errorf("market: index not available at %s", url)
	case http.StatusForbidden, http.StatusTooManyRequests:
		return nil, fmt.Errorf("market: rate limited (status %d)", resp.StatusCode)
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("market: index fetch returned status %d: %s", resp.StatusCode, string(body))
	}

	var idx struct {
		Plugins []*plugin.MarketManifest `json:"plugins"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&idx); err != nil {
		return nil, fmt.Errorf("market: decode index: %w", err)
	}

	entries := make([]MarketEntry, 0, len(idx.Plugins))
	for _, mm := range idx.Plugins {
		entries = append(entries, MarketEntry{MarketManifest: mm, Status: s.computeStatus(mm)})
	}
	return entries, nil
}

func (s *MarketService) computeStatus(mm *plugin.MarketManifest) string {
	key := plugin.AssetKey(runtime.GOOS, runtime.GOARCH)
	if mm.Release == nil || mm.Release.Assets[key] == "" {
		return "unsupported_platform"
	}

	pluginPath := filepath.Join(s.pluginDir, mm.ProviderType)
	info, err := os.Stat(pluginPath)
	if err != nil || !info.IsDir() {
		return "not_installed"
	}

	meta, err := plugin.ReadMarketMeta(s.pluginDir, mm.ProviderType)
	if err != nil || meta == nil {
		return "installed_manual"
	}

	if plugin.CompareVersions(meta.InstalledVersion, mm.Version) < 0 {
		return "update_available"
	}
	return "installed"
}

func (s *MarketService) GetMarketManifest(ctx context.Context, providerType string) (*plugin.MarketManifest, error) {
	entries, err := s.ListMarket(ctx)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.ProviderType == providerType {
			return e.MarketManifest, nil
		}
	}
	return nil, fmt.Errorf("market: plugin %q not found in market listing", providerType)
}

func (s *MarketService) Install(ctx context.Context, providerType string) (*ReloadResult, error) {
	if err := plugin.ValidateProviderType(providerType); err != nil {
		return nil, err
	}

	targetDir := filepath.Join(s.pluginDir, providerType)
	if _, err := os.Stat(targetDir); err == nil {
		return nil, fmt.Errorf("market: plugin %q is already installed", providerType)
	}

	mm, err := s.GetMarketManifest(ctx, providerType)
	if err != nil {
		return nil, err
	}

	if err := plugin.ValidateReleaseRepo(mm.Release.Repo); err != nil {
		return nil, err
	}

	if err := plugin.ValidateBinaryName(mm.Binary); err != nil {
		return nil, err
	}

	key := plugin.AssetKey(runtime.GOOS, runtime.GOARCH)
	assetName := mm.Release.Assets[key]
	if assetName == "" {
		return nil, fmt.Errorf("market: plugin %q has no binary for %s", providerType, key)
	}

	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s",
		mm.Release.Repo, mm.Release.Tag, assetName)

	tmpDir := filepath.Join(s.pluginDir, ".tmp-"+providerType)
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return nil, fmt.Errorf("market: create temp dir: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			s.logger.Warn("market: failed to clean up temp dir", slog.String("dir", tmpDir), slog.Any("error", err))
		}
	}()

	binaryPath := filepath.Join(tmpDir, mm.Binary)
	if err := s.downloadFile(ctx, downloadURL, binaryPath); err != nil {
		return nil, err
	}

	if err := os.Chmod(binaryPath, 0o755); err != nil {
		return nil, fmt.Errorf("market: chmod binary: %w", err)
	}

	computedSHA256, err := sha256File(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("market: compute sha256: %w", err)
	}

	if expected, ok := mm.Release.Checksums[key]; ok && expected != "" && expected != computedSHA256 {
		return nil, fmt.Errorf("market: plugin %q checksum mismatch for %s: expected %s, got %s", providerType, key, expected, computedSHA256)
	}

	localManifest := *mm
	localManifest.OS = runtime.GOOS
	localManifest.Arch = runtime.GOARCH
	localManifest.SHA256 = computedSHA256

	manifestData, err := json.MarshalIndent(&localManifest.Manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("market: marshal local manifest: %w", err)
	}
	manifestPath := filepath.Join(tmpDir, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestData, 0o644); err != nil {
		return nil, fmt.Errorf("market: write manifest: %w", err)
	}

	meta := plugin.NewMarketMeta("official", mm.Version, mm.Version, mm.Release.Tag)
	if err := plugin.WriteMarketMeta(tmpDir, ".", meta); err != nil {
		return nil, err
	}

	if err := os.Rename(tmpDir, targetDir); err != nil {
		return nil, fmt.Errorf("market: atomic rename: %w", err)
	}

	reloader := GlobalReloader()
	if reloader == nil {
		return nil, fmt.Errorf("market: reloader not initialized")
	}
	return reloader.ReloadNow(ctx), nil
}

func (s *MarketService) Delete(ctx context.Context, providerType string) (*ReloadResult, error) {
	if err := plugin.ValidateProviderType(providerType); err != nil {
		return nil, err
	}

	targetDir := filepath.Join(s.pluginDir, providerType)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("market: plugin %q is not installed", providerType)
	}

	meta, err := plugin.ReadMarketMeta(s.pluginDir, providerType)
	if err != nil || meta == nil {
		return nil, fmt.Errorf("market: plugin %q is not managed by marketplace", providerType)
	}

	if err := os.RemoveAll(targetDir); err != nil {
		return nil, fmt.Errorf("market: remove plugin dir: %w", err)
	}

	reloader := GlobalReloader()
	if reloader == nil {
		return nil, fmt.Errorf("market: reloader not initialized")
	}
	return reloader.ReloadNow(ctx), nil
}

func (s *MarketService) Update(ctx context.Context, providerType string) (*ReloadResult, error) {
	if err := plugin.ValidateProviderType(providerType); err != nil {
		return nil, err
	}

	targetDir := filepath.Join(s.pluginDir, providerType)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("market: plugin %q is not installed", providerType)
	}

	mm, err := s.GetMarketManifest(ctx, providerType)
	if err != nil {
		return nil, err
	}

	meta, err := plugin.ReadMarketMeta(s.pluginDir, providerType)
	if err != nil || meta == nil {
		return nil, fmt.Errorf("market: plugin %q is not managed by marketplace", providerType)
	}

	if plugin.CompareVersions(meta.InstalledVersion, mm.Version) >= 0 {
		return nil, fmt.Errorf("market: plugin %q is already up to date", providerType)
	}

	if err := plugin.ValidateReleaseRepo(mm.Release.Repo); err != nil {
		return nil, err
	}

	if err := plugin.ValidateBinaryName(mm.Binary); err != nil {
		return nil, err
	}

	key := plugin.AssetKey(runtime.GOOS, runtime.GOARCH)
	assetName := mm.Release.Assets[key]
	if assetName == "" {
		return nil, fmt.Errorf("market: plugin %q has no binary for %s", providerType, key)
	}

	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s",
		mm.Release.Repo, mm.Release.Tag, assetName)

	tmpDir := filepath.Join(s.pluginDir, ".tmp-"+providerType)
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return nil, fmt.Errorf("market: create temp dir: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			s.logger.Warn("market: failed to clean up temp dir", slog.String("dir", tmpDir), slog.Any("error", err))
		}
	}()

	binaryPath := filepath.Join(tmpDir, mm.Binary)
	if err := s.downloadFile(ctx, downloadURL, binaryPath); err != nil {
		return nil, err
	}

	if err := os.Chmod(binaryPath, 0o755); err != nil {
		return nil, fmt.Errorf("market: chmod binary: %w", err)
	}

	computedSHA256, err := sha256File(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("market: compute sha256: %w", err)
	}

	if expected, ok := mm.Release.Checksums[key]; ok && expected != "" && expected != computedSHA256 {
		return nil, fmt.Errorf("market: plugin %q checksum mismatch for %s: expected %s, got %s", providerType, key, expected, computedSHA256)
	}

	localManifest := *mm
	localManifest.OS = runtime.GOOS
	localManifest.Arch = runtime.GOARCH
	localManifest.SHA256 = computedSHA256

	manifestData, err := json.MarshalIndent(&localManifest.Manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("market: marshal local manifest: %w", err)
	}
	manifestPath := filepath.Join(tmpDir, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestData, 0o644); err != nil {
		return nil, fmt.Errorf("market: write manifest: %w", err)
	}

	newMeta := plugin.NewMarketMeta("official", mm.Version, mm.Version, mm.Release.Tag)
	if err := plugin.WriteMarketMeta(tmpDir, ".", newMeta); err != nil {
		return nil, err
	}

	backupDir := filepath.Join(s.pluginDir, ".bak-"+providerType)
	_ = os.RemoveAll(backupDir)
	if err := os.Rename(targetDir, backupDir); err != nil {
		return nil, fmt.Errorf("market: backup old plugin dir: %w", err)
	}
	if err := os.Rename(tmpDir, targetDir); err != nil {
		if rerr := os.Rename(backupDir, targetDir); rerr != nil {
			return nil, fmt.Errorf("market: install rename failed (%v) and rollback failed (%v)", err, rerr)
		}
		return nil, fmt.Errorf("market: install rename: %w", err)
	}
	if err := os.RemoveAll(backupDir); err != nil {
		s.logger.Warn("market: failed to clean up backup dir", slog.String("dir", backupDir), slog.Any("error", err))
	}

	reloader := GlobalReloader()
	if reloader == nil {
		return nil, fmt.Errorf("market: reloader not initialized")
	}
	return reloader.ReloadNow(ctx), nil
}

func (s *MarketService) downloadFile(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("market: create download request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("market: download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("market: download returned status %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("market: create dest file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("market: write download: %w", err)
	}
	return nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
