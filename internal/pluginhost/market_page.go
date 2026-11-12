package pluginhost

import (
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/certimate-go/certimate/pkg/plugin"
)

type MarketQuery struct {
	Page    int
	PerPage int
	Keyword string
	Status  string
	Locale  string
	Refresh bool
}

type MarketCard struct {
	ProviderType   string     `json:"provider_type"`
	DisplayNameKey string     `json:"display_name_key"`
	DeployCategory string     `json:"deploy_category"`
	Version        string     `json:"version"`
	Description    string     `json:"description"`
	IconURL        string     `json:"icon_url,omitempty"`
	Status         string     `json:"status"`
	Job            *JobStatus `json:"job,omitempty"`
}

type MarketPage struct {
	Items         []MarketCard `json:"items"`
	Page          int          `json:"page"`
	PerPage       int          `json:"perPage"`
	TotalItems    int          `json:"totalItems"`
	TotalPages    int          `json:"totalPages"`
	SyncedAt      *time.Time   `json:"syncedAt,omitempty"`
	RefreshFailed bool         `json:"refreshFailed"`
	Refreshing    bool         `json:"refreshing"`
	ActiveJobs    []JobStatus  `json:"activeJobs"`
}

func (s *MarketService) ListMarketPage(ctx context.Context, query MarketQuery) (*MarketPage, error) {
	if _, err := s.marketSnapshot(ctx, !query.Refresh, query.Refresh); err != nil {
		return nil, err
	}
	s.mu.RLock()
	manifests, syncedAt, failed, refreshing := s.cache, s.cachedAt, s.refreshErr != nil, s.refreshing != nil
	s.mu.RUnlock()

	var registered map[string]*registeredPlugin
	if catalog := GlobalCatalog(); catalog != nil {
		registered = catalog.registered()
	}
	dirs, err := os.ReadDir(s.pluginDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	installed := make(map[string]bool, len(dirs))
	for _, dir := range dirs {
		if dir.IsDir() {
			installed[dir.Name()] = true
		}
	}
	keyword := strings.ToLower(strings.TrimSpace(query.Keyword))
	matches := make([]*plugin.MarketManifest, 0)
	include := func(mm *plugin.MarketManifest) {
		isInstalled := installed[mm.ProviderType]
		if query.Status == "installed" && !isInstalled || query.Status == "not_installed" && isInstalled {
			return
		}
		if keyword != "" && !marketMatches(mm, registered[mm.ProviderType], query.Locale, keyword) {
			return
		}
		matches = append(matches, mm)
	}
	seen := make(map[string]bool, len(manifests))
	for _, mm := range manifests {
		seen[mm.ProviderType] = true
		include(mm)
	}
	for provider, p := range registered {
		if !seen[provider] && installed[provider] {
			include(&plugin.MarketManifest{Manifest: *p.source.Manifest})
		}
	}
	slices.SortFunc(matches, func(a, b *plugin.MarketManifest) int {
		return strings.Compare(a.ProviderType, b.ProviderType)
	})

	perPage := min(max(1, query.PerPage), 60)
	totalPages := (len(matches) + perPage - 1) / perPage
	page := min(max(1, query.Page), max(1, totalPages))
	start := (page - 1) * perPage
	visible := matches[start:min(start+perPage, len(matches))]
	result := &MarketPage{
		Items: make([]MarketCard, 0, len(visible)), Page: page, PerPage: perPage,
		TotalItems: len(matches), TotalPages: totalPages, RefreshFailed: failed, Refreshing: refreshing,
		ActiveJobs: s.jobs.active(),
	}
	if !syncedAt.IsZero() {
		result.SyncedAt = &syncedAt
	}
	for _, mm := range visible {
		status := s.computeStatus(mm)
		if installed[mm.ProviderType] && status == "unsupported_platform" {
			status = "installed_manual"
		}
		card := MarketCard{
			ProviderType: mm.ProviderType, DisplayNameKey: mm.DisplayNameKey,
			DeployCategory: mm.DeployCategory, Version: mm.Version,
			Description: mm.Description, IconURL: s.marketIconURL(mm.IconPath), Status: status,
		}
		if job, ok := s.JobStatus(mm.ProviderType); ok && job.State != JobInstalled {
			card.Job = &job
		}
		result.Items = append(result.Items, card)
	}
	return result, nil
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

func (s *MarketService) marketIconURL(iconPath string) string {
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
