package pluginhost

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	defaultDebounce     = 500 * time.Millisecond
	defaultPollInterval = 30 * time.Second
)

type Watcher struct {
	dir          string
	debounce     time.Duration
	pollInterval time.Duration
	reloadCh     chan struct{}
	logger       *slog.Logger
}

func NewWatcher(dir string, logger *slog.Logger) *Watcher {
	if logger == nil {
		logger = slog.Default()
	}
	return &Watcher{
		dir:          dir,
		debounce:     defaultDebounce,
		pollInterval: defaultPollInterval,
		reloadCh:     make(chan struct{}, 1),
		logger:       logger.With(slog.String("component", "pluginwatcher")),
	}
}

func (w *Watcher) ReloadTrigger() <-chan struct{} {
	return w.reloadCh
}

func (w *Watcher) Start(ctx context.Context) {
	go w.pollLoop(ctx)

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		w.logger.Warn("fsnotify init failed, using polling-only mode", slog.Any("error", err))
		return
	}
	if err := fsw.Add(w.dir); err != nil {
		w.logger.Warn("fsnotify cannot watch plugin dir, using polling-only mode", slog.String("dir", w.dir), slog.Any("error", err))
		fsw.Close()
		return
	}
	go w.fsnotifyLoop(ctx, fsw)
}

func (w *Watcher) Stop() {
}

func (w *Watcher) trigger() {
	select {
	case w.reloadCh <- struct{}{}:
	default:
	}
}

func (w *Watcher) fsnotifyLoop(ctx context.Context, fsw *fsnotify.Watcher) {
	defer fsw.Close()

	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}
	timerActive := false

	for {
		select {
		case <-ctx.Done():
			if timerActive {
				timer.Stop()
			}
			return
		case event, ok := <-fsw.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
				if !timerActive {
					timer.Reset(w.debounce)
					timerActive = true
				}
			}
		case err, ok := <-fsw.Errors:
			if !ok {
				return
			}
			w.logger.Warn("fsnotify error", slog.Any("error", err))
		case <-timer.C:
			timerActive = false
			w.trigger()
		}
	}
}

func (w *Watcher) pollLoop(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.trigger()
		}
	}
}

func (w *Watcher) Fingerprint() map[string][2]int64 {
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return nil
	}
	fp := make(map[string][2]int64, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fp[entry.Name()] = [2]int64{info.ModTime().Unix(), info.Size()}
	}
	return fp
}
