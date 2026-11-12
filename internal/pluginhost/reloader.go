package pluginhost

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/certimate-go/certimate/pkg/plugin"
)

type Reloader struct {
	cfg     plugin.PluginConfig
	catalog *Catalog
	manager *plugin.Manager
	mu      sync.Mutex
	logger  *slog.Logger
}

func NewReloader(cfg plugin.PluginConfig, catalog *Catalog, logger *slog.Logger) *Reloader {
	if logger == nil {
		logger = slog.Default()
	}
	return &Reloader{cfg: cfg, catalog: catalog, manager: plugin.NewManager(cfg, logger), logger: logger.With(slog.String("component", "pluginreloader"))}
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

func (r *Reloader) lock(ctx context.Context) error {
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if r.mu.TryLock() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (r *Reloader) reloadLocked(ctx context.Context) *ReloadResult {
	result := &ReloadResult{}
	discovered, failures := plugin.Discover(ctx, r.cfg)
	for _, failure := range failures {
		result.Errors = append(result.Errors, fmt.Sprintf("discovery: %s: %v", failure.Dir, failure.Err))
	}
	if ctx.Err() != nil {
		result.Errors = append(result.Errors, ctx.Err().Error())
		return result
	}
	current := r.catalog.registered()
	loaded := make(map[string]bool)
	for _, dp := range discovered {
		name := dp.Manifest.ProviderType
		fingerprint, err := pluginFingerprint(dp)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("inspect %q: %v", name, err))
			continue
		}
		old := current[name]
		if old != nil && old.fingerprint == fingerprint {
			continue
		}
		if err := registerOne(ctx, r.manager, dp, r.catalog, r.logger, fingerprint); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("load %q: %v", name, err))
			continue
		}
		loaded[name] = true
		if old == nil {
			result.Added = append(result.Added, name)
		} else {
			result.Changed = append(result.Changed, name)
		}
	}
	for name, old := range current {
		if loaded[name] {
			continue
		}
		if _, err := os.Stat(old.source.Dir); os.IsNotExist(err) {
			if old.ownsAccess {
				if reusers := r.catalog.accessReusers(old.accessType); len(reusers) > 0 {
					result.Errors = append(result.Errors, fmt.Sprintf("access owner %q is missing; access type %q still used by %v", name, old.accessType, reusers))
				}
			}
			r.catalog.unregister(name)
			result.Removed = append(result.Removed, name)
		}
	}
	sort.Strings(result.Added)
	sort.Strings(result.Changed)
	sort.Strings(result.Removed)
	return result
}

func (r *Reloader) activate(ctx context.Context, providerType, tmpDir string, update bool) (*ReloadResult, error) {
	if err := r.lock(ctx); err != nil {
		return nil, err
	}
	defer r.mu.Unlock()
	targetDir := filepath.Join(r.cfg.PluginDir, providerType)
	if existing := r.catalog.registered()[providerType]; existing != nil {
		if !update || filepath.Clean(existing.source.Dir) != filepath.Clean(targetDir) {
			return nil, ErrAlreadyInstalled
		}
	}
	dp, err := plugin.DiscoverPlugin(tmpDir, r.cfg)
	if err != nil {
		return nil, err
	}
	if dp.Manifest.ProviderType != providerType {
		return nil, fmt.Errorf("plugin identity does not match %q", providerType)
	}
	prepared, err := preparePlugin(ctx, r.manager, dp, r.logger, "")
	if err != nil {
		return nil, err
	}
	published := false
	defer func() {
		if !published {
			os.RemoveAll(prepared.runtimeDir)
		}
	}()
	backupDir := filepath.Join(r.cfg.PluginDir, ".bak-"+providerType)
	if update {
		if err := os.RemoveAll(backupDir); err != nil {
			return nil, err
		}
		if err := os.Rename(targetDir, backupDir); err != nil {
			return nil, err
		}
	} else if _, err := os.Stat(targetDir); err == nil {
		return nil, ErrAlreadyInstalled
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	rollback := func(cause error) (*ReloadResult, error) {
		if err := os.RemoveAll(targetDir); err != nil {
			return nil, fmt.Errorf("%w: %v (cause: %v)", ErrUpdateCorrupted, err, cause)
		}
		if update {
			if err := os.Rename(backupDir, targetDir); err != nil {
				return nil, fmt.Errorf("%w: %v (cause: %v)", ErrUpdateCorrupted, err, cause)
			}
			return nil, fmt.Errorf("%w: %v", ErrUpdateRolledBack, cause)
		}
		return nil, cause
	}
	if err := os.Rename(tmpDir, targetDir); err != nil {
		if update {
			if restoreErr := os.Rename(backupDir, targetDir); restoreErr != nil {
				return nil, fmt.Errorf("%w: %v", ErrUpdateCorrupted, restoreErr)
			}
		}
		return nil, err
	}
	finalDP, err := plugin.DiscoverPlugin(targetDir, r.cfg)
	if err != nil {
		return rollback(err)
	}
	prepared.source = finalDP
	prepared.fingerprint, err = pluginFingerprint(finalDP)
	if err != nil {
		return rollback(err)
	}
	ordered, resolveErrs := r.catalog.resolveAccessOwnership([]*registeredPlugin{prepared})
	if len(resolveErrs) > 0 {
		return rollback(resolveErrs[0])
	}
	if err := r.catalog.publish(ordered[0]); err != nil {
		return rollback(err)
	}
	published = true
	if update {
		if err := os.RemoveAll(backupDir); err != nil {
			r.logger.Warn("could not remove plugin backup", slog.Any("error", err))
		}
		return &ReloadResult{Changed: []string{providerType}}, nil
	}
	return &ReloadResult{Added: []string{providerType}}, nil
}

func (r *Reloader) remove(ctx context.Context, providerType string) (*ReloadResult, error) {
	if err := r.lock(ctx); err != nil {
		return nil, err
	}
	defer r.mu.Unlock()
	if err := r.catalog.refuseRemoval(providerType); err != nil {
		return nil, err
	}
	if err := os.RemoveAll(filepath.Join(r.cfg.PluginDir, providerType)); err != nil {
		return nil, err
	}
	r.catalog.unregister(providerType)
	return &ReloadResult{Removed: []string{providerType}}, nil
}

func (r *Reloader) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.catalog.Close()
}
