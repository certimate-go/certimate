package pluginhost

import (
	"context"
	"crypto/ed25519"
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
	return plugin.OfficialMarketTrust{
		Epoch: 1,
		Keys: map[string]ed25519.PublicKey{
			"certimate-market-2026-11": {
				0x81, 0xec, 0xae, 0x9b, 0xdb, 0xe9, 0x14, 0x48,
				0x49, 0x0e, 0xd8, 0x9d, 0xef, 0x6b, 0x69, 0xd1,
				0x47, 0x42, 0x04, 0x99, 0x6d, 0xe9, 0x59, 0xc6,
				0xe0, 0x68, 0x3c, 0x9f, 0xd9, 0x28, 0x01, 0x9a,
			},
		},
	}
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
	if s.indexURL == plugin.OfficialMarketGitHubIndexURL {
		return plugin.GitHubRawBase + "/" + release.Repository + "/" + release.ReleaseCommit + "/" + artifact.Path, nil
	}
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
