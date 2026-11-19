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

func (s *MarketService) OfficialIcon(ctx context.Context, provider, digest string) ([]byte, string, error) {
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
	if catalog.canInstallFromCatalogLocked() {
		for _, row := range catalog.cache {
			if row.ProviderType == provider && row.Official != nil && row.Official.Icon != nil && row.Official.Icon.SHA256 == digest {
				selected = row
				break
			}
		}
	}
	catalog.mu.RUnlock()
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
	resp, err := catalog.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK || resp.ContentLength > officialIconMaxBytes {
		return nil, "", fmt.Errorf("market: icon unavailable")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, officialIconMaxBytes+1))
	if err != nil || len(body) > officialIconMaxBytes || fmt.Sprintf("%x", sha256.Sum256(body)) != digest {
		return nil, "", fmt.Errorf("market: invalid official icon size or checksum")
	}
	dimensions, format, err := image.DecodeConfig(bytes.NewReader(body))
	if err != nil || format != icon.Format || dimensions.Width < 1 || dimensions.Height < 1 ||
		dimensions.Width > 1024 || dimensions.Height > 1024 {
		return nil, "", fmt.Errorf("market: invalid official icon format or dimensions")
	}
	if _, _, err := image.Decode(bytes.NewReader(body)); err != nil || !catalog.stillApproves(selected) {
		return nil, "", fmt.Errorf("market: invalid or withdrawn official icon")
	}
	return body, format, nil
}
