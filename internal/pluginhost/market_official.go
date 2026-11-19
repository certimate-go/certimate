package pluginhost

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/certimate-go/certimate/internal/repository"
	"github.com/certimate-go/certimate/pkg/plugin"
)

func builtInOfficialMarketTrust() plugin.OfficialMarketTrust {
	return plugin.OfficialMarketTrust{Epoch: 1}
}

func (s *marketCatalog) isOfficial() bool {
	return s.source == plugin.OfficialMarketSource()
}

func (s *marketCatalog) cacheKey() string {
	if s.isOfficial() {
		return "official:signed-index"
	}
	return s.indexURL
}

func (s *marketCatalog) loadVerifiedOfficial(info repository.PluginMarketSync) {
	if info.OfficialState == nil || info.OfficialState.Epoch < s.officialTrust.Epoch {
		return
	}
	s.syncInfo.OfficialState = info.OfficialState
	snapshot, err := plugin.VerifyOfficialMarketIndex(info.OfficialIndex, s.officialTrust, *info.OfficialState, time.Now())
	if (err != nil && !errors.Is(err, plugin.ErrOfficialIndexExpired)) || snapshot == nil || snapshot.State != *info.OfficialState {
		return
	}
	info.OfficialExpiresAt = snapshot.ExpiresAt
	s.cache, s.syncInfo = snapshot.Plugins, info
	s.cachedAt = time.Time{}
}

func (s *marketCatalog) canInstallFromCatalog() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.canInstallFromCatalogLocked()
}

func (s *marketCatalog) canInstallFromCatalogLocked() bool {
	return !s.isOfficial() || (!s.retired && s.refreshErr == nil && !s.cachedAt.IsZero() &&
		s.syncInfo.OfficialMirror == s.indexURL && s.syncInfo.OfficialState != nil && time.Now().Before(s.syncInfo.OfficialExpiresAt))
}

func (s *marketCatalog) stillApproves(manifest *plugin.MarketManifest) bool {
	if !s.isOfficial() {
		return true
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.canInstallFromCatalogLocked() || manifest.Detail == nil {
		return false
	}
	for _, row := range s.cache {
		if row.ProviderType == manifest.ProviderType {
			return row.Version == manifest.Version && row.Detail.SHA256 == manifest.Detail.SHA256
		}
	}
	return false
}

func (s *marketCatalog) officialArtifactURL(release *plugin.OfficialMarketRelease, artifact plugin.OfficialMarketArtifact) (string, error) {
	base, ok := strings.CutSuffix(s.indexURL, plugin.OfficialMarketIndexPath)
	if !ok || !strings.HasPrefix(base, "https://") {
		return "", fmt.Errorf("market: invalid official CDN index URL")
	}
	return base + "/gh/" + release.Repository + "@" + release.ReleaseCommit + "/" + artifact.Path, nil
}

func (s *marketCatalog) officialIndexFor(ctx context.Context, info *repository.PluginMarketSync, body []byte) ([]*plugin.MarketManifest, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	floor := plugin.OfficialIndexState{}
	if info.OfficialState != nil {
		floor = *info.OfficialState
	}
	snapshot, err := plugin.VerifyOfficialMarketIndex(body, s.officialTrust, floor, time.Now())
	if err != nil {
		return nil, err
	}
	info.OfficialIndex, info.OfficialState, info.OfficialMirror = body, &snapshot.State, s.indexURL
	info.OfficialExpiresAt = snapshot.ExpiresAt
	return snapshot.Plugins, nil
}

func copyOfficialTrust(trust plugin.OfficialMarketTrust) plugin.OfficialMarketTrust {
	keys := maps.Clone(trust.Keys)
	for id, publicKey := range keys {
		keys[id] = slices.Clone(publicKey)
	}
	return plugin.OfficialMarketTrust{Epoch: trust.Epoch, Keys: keys}
}
