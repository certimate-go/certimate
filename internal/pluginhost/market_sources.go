package pluginhost

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/repository"
	"github.com/certimate-go/certimate/pkg/plugin"
)

type marketSourceStore interface {
	List(context.Context) ([]plugin.MarketSource, error)
	Add(context.Context, plugin.MarketSource) error
	Remove(context.Context, string) error
}

type MarketSourceStatus struct {
	plugin.MarketSource
	Official   bool       `json:"official"`
	SyncedAt   *time.Time `json:"syncedAt,omitempty"`
	Error      string     `json:"error,omitempty"`
	Stale      bool       `json:"stale"`
	Refreshing bool       `json:"refreshing"`
}

var (
	ErrInvalidMarket     = domain.NewError(400, "market: expected a public GitHub repository URL")
	ErrOfficialMarket    = domain.NewError(409, "market: the official market cannot be changed")
	ErrDuplicateMarket   = domain.NewError(409, "market: repository is already registered")
	ErrMarketBusy        = domain.NewError(409, "market: source has an active installation or update")
	ErrMarketUnavailable = domain.NewError(404, "market: source is not configured")
)

func (s *MarketService) loadSources(ctx context.Context) error {
	if s.sources != nil {
		return nil
	}
	var sources []plugin.MarketSource
	if s.sourceStore != nil {
		var err error
		sources, err = s.sourceStore.List(ctx)
		if err != nil {
			return err
		}
	}
	s.sources = map[string]*marketCatalog{s.source.ID: s.marketCatalog}
	for _, source := range sources {
		canonical, err := plugin.ParseMarketSource("https://github.com/" + source.Repository)
		if err != nil || canonical != source || source.ID == s.source.ID {
			s.sources = nil
			return fmt.Errorf("market: invalid stored source %q", source.ID)
		}
		s.sources[source.ID] = s.customCatalog(source)
	}
	return nil
}

func (s *MarketService) customCatalog(source plugin.MarketSource) *marketCatalog {
	return &marketCatalog{
		source: source, indexURL: source.IndexURL(),
		details: make(map[string]*plugin.MarketManifest),
		store:   s.store, pluginDir: s.pluginDir, cacheTTL: s.cacheTTL,
		downloadMirror: s.downloadMirror, stallTimeout: s.stallTimeout,
		httpClient: s.httpClient, downloadClient: s.downloadClient, logger: s.logger,
	}
}

func (s *MarketService) sourceCatalogs(ctx context.Context) ([]*marketCatalog, error) {
	s.sourcesMu.Lock()
	defer s.sourcesMu.Unlock()
	if err := s.loadSources(ctx); err != nil {
		return nil, err
	}
	result := make([]*marketCatalog, 0, len(s.sources))
	for _, catalog := range s.sources {
		result = append(result, catalog)
	}
	slices.SortFunc(result, func(a, b *marketCatalog) int {
		if a.source.ID == s.source.ID {
			if b.source.ID == s.source.ID {
				return 0
			}
			return -1
		}
		if b.source.ID == s.source.ID {
			return 1
		}
		return strings.Compare(a.source.Repository, b.source.Repository)
	})
	return result, nil
}

func (s *MarketService) leaseSource(ctx context.Context, id string) (*marketCatalog, func(), error) {
	s.sourcesMu.Lock()
	defer s.sourcesMu.Unlock()
	if err := s.loadSources(ctx); err != nil {
		return nil, nil, err
	}
	catalog := s.sources[id]
	if catalog == nil {
		return nil, nil, ErrMarketUnavailable
	}
	s.leases[id]++
	return catalog, func() {
		s.sourcesMu.Lock()
		defer s.sourcesMu.Unlock()
		s.leases[id]--
		if s.leases[id] == 0 {
			delete(s.leases, id)
		}
	}, nil
}

func (s *MarketService) ListSources(ctx context.Context) ([]MarketSourceStatus, error) {
	catalogs, err := s.sourceCatalogs(ctx)
	if err != nil {
		return nil, err
	}
	statuses := make([]MarketSourceStatus, 0, len(catalogs))
	for _, catalog := range catalogs {
		if status, ok := catalog.sourceStatus(); ok {
			statuses = append(statuses, status)
		}
	}
	return statuses, nil
}

func (s *MarketService) AddSource(ctx context.Context, repositoryURL string) (MarketSourceStatus, error) {
	source, err := plugin.ParseMarketSource(repositoryURL)
	if err != nil {
		return MarketSourceStatus{}, ErrInvalidMarket
	}
	if source.ID == s.source.ID {
		return MarketSourceStatus{}, ErrOfficialMarket
	}
	s.sourcesMu.Lock()
	err = s.checkSourceAddition(ctx, source.ID)
	s.sourcesMu.Unlock()
	if err != nil {
		return MarketSourceStatus{}, err
	}

	catalog := s.customCatalog(source)
	var info repository.PluginMarketSync
	entries, err := catalog.fetchMarketIndex(ctx, &info)
	if err != nil {
		return MarketSourceStatus{}, domain.NewError(http.StatusBadGateway, err.Error())
	}
	info.SyncedAt = time.Now()

	s.sourcesMu.Lock()
	defer s.sourcesMu.Unlock()
	if err := s.checkSourceAddition(ctx, source.ID); err != nil {
		return MarketSourceStatus{}, err
	}
	if catalog.store != nil {
		if err := catalog.store.Save(ctx, catalog.indexURL, entries, info); err != nil {
			return MarketSourceStatus{}, err
		}
	}
	if err := s.sourceStore.Add(ctx, source); err != nil {
		return MarketSourceStatus{}, err
	}
	catalog.cache, catalog.syncInfo, catalog.cachedAt = entries, info, info.SyncedAt
	catalog.loaded = true
	s.sources[source.ID] = catalog
	return MarketSourceStatus{MarketSource: source, SyncedAt: &info.SyncedAt}, nil
}

func (s *MarketService) checkSourceAddition(ctx context.Context, id string) error {
	if err := s.loadSources(ctx); err != nil {
		return err
	}
	if s.sources[id] != nil {
		return ErrDuplicateMarket
	}
	if s.sourceStore == nil {
		return fmt.Errorf("market: source storage is unavailable")
	}
	return nil
}

func (s *MarketService) RemoveSource(ctx context.Context, id string) error {
	if id == s.source.ID {
		return ErrOfficialMarket
	}
	s.sourcesMu.Lock()
	defer s.sourcesMu.Unlock()
	if err := s.loadSources(ctx); err != nil {
		return err
	}
	catalog := s.sources[id]
	if catalog == nil {
		return ErrMarketUnavailable
	}
	if s.leases[id] > 0 {
		return ErrMarketBusy
	}
	catalog.mu.Lock()
	defer catalog.mu.Unlock()
	if err := s.sourceStore.Remove(ctx, id); err != nil {
		return err
	}
	catalog.retired = true
	delete(s.sources, id)
	s.jobs.removeSource(id)
	return nil
}

func (s *MarketService) RefreshSource(ctx context.Context, id string) (MarketSourceStatus, error) {
	s.sourcesMu.Lock()
	if err := s.loadSources(ctx); err != nil {
		s.sourcesMu.Unlock()
		return MarketSourceStatus{}, err
	}
	catalog := s.sources[id]
	s.sourcesMu.Unlock()
	if catalog == nil {
		return MarketSourceStatus{}, ErrMarketUnavailable
	}
	_, _ = catalog.marketSnapshot(ctx, false, true)
	status, ok := catalog.sourceStatus()
	if !ok {
		return MarketSourceStatus{}, ErrMarketUnavailable
	}
	return status, nil
}

func (s *marketCatalog) sourceStatus() (MarketSourceStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	status := MarketSourceStatus{MarketSource: s.source, Official: s.source == plugin.OfficialMarketSource(), Refreshing: s.refreshing != nil}
	if !s.cachedAt.IsZero() {
		syncedAt := s.cachedAt
		status.SyncedAt = &syncedAt
	}
	if s.refreshErr != nil {
		status.Error = s.refreshErr.Error()
		status.Stale = s.cache != nil
	}
	return status, !s.retired
}

func catalogRedirect(req *http.Request, via []*http.Request) error {
	first := via[0].URL
	if len(via) >= 10 || req.URL.User != nil || req.URL.Scheme != first.Scheme || !strings.EqualFold(req.URL.Host, first.Host) {
		return fmt.Errorf("market: unexpected catalog redirect")
	}
	return nil
}

func releaseRedirect(mirror string) func(*http.Request, []*http.Request) error {
	if mirror != "" {
		return nil
	}
	return func(req *http.Request, via []*http.Request) error {
		if len(via) < 10 && req.URL.User == nil && req.URL.Scheme == "https" {
			switch strings.ToLower(req.URL.Host) {
			case "github.com", "release-assets.githubusercontent.com", "objects.githubusercontent.com":
				return nil
			}
		}
		return fmt.Errorf("market: unexpected release redirect")
	}
}
