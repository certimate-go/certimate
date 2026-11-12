package pluginhost

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/certimate-go/certimate/internal/repository"
	"github.com/certimate-go/certimate/pkg/plugin"
)

type marketCache interface {
	Load(context.Context, string) ([]*plugin.MarketManifest, repository.PluginMarketSync, error)
	Save(context.Context, string, []*plugin.MarketManifest, repository.PluginMarketSync) error
}

func (s *MarketService) manifests(ctx context.Context) ([]*plugin.MarketManifest, error) {
	return s.marketSnapshot(ctx, false, false)
}

func (s *MarketService) marketSnapshot(ctx context.Context, background, force bool) ([]*plugin.MarketManifest, error) {
	s.mu.Lock()
	if !s.loaded {
		if s.store != nil {
			entries, info, err := s.store.Load(ctx, s.indexURL)
			if err != nil {
				s.logger.Warn("market: load cache", slog.Any("error", err))
			} else if !slices.ContainsFunc(entries, func(entry *plugin.MarketManifest) bool {
				return entry.Release != nil
			}) {
				s.cache, s.syncInfo, s.cachedAt = entries, info, info.SyncedAt
			}
		}
		s.loaded = true
	}
	entries, lastErr := s.cache, s.refreshErr
	if !force && (entries != nil && time.Since(s.cachedAt) < s.cacheTTL || time.Now().Before(s.retryAt)) {
		s.mu.Unlock()
		if entries != nil {
			return entries, nil
		}
		return nil, lastErr
	}
	if pending := s.refreshing; pending != nil {
		s.mu.Unlock()
		if background && entries != nil {
			return entries, nil
		}
		select {
		case <-pending:
			s.mu.RLock()
			defer s.mu.RUnlock()
			if s.cache != nil {
				return s.cache, nil
			}
			return nil, s.refreshErr
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	s.refreshing = make(chan struct{})
	info := s.syncInfo
	s.mu.Unlock()

	if background && entries != nil {
		go s.refreshMarket(context.Background(), info)
		return entries, nil
	}
	return s.refreshMarket(ctx, info)
}

func (s *MarketService) refreshMarket(ctx context.Context, info repository.PluginMarketSync) ([]*plugin.MarketManifest, error) {
	entries, err := s.fetchMarketIndex(ctx, &info)
	if err == nil {
		info.SyncedAt = time.Now()
		if s.store != nil {
			err = s.store.Save(ctx, s.indexURL, entries, info)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	defer func() {
		close(s.refreshing)
		s.refreshing = nil
	}()
	s.refreshErr = err
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			s.retryAt = time.Now().Add(time.Minute)
		}
		s.logger.Warn("market: refresh failed", slog.Any("error", err))
		if s.cache != nil {
			return s.cache, nil
		}
		return nil, err
	}
	if entries != nil {
		s.cache = entries
	}
	s.syncInfo, s.cachedAt = info, info.SyncedAt
	s.retryAt = time.Time{}
	return s.cache, nil
}

func (s *MarketService) fetchMarketIndex(ctx context.Context, info *repository.PluginMarketSync) ([]*plugin.MarketManifest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.indexURL, nil)
	if err != nil {
		return nil, fmt.Errorf("market: create request: %w", err)
	}
	if info.ETag != "" {
		req.Header.Set("If-None-Match", info.ETag)
	}
	if info.LastModified != "" {
		req.Header.Set("If-Modified-Since", info.LastModified)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("market: fetch index: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotModified:
		if !info.SyncedAt.IsZero() {
			return nil, nil
		}
		return nil, errors.New("market: 304 without cached index")
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, fmt.Errorf("market: index not available at %s", s.indexURL)
	case http.StatusForbidden, http.StatusTooManyRequests:
		return nil, fmt.Errorf("market: rate limited (status %d)", resp.StatusCode)
	default:
		return nil, fmt.Errorf("market: index fetch returned status %d", resp.StatusCode)
	}

	const maxIndexBytes = 16 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxIndexBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxIndexBytes {
		return nil, errors.New("market: index exceeds size limit")
	}
	var index struct {
		Plugins []*plugin.MarketManifest `json:"plugins"`
	}
	if err := json.Unmarshal(body, &index); err != nil {
		return nil, fmt.Errorf("market: decode index: %w", err)
	}
	if index.Plugins == nil {
		return nil, errors.New("market: missing plugins array")
	}
	seen := make(map[string]bool, len(index.Plugins))
	for i, manifest := range index.Plugins {
		if manifest == nil {
			return nil, fmt.Errorf("market: null plugin at index %d", i)
		}
		if err := plugin.ValidateProviderType(manifest.ProviderType); err != nil {
			return nil, err
		}
		if seen[manifest.ProviderType] {
			return nil, fmt.Errorf("market: duplicate provider_type %q", manifest.ProviderType)
		}
		seen[manifest.ProviderType] = true
		if manifest.Release != nil {
			return nil, fmt.Errorf("market: unexpected inline release for %q", manifest.ProviderType)
		}
		if detail := manifest.Detail; detail != nil {
			if !validDetailSHA(detail.SHA256) || manifest.Version == "" || plugin.ValidateReleaseRepo(detail.Repo) != nil || len(detail.Platforms) == 0 || slices.Contains(detail.Platforms, "") {
				return nil, fmt.Errorf("market: invalid detail reference for %q", manifest.ProviderType)
			}
		}
	}
	info.ETag = resp.Header.Get("ETag")
	info.LastModified = resp.Header.Get("Last-Modified")
	return index.Plugins, nil
}
