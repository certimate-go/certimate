package pluginhost

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/certimate-go/certimate/internal/certmgmt/deployers"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/providerschema"
	"github.com/certimate-go/certimate/pkg/plugin"
)

type Reloader struct {
	cfg     plugin.PluginConfig
	catalog *Catalog
	manager *plugin.Manager
	state   map[string]*plugin.DiscoveredPlugin
	mu      sync.Mutex
	logger  *slog.Logger
}

func NewReloader(cfg plugin.PluginConfig, catalog *Catalog, logger *slog.Logger) *Reloader {
	if logger == nil {
		logger = slog.Default()
	}
	return &Reloader{
		cfg:     cfg,
		catalog: catalog,
		manager: plugin.NewManager(cfg, logger),
		state:   make(map[string]*plugin.DiscoveredPlugin),
		logger:  logger.With(slog.String("component", "pluginreloader")),
	}
}

func (r *Reloader) InitFromCatalog() {
	dps, _ := plugin.Discover(context.Background(), r.cfg)
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, dp := range dps {
		r.state[dp.Manifest.ProviderType] = dp
	}
}

func (r *Reloader) Start(ctx context.Context, watcher *Watcher) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-watcher.ReloadTrigger():
				r.ReloadNow(ctx)
			}
		}
	}()
}

type ReloadResult struct {
	Added   []string `json:"added"`
	Changed []string `json:"changed"`
	Removed []string `json:"removed"`
	Errors  []string `json:"errors"`
}

func (r *Reloader) ReloadNow(ctx context.Context) *ReloadResult {
	if !r.mu.TryLock() {
		return &ReloadResult{Errors: []string{"reload already in progress"}}
	}
	defer r.mu.Unlock()
	return r.reloadLocked(ctx)
}

func (r *Reloader) ReloadWait(ctx context.Context) *ReloadResult {
	backoff := 50 * time.Millisecond
	deadline := time.Now().Add(10 * time.Second)
	for {
		if r.mu.TryLock() {
			res := r.reloadLocked(ctx)
			r.mu.Unlock()
			return res
		}
		if !time.Now().Before(deadline) {
			break
		}
		wait := backoff
		if d := time.Until(deadline); d < wait {
			wait = d
		}
		select {
		case <-ctx.Done():
			return &ReloadResult{Errors: []string{"reload wait cancelled: " + ctx.Err().Error()}}
		case <-time.After(wait):
		}
		backoff *= 2
		if backoff > time.Second {
			backoff = time.Second
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.reloadLocked(ctx)
}

func (r *Reloader) reloadLocked(ctx context.Context) *ReloadResult {
	result := &ReloadResult{}

	discovered, failures := plugin.Discover(ctx, r.cfg)
	for _, f := range failures {
		errMsg := fmt.Sprintf("discovery: %s: %v", f.Dir, f.Err)
		r.logger.Warn("plugin reload discovery skipped", slog.String("dir", f.Dir), slog.Any("error", f.Err))
		result.Errors = append(result.Errors, errMsg)
	}

	next := make(map[string]*plugin.DiscoveredPlugin, len(discovered))
	for _, dp := range discovered {
		next[dp.Manifest.ProviderType] = dp
	}

	for pt := range r.state {
		if _, ok := next[pt]; !ok {
			r.unregisterOne(pt)
			result.Removed = append(result.Removed, pt)
		}
	}

	for pt, dp := range next {
		if _, ok := r.state[pt]; ok {
			r.unregisterOne(pt)
			if err := registerOne(ctx, r.manager, dp, r.catalog, r.logger); err != nil {
				errMsg := fmt.Sprintf("update %q: %v", pt, err)
				r.logger.Error("plugin reload update failed", slog.String("provider", pt), slog.Any("error", err))
				result.Errors = append(result.Errors, errMsg)
				continue
			}
			result.Changed = append(result.Changed, pt)
		} else {
			if err := registerOne(ctx, r.manager, dp, r.catalog, r.logger); err != nil {
				errMsg := fmt.Sprintf("add %q: %v", pt, err)
				r.logger.Error("plugin reload add failed", slog.String("provider", pt), slog.Any("error", err))
				result.Errors = append(result.Errors, errMsg)
				continue
			}
			result.Added = append(result.Added, pt)
		}
	}

	r.state = next

	r.logger.Info("plugin reload complete",
		slog.Int("added", len(result.Added)),
		slog.Int("changed", len(result.Changed)),
		slog.Int("removed", len(result.Removed)),
		slog.Int("errors", len(result.Errors)))
	return result
}

func (r *Reloader) unregisterOne(providerType string) {
	unregisterEnvelope(providerType)
	deployers.Registries.Unregister(domain.DeploymentProviderType(providerType))
	r.catalog.Remove(providerType)
	r.logger.Debug("plugin unregistered", slog.String("provider", providerType))
}

func unregisterEnvelope(providerType string) {
	if env, ok := providerschema.Envelopes.GetEnvelope(providerType); ok && env != nil {
		providerschema.Envelopes.UnregisterEnvelope(providerType)
	}
}
