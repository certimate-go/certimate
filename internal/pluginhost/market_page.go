package pluginhost

import (
	"cmp"
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/certimate-go/certimate/pkg/plugin"
)

type MarketQuery struct {
	Page       int
	PerPage    int
	Keyword    string
	Status     string
	Capability string
	Locale     string
	Refresh    bool
	Source     string
}

type MarketCard struct {
	Source          plugin.MarketSource  `json:"source"`
	InstalledOrigin *plugin.MarketSource `json:"installedOrigin,omitempty"`
	Official        bool                 `json:"official"`
	Listed          bool                 `json:"listed"`
	SourceAvailable bool                 `json:"sourceAvailable"`
	CanInstall      bool                 `json:"canInstall"`
	CanUpdate       bool                 `json:"canUpdate"`
	CanDelete       bool                 `json:"canDelete"`

	ProviderType   string     `json:"provider_type"`
	DisplayNameKey string     `json:"display_name_key"`
	DeployCategory string     `json:"deploy_category"`
	Capability     string     `json:"capability,omitempty"`
	Version        string     `json:"version"`
	Description    string     `json:"description"`
	IconURL        string     `json:"icon_url,omitempty"`
	Status         string     `json:"status"`
	Job            *JobStatus `json:"job,omitempty"`
}

type MarketPage struct {
	Sources       []MarketSourceStatus `json:"sources"`
	Items         []MarketCard         `json:"items"`
	Page          int                  `json:"page"`
	PerPage       int                  `json:"perPage"`
	TotalItems    int                  `json:"totalItems"`
	TotalPages    int                  `json:"totalPages"`
	SyncedAt      *time.Time           `json:"syncedAt,omitempty"`
	RefreshFailed bool                 `json:"refreshFailed"`
	Refreshing    bool                 `json:"refreshing"`
	ActiveJobs    []JobStatus          `json:"activeJobs"`
}

type localInstallation struct {
	manifest plugin.Manifest
	origin   plugin.MarketSource
	version  string
}

func (s *MarketService) localInstallations() (map[string]localInstallation, error) {
	dirs, err := os.ReadDir(s.pluginDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	installed := make(map[string]localInstallation, len(dirs))
	for _, dir := range dirs {
		provider := dir.Name()
		if !dir.IsDir() || plugin.ValidateProviderType(provider) != nil {
			continue
		}
		manifest := plugin.Manifest{ProviderType: provider}
		if body, err := os.ReadFile(filepath.Join(s.pluginDir, provider, "manifest.json")); err == nil {
			if parsed, err := plugin.ParseManifest(body); err == nil && parsed.ProviderType == provider {
				manifest = *parsed
			}
		}
		meta, _ := plugin.ReadMarketMeta(s.pluginDir, provider)
		local := localInstallation{manifest: manifest, origin: meta.Origin(), version: manifest.Version}
		if local.origin.ID != "" && meta.InstalledVersion != "" {
			local.version = meta.InstalledVersion
		}
		installed[provider] = local
	}
	return installed, nil
}

func (s *MarketService) ListMarketPage(ctx context.Context, query MarketQuery) (*MarketPage, error) {
	catalogs, err := s.sourceCatalogs(ctx)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	for _, catalog := range catalogs {
		if query.Source != "" && query.Source != catalog.source.ID {
			continue
		}
		wg.Go(func() { _, _ = catalog.marketSnapshot(ctx, !query.Refresh, query.Refresh) })
	}
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	installed, err := s.localInstallations()
	if err != nil {
		return nil, err
	}
	var registered map[string]*registeredPlugin
	if catalog := GlobalCatalog(); catalog != nil {
		registered = catalog.registered()
	}

	result := &MarketPage{Sources: make([]MarketSourceStatus, 0, len(catalogs)), ActiveJobs: s.jobs.active()}
	keyword := strings.ToLower(strings.TrimSpace(query.Keyword))
	matches := make([]MarketCard, 0)
	include := func(card MarketCard, mm *plugin.MarketManifest, owns bool) {
		if query.Source != "" && card.Source.ID != query.Source {
			return
		}
		if query.Status == "installed" && !owns || query.Status == "not_installed" && owns {
			return
		}
		if query.Capability != "" && query.Capability != "all" && card.Capability != query.Capability {
			return
		}
		var runtimePlugin *registeredPlugin
		if owns {
			runtimePlugin = registered[card.ProviderType]
		}
		if keyword != "" && !marketMatches(mm, runtimePlugin, query.Locale, keyword) {
			return
		}
		if job, ok := s.JobStatus(card.ProviderType); ok && job.Source == card.Source && job.State != JobInstalled {
			card.Job = &job
		}
		matches = append(matches, card)
	}
	catalogPresent := make(map[string]bool, len(catalogs))
	ownerRows := make(map[string]bool, len(installed))
	for _, catalog := range catalogs {
		status, ok := catalog.sourceStatus()
		if !ok {
			continue
		}
		result.Sources = append(result.Sources, status)
		catalogPresent[catalog.source.ID] = true
		if query.Source != "" && query.Source != status.ID {
			continue
		}
		result.RefreshFailed = result.RefreshFailed || status.Error != ""
		result.Refreshing = result.Refreshing || status.Refreshing
		if status.SyncedAt != nil && (result.SyncedAt == nil || status.SyncedAt.After(*result.SyncedAt)) {
			result.SyncedAt = status.SyncedAt
		}
		catalog.mu.RLock()
		manifests := catalog.cache
		catalog.mu.RUnlock()
		for _, mm := range manifests {
			card := newMarketCard(mm, catalog.source)
			card.IconURL = catalog.marketIconURL(mm.IconPath)
			local, exists := installed[mm.ProviderType]
			owns := exists && local.origin == catalog.source
			supported := mm.Detail != nil && slices.Contains(mm.Detail.Platforms, plugin.AssetKey(runtime.GOOS, runtime.GOARCH))
			switch {
			case owns:
				ownerRows[mm.ProviderType] = true
				card.Status = "installed"
				card.CanDelete = true
				card.CanUpdate = supported && plugin.CompareVersions(local.version, mm.Version) < 0
				if card.CanUpdate {
					card.Status = "update_available"
				}
			case exists:
				card.Status = "ownership_conflict"
			case supported:
				card.Status = "not_installed"
				card.CanInstall = true
			default:
				card.Status = "unsupported_platform"
			}
			if exists {
				card.InstalledOrigin = &local.origin
			}
			card.SourceAvailable = true
			card.Listed = true
			include(card, mm, owns)
		}
	}
	for provider, local := range installed {
		if ownerRows[provider] {
			continue
		}
		mm := &plugin.MarketManifest{Manifest: local.manifest}
		card := newMarketCard(mm, local.origin)
		card.Version = local.version
		card.InstalledOrigin = &local.origin
		card.Status = "installed_manual"
		if local.origin.ID != "" {
			card.Status = "installed"
			card.CanDelete = true
			card.SourceAvailable = catalogPresent[local.origin.ID]
		}
		include(card, mm, true)
	}
	slices.SortFunc(matches, func(a, b MarketCard) int {
		return cmp.Or(cmp.Compare(a.ProviderType, b.ProviderType), cmp.Compare(a.Source.Repository, b.Source.Repository))
	})
	result.PerPage = min(max(1, query.PerPage), 60)
	result.TotalItems = len(matches)
	result.TotalPages = (len(matches) + result.PerPage - 1) / result.PerPage
	result.Page = min(max(1, query.Page), max(1, result.TotalPages))
	start := (result.Page - 1) * result.PerPage
	result.Items = matches[start:min(start+result.PerPage, len(matches))]
	return result, nil
}

func newMarketCard(mm *plugin.MarketManifest, source plugin.MarketSource) MarketCard {
	capability := mm.Capability
	if capability == "" {
		capability = plugin.CapabilityDeploy
	}
	return MarketCard{
		Source: source, Official: source == plugin.OfficialMarketSource(),
		ProviderType: mm.ProviderType, DisplayNameKey: mm.DisplayNameKey,
		DeployCategory: mm.DeployCategory, Capability: capability, Version: mm.Version, Description: mm.Description,
	}
}

func marketMatches(mm *plugin.MarketManifest, p *registeredPlugin, locale, keyword string) bool {
	if strings.Contains(strings.ToLower(mm.ProviderType), keyword) || strings.Contains(strings.ToLower(mm.Description), keyword) {
		return true
	}
	if p == nil || p.entry == nil {
		return false
	}
	lang, _, _ := strings.Cut(locale, "-")
	for _, name := range []string{p.entry.I18n[locale][mm.DisplayNameKey], p.entry.I18n[lang][mm.DisplayNameKey], p.entry.I18n["en"][mm.DisplayNameKey]} {
		if strings.Contains(strings.ToLower(name), keyword) {
			return true
		}
	}
	return false
}

func (s *marketCatalog) marketIconURL(iconPath string) string {
	if !fs.ValidPath(iconPath) || iconPath == "." {
		return ""
	}
	if u := jsDelivrIconURL(s.indexURL, iconPath); u != "" {
		return u
	}
	base, err := url.Parse(s.indexURL)
	if err != nil || base.Scheme != "https" && base.Scheme != "http" {
		return ""
	}
	return base.ResolveReference(&url.URL{Path: iconPath}).String()
}

func jsDelivrIconURL(indexURL, iconPath string) string {
	u, err := url.Parse(indexURL)
	if err != nil || strings.ToLower(u.Host) != "raw.githubusercontent.com" {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 3 {
		return ""
	}
	owner, repo, ref := parts[0], parts[1], parts[2]
	if ref == "refs" && len(parts) >= 5 && parts[3] == "heads" {
		ref = parts[4]
	}
	return fmt.Sprintf("https://cdn.jsdelivr.net/gh/%s/%s@%s/%s", owner, repo, ref, iconPath)
}
