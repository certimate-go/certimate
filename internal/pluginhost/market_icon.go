package pluginhost

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/certimate-go/certimate/pkg/plugin"
)

const officialIconMaxBytes = 256 << 10

func (s *marketCatalog) officialIconURL(manifest *plugin.MarketManifest) string {
	if manifest.Official == nil || manifest.Official.Icon == nil {
		return ""
	}
	revision := sha256.Sum256([]byte(s.indexURL))
	return fmt.Sprintf("/api/plugin/market/official/icon/%s/%s?mirror=%x", manifest.ProviderType, manifest.Official.Icon.SHA256, revision[:8])
}

func (s *MarketService) OfficialIcon(ctx context.Context, provider, digest string) (_ []byte, _ string, err error) {
	var mirrorURL string
	var upstreamStatus int
	defer func() {
		if err == nil {
			return
		}
		attrs := []slog.Attr{slog.String("provider", provider), slog.Any("error", err)}
		if mirrorURL != "" {
			attrs = append(attrs, slog.String("mirror_url", mirrorURL))
		}
		if upstreamStatus != 0 {
			attrs = append(attrs, slog.Int("upstream_status", upstreamStatus))
		}
		s.logger.LogAttrs(ctx, slog.LevelWarn, "market: official icon load failed", attrs...)
	}()

	if plugin.ValidateProviderType(provider) != nil || !validDetailSHA(digest) {
		return nil, "", fmt.Errorf("market: invalid icon reference")
	}
	catalog, release, err := s.leaseSource(ctx, plugin.OfficialMarketSource().ID)
	if err != nil {
		return nil, "", err
	}
	defer release()
	catalog.mu.RLock()
	var selected *plugin.MarketManifest
	verified := catalog.canInstallFromCatalogLocked()
	if verified {
		for _, row := range catalog.cache {
			if row.ProviderType == provider && row.Official != nil && row.Official.Icon != nil && row.Official.Icon.SHA256 == digest {
				selected = row
				break
			}
		}
	}
	catalog.mu.RUnlock()
	if !verified {
		return nil, "", fmt.Errorf("market: official icon requires a current verified catalog")
	}
	if selected == nil {
		return nil, "", fmt.Errorf("market: icon is not in the verified catalog for this mirror")
	}
	icon := selected.Official.Icon
	indexDirectory, ok := strings.CutSuffix(catalog.indexURL, "/index.json")
	if !ok {
		return nil, "", fmt.Errorf("market: invalid official index path")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, indexDirectory+"/market/icons/"+digest+"."+icon.Format, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Accept-Encoding", "identity")
	mirrorURL = req.URL.Redacted()
	resp, err := catalog.httpClient.Do(req)
	if resp != nil {
		upstreamStatus = resp.StatusCode
	}
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("market: icon unavailable")
	}
	if resp.ContentLength > officialIconMaxBytes {
		return nil, "", fmt.Errorf("market: official icon exceeds size limit")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, officialIconMaxBytes+1))
	if err != nil {
		return nil, "", fmt.Errorf("market: read official icon: %w", err)
	}
	if len(body) > officialIconMaxBytes {
		return nil, "", fmt.Errorf("market: official icon exceeds size limit")
	}
	if fmt.Sprintf("%x", sha256.Sum256(body)) != digest {
		return nil, "", fmt.Errorf("market: official icon checksum mismatch")
	}
	dimensions, format, err := image.DecodeConfig(bytes.NewReader(body))
	if err != nil {
		return nil, "", fmt.Errorf("market: read official icon format: %w", err)
	}
	if format != icon.Format || dimensions.Width < 1 || dimensions.Height < 1 ||
		dimensions.Width > 1024 || dimensions.Height > 1024 {
		return nil, "", fmt.Errorf("market: invalid official icon format or dimensions")
	}
	if _, _, err := image.Decode(bytes.NewReader(body)); err != nil {
		return nil, "", fmt.Errorf("market: decode official icon: %w", err)
	}
	if !catalog.stillApproves(selected) {
		return nil, "", fmt.Errorf("market: official icon is no longer in the verified catalog")
	}
	return body, format, nil
}
