package pluginhost

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/certimate-go/certimate/pkg/plugin"
)

func validDetailSHA(value string) bool {
	digest, err := hex.DecodeString(value)
	return err == nil && len(digest) == sha256.Size && value == strings.ToLower(value)
}

func (s *MarketService) resolveMarketDetail(ctx context.Context, summary *plugin.MarketManifest) (*plugin.MarketManifest, error) {
	ref := summary.Detail
	if ref == nil {
		return nil, fmt.Errorf("market: plugin %q has no release", summary.ProviderType)
	}
	if !validDetailSHA(ref.SHA256) {
		return nil, fmt.Errorf("market: invalid detail source for %q", summary.ProviderType)
	}

	s.mu.RLock()
	detail := s.details[ref.SHA256]
	s.mu.RUnlock()
	uncached := detail == nil
	if uncached {
		var err error
		detail, err = s.fetchMarketDetail(ctx, ref.SHA256)
		if err != nil {
			return nil, fmt.Errorf("market: resolve detail for %q: %w", summary.ProviderType, err)
		}
	}

	if detail.ProviderType != summary.ProviderType || detail.Version != summary.Version ||
		detail.ProtocolVersion != summary.ProtocolVersion || detail.Release.Repo != ref.Repo ||
		detail.MinCoreVersion != summary.MinCoreVersion || detail.MaxCoreVersion != summary.MaxCoreVersion {
		return nil, fmt.Errorf("market: detail does not match catalog entry for %q", summary.ProviderType)
	}

	if uncached {
		s.mu.Lock()
		if len(s.details) >= 128 {
			clear(s.details)
		}
		s.details[ref.SHA256] = detail
		s.mu.Unlock()
	}
	return detail, nil
}

func (s *MarketService) fetchMarketDetail(ctx context.Context, digest string) (*plugin.MarketManifest, error) {
	base, err := url.Parse(s.indexURL)
	if err != nil {
		return nil, err
	}
	target := base.ResolveReference(&url.URL{Path: "market/details/" + digest + ".json"})
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("detail fetch returned status %d", resp.StatusCode)
	}
	const maxDetailBytes = 64 << 10
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxDetailBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxDetailBytes {
		return nil, fmt.Errorf("detail exceeds size limit")
	}
	sum := sha256.Sum256(body)
	if hex.EncodeToString(sum[:]) != digest {
		return nil, fmt.Errorf("detail checksum mismatch")
	}
	return plugin.ParseMarketManifest(body)
}
