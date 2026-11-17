package pluginhost

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/repository"
	"github.com/certimate-go/certimate/pkg/plugin"
	xhttp "github.com/certimate-go/certimate/pkg/utils/http"
)

const (
	maxDownloadBytes     int64 = 512 * 1024 * 1024
	maxUncompressedBytes int64 = 1024 * 1024 * 1024
	maxZipEntries              = 64

	defaultDownloadStallTimeout = 60 * time.Second
)

var (
	ErrMissingProviderType = errors.New("market: providerType is required")
	ErrOwnershipConflict   = domain.NewError(409, "market: provider belongs to a different installation source")
	ErrInstallInProgress   = domain.NewError(409, "market: an install job is already running for this source")
	ErrAlreadyInstalled    = domain.NewError(409, "market: plugin is already installed")
	ErrUpdateRolledBack    = errors.New("market: update failed; previous version restored")
	ErrUpdateCorrupted     = errors.New("market: update failed and rollback also failed; plugin left in an inconsistent state")
)

type MarketEntry struct {
	*plugin.MarketManifest
	Status string `json:"status"`
}

type MarketService struct {
	*marketCatalog
	sourcesMu   sync.Mutex
	sourceStore marketSourceStore
	sources     map[string]*marketCatalog
	leases      map[string]int
	jobs        *jobStore
	ops         *pluginOps
}

type marketCatalog struct {
	source         plugin.MarketSource
	retired        bool
	details        map[string]*plugin.MarketManifest
	indexURL       string
	downloadMirror string
	pluginDir      string
	cache          []*plugin.MarketManifest
	cachedAt       time.Time
	store          marketCache
	loaded         bool
	syncInfo       repository.PluginMarketSync
	refreshing     chan struct{}
	retryAt        time.Time
	refreshErr     error
	cacheTTL       time.Duration
	mu             sync.RWMutex
	httpClient     *http.Client
	downloadClient *http.Client
	stallTimeout   time.Duration
	logger         *slog.Logger
}

type MarketConfig struct {
	Cache             marketCache
	Sources           marketSourceStore
	CatalogTransport  http.RoundTripper
	DownloadTransport http.RoundTripper
	IndexURL          string
	DownloadMirror    string
	PluginDir         string
	CacheTTL          time.Duration
	StallTimeout      time.Duration
	Logger            *slog.Logger
}

func NewMarketService(cfg MarketConfig) *MarketService {
	official := newMarketCatalog(cfg, plugin.OfficialMarketSource())
	return &MarketService{
		marketCatalog: official,
		sourceStore:   cfg.Sources,
		leases:        make(map[string]int),
		jobs:          newJobStore(),
		ops:           newPluginOps(),
	}
}

func newMarketCatalog(cfg MarketConfig, source plugin.MarketSource) *marketCatalog {
	if cfg.CacheTTL <= 0 {
		cfg.CacheTTL = 5 * time.Minute
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.IndexURL == "" {
		cfg.IndexURL = source.IndexURL()
	}
	if cfg.StallTimeout <= 0 {
		cfg.StallTimeout = defaultDownloadStallTimeout
	}
	dlTransport := xhttp.NewDefaultTransport()
	dlTransport.ResponseHeaderTimeout = 30 * time.Second
	var downloadTransport http.RoundTripper = dlTransport
	if cfg.DownloadTransport != nil {
		downloadTransport = cfg.DownloadTransport
	}
	return &marketCatalog{
		source:         source,
		details:        make(map[string]*plugin.MarketManifest),
		store:          cfg.Cache,
		indexURL:       cfg.IndexURL,
		downloadMirror: cfg.DownloadMirror,
		pluginDir:      cfg.PluginDir,
		cacheTTL:       cfg.CacheTTL,
		httpClient:     &http.Client{Transport: cfg.CatalogTransport, Timeout: 30 * time.Second, CheckRedirect: catalogRedirect},
		downloadClient: &http.Client{Transport: downloadTransport, CheckRedirect: releaseRedirect(cfg.DownloadMirror)},
		stallTimeout:   cfg.StallTimeout,
		logger:         cfg.Logger.With(slog.String("component", "market")),
	}
}

func (s *MarketService) ListMarket(ctx context.Context) ([]MarketEntry, error) {
	manifests, err := s.manifests(ctx)
	if err != nil {
		return nil, err
	}
	entries := make([]MarketEntry, len(manifests))
	for i, mm := range manifests {
		entries[i] = MarketEntry{MarketManifest: mm, Status: s.computeStatus(mm)}
	}
	return entries, nil
}

func (s *marketCatalog) computeStatus(mm *plugin.MarketManifest) string {
	key := plugin.AssetKey(runtime.GOOS, runtime.GOARCH)
	if mm.Detail == nil || !slices.Contains(mm.Detail.Platforms, key) {
		return "unsupported_platform"
	}

	pluginPath := filepath.Join(s.pluginDir, mm.ProviderType)
	info, err := os.Stat(pluginPath)
	if err != nil || !info.IsDir() {
		return "not_installed"
	}

	meta, err := plugin.ReadMarketMeta(s.pluginDir, mm.ProviderType)
	if err != nil || meta.Origin() != s.source {
		return "installed_manual"
	}

	if plugin.CompareVersions(meta.InstalledVersion, mm.Version) < 0 {
		return "update_available"
	}
	return "installed"
}

func (s *marketCatalog) GetMarketManifest(ctx context.Context, providerType string) (*plugin.MarketManifest, error) {
	manifests, err := s.manifests(ctx)
	if err != nil {
		return nil, err
	}
	for _, mm := range manifests {
		if mm != nil && mm.ProviderType == providerType {
			return s.resolveMarketDetail(ctx, mm)
		}
	}
	return nil, fmt.Errorf("market: plugin %q not found in market listing", providerType)
}

func (s *MarketService) JobStatus(providerType string) (JobStatus, bool) {
	j := s.jobs.get(providerType)
	if j == nil {
		return JobStatus{}, false
	}
	return j.Status(), true
}

func (s *MarketService) Install(ctx context.Context, providerType string, sources ...string) (*InstallJob, error) {
	if err := plugin.ValidateProviderType(providerType); err != nil {
		return nil, err
	}
	id := s.source.ID
	if len(sources) > 0 && sources[0] != "" {
		id = sources[0]
	}

	if err := s.ops.claim(providerType, opInstall); err != nil {
		if status, ok := s.JobStatus(providerType); ok && !status.State.Terminal() {
			if status.Source.ID == id {
				return nil, ErrInstallInProgress
			}
			return nil, ErrOwnershipConflict
		}
		return nil, err
	}
	catalog, release, err := s.leaseSource(ctx, id)
	if err != nil {
		s.ops.release(providerType)
		return nil, err
	}
	targetDir := filepath.Join(s.pluginDir, providerType)
	if _, err := os.Lstat(targetDir); !os.IsNotExist(err) {
		if err == nil {
			meta, _ := plugin.ReadMarketMeta(s.pluginDir, providerType)
			origin := meta.Origin().ID
			if origin != id && (len(sources) > 0 || origin != "") {
				err = ErrOwnershipConflict
			} else {
				err = ErrAlreadyInstalled
			}
		}
		release()
		s.ops.release(providerType)
		return nil, err
	}

	job := newInstallJob(providerType)
	job.source = catalog.source
	s.jobs.set(job)
	go s.runInstall(catalog, providerType, job, release)
	return job, nil
}

func (s *MarketService) runInstall(catalog *marketCatalog, providerType string, job *InstallJob, release func()) {
	defer s.ops.release(providerType)
	defer release()
	ctx := context.Background()

	mm, err := catalog.GetMarketManifest(ctx, providerType)
	if err != nil {
		job.fail(err.Error())
		return
	}
	tmpDir, err := catalog.installPipeline(ctx, mm, job)
	if err != nil {
		job.fail(err.Error())
		s.logger.Warn("market: async install failed", slog.String("provider", providerType), slog.Any("error", err))
		return
	}

	defer os.RemoveAll(tmpDir)
	job.setState(JobReloading, "reloading")
	reloader := GlobalReloader()
	if reloader == nil {
		job.fail("market: reloader not initialized")
		return
	}
	result, err := reloader.activate(ctx, providerType, tmpDir, false)
	if err != nil {
		job.fail(err.Error())
		return
	}
	job.succeed(result)
}

func (s *MarketService) Delete(ctx context.Context, providerType string, expected ...string) (*ReloadResult, error) {
	if err := plugin.ValidateProviderType(providerType); err != nil {
		return nil, err
	}
	if err := s.ops.claim(providerType, opDelete); err != nil {
		return nil, err
	}
	defer s.ops.release(providerType)

	targetDir := filepath.Join(s.pluginDir, providerType)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("market: plugin %q is not installed", providerType)
	}

	meta, _ := plugin.ReadMarketMeta(s.pluginDir, providerType)
	source := meta.Origin()
	if len(expected) > 0 && expected[0] != source.ID {
		return nil, ErrOwnershipConflict
	}
	if source.ID == "" {
		return nil, fmt.Errorf("market: plugin %q is not managed by marketplace", providerType)
	}

	reloader := GlobalReloader()
	if reloader == nil {
		return nil, fmt.Errorf("market: reloader not initialized")
	}
	return reloader.remove(ctx, providerType)
}

func (s *MarketService) Update(ctx context.Context, providerType string, expected ...string) (*ReloadResult, error) {
	if err := plugin.ValidateProviderType(providerType); err != nil {
		return nil, err
	}
	if err := s.ops.claim(providerType, opUpdate); err != nil {
		return nil, err
	}
	defer s.ops.release(providerType)

	targetDir := filepath.Join(s.pluginDir, providerType)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("market: plugin %q is not installed", providerType)
	}

	meta, _ := plugin.ReadMarketMeta(s.pluginDir, providerType)
	source := meta.Origin()
	if len(expected) > 0 && expected[0] != source.ID {
		return nil, ErrOwnershipConflict
	}
	if source.ID == "" {
		return nil, ErrOwnershipConflict
	}

	catalog, release, err := s.leaseSource(ctx, source.ID)
	if err != nil {
		return nil, err
	}
	defer release()
	mm, err := catalog.GetMarketManifest(ctx, providerType)
	if err != nil {
		return nil, err
	}

	if plugin.CompareVersions(meta.InstalledVersion, mm.Version) >= 0 {
		return nil, fmt.Errorf("market: plugin %q is already up to date", providerType)
	}

	tmpDir, err := catalog.installPipeline(ctx, mm, nil)
	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(tmpDir)
	reloader := GlobalReloader()
	if reloader == nil {
		return nil, fmt.Errorf("market: reloader not initialized")
	}
	return reloader.activate(ctx, providerType, tmpDir, true)
}

func (s *marketCatalog) installPipeline(ctx context.Context, mm *plugin.MarketManifest, job *InstallJob) (resultDir string, err error) {
	providerType := mm.ProviderType
	if mm.Release == nil {
		return "", fmt.Errorf("market: plugin %q has no release", providerType)
	}
	if err := s.validateReleaseRepo(mm.Release.Repo); err != nil {
		return "", err
	}
	if err := plugin.ValidateBinaryName(mm.Binary); err != nil {
		return "", err
	}

	key := plugin.AssetKey(runtime.GOOS, runtime.GOARCH)
	assetName := mm.Release.Assets[key]
	if assetName == "" {
		return "", fmt.Errorf("market: plugin %q has no asset for %s", providerType, key)
	}
	if !strings.HasSuffix(assetName, ".zip") || strings.ContainsAny(assetName, "/\\") {
		return "", fmt.Errorf("market: plugin %q asset %q is not a zip archive", providerType, assetName)
	}

	tmpDir := filepath.Join(s.pluginDir, ".tmp-"+providerType)
	_ = os.RemoveAll(tmpDir)
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", fmt.Errorf("market: create temp dir: %w", err)
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(tmpDir)
		}
	}()

	tag := mm.Release.Tag
	if tag == "" || strings.Contains(tag, "\\") || slices.ContainsFunc(strings.Split(tag, "/"), func(part string) bool { return part == "" || part == "." || part == ".." }) {
		return "", fmt.Errorf("market: invalid release tag %q", tag)
	}
	job.setState(JobDownloading, "downloading")
	downloadURL := (&url.URL{Scheme: "https", Host: "github.com", Path: "/" + mm.Release.Repo + "/releases/download/" + tag + "/" + assetName}).String()
	if s.downloadMirror != "" {
		downloadURL = strings.TrimRight(s.downloadMirror, "/") + "/" + downloadURL
	}
	download, err := os.CreateTemp(s.pluginDir, ".dl-*.zip")
	if err != nil {
		return "", err
	}
	downloadPath := download.Name()
	defer os.Remove(downloadPath)
	if err := download.Close(); err != nil {
		return "", err
	}
	if err := s.fetchZip(ctx, downloadURL, downloadPath, func(downloaded, total int64) {
		job.setProgress(downloaded, total)
	}); err != nil {
		return "", err
	}

	job.setState(JobVerifying, "verifying")
	computed, err := sha256Path(downloadPath)
	if err != nil {
		return "", fmt.Errorf("market: compute sha256: %w", err)
	}
	expected := mm.Release.Checksums[key]
	if expected == "" {
		return "", fmt.Errorf("market: plugin %q has no published checksum for %s", providerType, key)
	}
	if expected != computed {
		return "", fmt.Errorf("market: plugin %q checksum mismatch for %s: expected %s, got %s", providerType, key, expected, computed)
	}

	job.setState(JobExtracting, "extracting")
	zipFile, err := os.Open(downloadPath)
	if err != nil {
		return "", fmt.Errorf("market: open download: %w", err)
	}
	zipStat, err := zipFile.Stat()
	if err != nil {
		zipFile.Close()
		return "", fmt.Errorf("market: stat download: %w", err)
	}
	_, extractErr := extractZipArchive(zipFile, zipStat.Size(), tmpDir, mm.Binary)
	zipFile.Close()
	if extractErr != nil {
		return "", extractErr
	}

	bundledPath := filepath.Join(tmpDir, "manifest.json")
	bundled, err := os.ReadFile(bundledPath)
	if err != nil {
		return "", fmt.Errorf("market: read bundled manifest: %w", err)
	}
	bm, err := plugin.ParseManifest(bundled)
	if err != nil {
		return "", fmt.Errorf("market: parse bundled manifest: %w", err)
	}
	if bm.ProviderType != providerType {
		return "", fmt.Errorf("market: bundled manifest provider_type %q does not match %q", bm.ProviderType, providerType)
	}
	if bm.ProtocolVersion != mm.ProtocolVersion {
		return "", fmt.Errorf("market: bundled manifest protocol_version %d does not match index %d", bm.ProtocolVersion, mm.ProtocolVersion)
	}
	if bm.Binary != mm.Binary {
		return "", fmt.Errorf("market: bundled manifest binary %q does not match index %q", bm.Binary, mm.Binary)
	}

	if err := bm.Validate(); err != nil {
		return "", err
	}
	if bm.Version != mm.Version {
		return "", fmt.Errorf("market: bundled version %q does not match index %q", bm.Version, mm.Version)
	}
	if mm.AccessProviderType != "" && bm.AccessProviderType != mm.AccessProviderType {
		return "", fmt.Errorf("market: bundled access identity does not match index")
	}
	localManifest := *bm
	localManifest.OS = runtime.GOOS
	localManifest.Arch = runtime.GOARCH
	localManifest.SHA256 = computed
	manifestData, err := json.MarshalIndent(&localManifest, "", "  ")
	if err != nil {
		return "", fmt.Errorf("market: marshal manifest: %w", err)
	}
	if err := os.WriteFile(bundledPath, manifestData, 0o644); err != nil {
		return "", fmt.Errorf("market: write manifest: %w", err)
	}

	meta := plugin.NewMarketMeta(s.source.ID, mm.Version, mm.Version, mm.Release.Tag)
	meta.Market = &s.source
	if err := plugin.WriteMarketMeta(tmpDir, ".", meta); err != nil {
		return "", err
	}
	return tmpDir, nil
}

const maxDownloadRetries = 3

var (
	errDownloadNotRetryable = errors.New("market: download failed (non-retryable)")
	errDownloadStalled      = errors.New("market: download stalled (no progress within timeout)")
)

func (s *marketCatalog) fetchZip(ctx context.Context, url, destPath string, onProgress func(downloaded, total int64)) error {
	var lastErr error
	for attempt := 0; attempt <= maxDownloadRetries; attempt++ {
		err := s.downloadAttempt(ctx, url, destPath, onProgress)
		if err == nil {
			return nil
		}
		lastErr = err
		if !isRetryableDownloadError(err) || attempt == maxDownloadRetries {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoffFor(attempt)):
		}
	}
	return lastErr
}

func (s *marketCatalog) downloadAttempt(ctx context.Context, url, destPath string, onProgress func(downloaded, total int64)) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("market: open download file: %w", err)
	}
	defer f.Close()

	offset, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("market: seek download file: %w", err)
	}

	if offset > maxDownloadBytes {
		return fmt.Errorf("market: partial download exceeds size limit: %w", errDownloadNotRetryable)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("market: create download request: %w", err)
	}
	if offset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
	}

	resp, err := s.downloadClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var total int64
	switch resp.StatusCode {
	case http.StatusOK:
		if err := f.Truncate(0); err != nil {
			return fmt.Errorf("market: reset partial download: %w", err)
		}
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("market: seek partial download: %w", err)
		}
		offset = 0
		total = resp.ContentLength
	case http.StatusPartialContent:
		total = parseFullSize(resp.Header.Get("Content-Range"))
		if total <= 0 {
			total = offset + resp.ContentLength
		}
	default:
		if resp.StatusCode < 500 {
			_ = os.Remove(destPath)
			return fmt.Errorf("market: download returned status %d: %w", resp.StatusCode, errDownloadNotRetryable)
		}
		return fmt.Errorf("market: download returned status %d", resp.StatusCode)
	}

	if total > 0 && total > maxDownloadBytes {
		return fmt.Errorf("market: download size %d exceeds max %d: %w", total, maxDownloadBytes, errDownloadNotRetryable)
	}

	var activity atomic.Int64
	activity.Store(time.Now().UnixNano())
	stalled, stopWatchdog := s.watchDownloadStall(ctx, cancel, &activity)
	defer stopWatchdog()

	pr := &progressReader{
		r:        io.LimitReader(resp.Body, maxDownloadBytes-offset+1),
		total:    total,
		base:     offset,
		on:       onProgress,
		activity: &activity,
	}
	written, err := io.Copy(f, pr)
	if err != nil {
		if stalled.Load() {
			return fmt.Errorf("market: %w", errDownloadStalled)
		}
		return err
	}

	if size := offset + written; size > maxDownloadBytes {
		return fmt.Errorf("market: download exceeds size limit: %w", errDownloadNotRetryable)
	} else if total > 0 && size != total {
		return fmt.Errorf("market: download size mismatch: %d/%d bytes", size, total)
	}
	if total > 0 {
		if onProgress != nil {
			onProgress(total, total)
		}
	}
	return nil
}

func (s *marketCatalog) watchDownloadStall(ctx context.Context, cancel context.CancelFunc, activity *atomic.Int64) (*atomic.Bool, func()) {
	stalled := &atomic.Bool{}
	done := make(chan struct{})
	tick := time.NewTicker(stallTickInterval(s.stallTimeout))
	go func() {
		defer tick.Stop()
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-tick.C:
				last := time.Unix(0, activity.Load())
				if time.Since(last) >= s.stallTimeout {
					stalled.Store(true)
					cancel()
					return
				}
			}
		}
	}()
	return stalled, func() { close(done) }
}

func stallTickInterval(stall time.Duration) time.Duration {
	t := stall / 4
	if t < 10*time.Millisecond {
		t = 10 * time.Millisecond
	}
	if t > 15*time.Second {
		t = 15 * time.Second
	}
	return t
}

func isRetryableDownloadError(err error) bool {
	if errors.Is(err, errDownloadNotRetryable) {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	return true
}

func backoffFor(attempt int) time.Duration {
	d := time.Duration(1<<attempt) * time.Second
	return min(d, 10*time.Second)
}

func parseFullSize(contentRange string) int64 {
	if contentRange == "" {
		return -1
	}
	idx := strings.LastIndex(contentRange, "/")
	if idx < 0 {
		return -1
	}
	n, err := strconv.ParseInt(strings.TrimSpace(contentRange[idx+1:]), 10, 64)
	if err != nil {
		return -1
	}
	return n
}

type progressReader struct {
	r        io.Reader
	total    int64
	base     int64
	read     int64
	last     int64
	on       func(downloaded, total int64)
	activity *atomic.Int64
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	if n > 0 {
		pr.read += int64(n)
		if pr.activity != nil {
			pr.activity.Store(time.Now().UnixNano())
		}
		abs := pr.base + pr.read
		if pr.on != nil && (abs-pr.last >= 64*1024 || err != nil) {
			pr.last = abs
			pr.on(abs, pr.total)
		}
	}
	return n, err
}

func sha256Path(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func extractZipArchive(r io.ReaderAt, size int64, dest, binaryName string) (string, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return "", fmt.Errorf("market: open zip: %w", err)
	}
	if len(zr.File) > maxZipEntries {
		return "", fmt.Errorf("market: zip has too many entries (%d)", len(zr.File))
	}

	destClean := filepath.Clean(dest)
	sep := string(os.PathSeparator)
	var (
		total   int64
		binPath string
	)
	for _, f := range zr.File {
		name := filepath.Clean(f.Name)
		if zipEntryEscapes(f.Name, dest, sep) {
			return "", fmt.Errorf("market: zip entry %q escapes destination", f.Name)
		}

		mode := f.Mode()
		if mode&(os.ModeSymlink|os.ModeDevice|os.ModeNamedPipe|os.ModeSocket) != 0 {
			return "", fmt.Errorf("market: zip entry %q has unsafe type", f.Name)
		}

		target := filepath.Join(dest, name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, mode.Perm()); err != nil {
				return "", fmt.Errorf("market: mkdir %q: %w", name, err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return "", fmt.Errorf("market: mkdir for %q: %w", name, err)
		}

		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode.Perm()|0o600)
		if err != nil {
			return "", fmt.Errorf("market: create %q: %w", name, err)
		}
		rc, err := f.Open()
		if err != nil {
			out.Close()
			return "", fmt.Errorf("market: open zip entry %q: %w", name, err)
		}
		n, copyErr := io.Copy(out, io.LimitReader(rc, maxUncompressedBytes+1))
		rc.Close()
		out.Close()
		if copyErr != nil {
			return "", fmt.Errorf("market: extract %q: %w", name, copyErr)
		}
		total += n
		if total > maxUncompressedBytes {
			return "", fmt.Errorf("market: zip exceeds max uncompressed size %d", maxUncompressedBytes)
		}

		if name == binaryName {
			if err := os.Chmod(target, 0o755); err != nil {
				return "", fmt.Errorf("market: chmod binary: %w", err)
			}
			binPath = target
		}
	}

	if binPath == "" {
		return "", fmt.Errorf("market: zip missing binary entry %q", binaryName)
	}
	_ = destClean
	return binPath, nil
}

func zipEntryEscapes(name, dest, sep string) bool {
	if filepath.IsAbs(name) {
		return true
	}
	cleaned := filepath.Clean(name)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+sep) {
		return true
	}
	target := filepath.Clean(filepath.Join(dest, cleaned))
	d := filepath.Clean(dest)
	if target != d && !strings.HasPrefix(target, d+sep) {
		return true
	}
	return false
}

func SweepOrphans(pluginDir string, logger *slog.Logger) {
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		switch {
		case !e.IsDir() && strings.HasPrefix(name, ".dl-") && strings.HasSuffix(name, ".zip"):
			if err := os.Remove(filepath.Join(pluginDir, name)); err != nil {
				logger.Warn("market: failed to remove orphan download", slog.String("file", name), slog.Any("error", err))
			}
		case strings.HasPrefix(name, ".tmp-"):
			if err := os.RemoveAll(filepath.Join(pluginDir, name)); err != nil {
				logger.Warn("market: failed to remove orphan temp dir", slog.String("dir", name), slog.Any("error", err))
			}
		case strings.HasPrefix(name, ".bak-"):
			pt := strings.TrimPrefix(name, ".bak-")
			target := filepath.Join(pluginDir, pt)
			if _, err := os.Stat(target); err == nil {
				_ = os.RemoveAll(filepath.Join(pluginDir, name))
			} else {
				if err := os.Rename(filepath.Join(pluginDir, name), target); err != nil {
					logger.Warn("market: failed to restore plugin from backup", slog.String("provider", pt), slog.Any("error", err))
				} else {
					logger.Info("market: restored plugin from backup after interrupted update", slog.String("provider", pt))
				}
			}
		}
	}
}
