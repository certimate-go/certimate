package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type DiscoveredPlugin struct {
	Manifest   *Manifest
	Dir        string
	BinaryPath string
}

type DiscoveryFailure struct {
	ProviderType string
	Dir          string
	Err          error
}

func (f DiscoveryFailure) Error() string {
	return fmt.Sprintf("plugin discovery %s: %v", f.Dir, f.Err)
}

func (f DiscoveryFailure) Unwrap() error {
	return f.Err
}

func Discover(ctx context.Context, cfg PluginConfig) ([]*DiscoveredPlugin, []DiscoveryFailure) {
	cfg.defaults()

	entries, err := os.ReadDir(cfg.PluginDir)
	if err != nil {
		return nil, []DiscoveryFailure{{Dir: cfg.PluginDir, Err: fmt.Errorf("read plugin dir: %w", err)}}
	}

	var found []*DiscoveredPlugin
	var failures []DiscoveryFailure
	for _, entry := range entries {
		if ctx.Err() != nil {
			break
		}
		if !entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		dir := filepath.Join(cfg.PluginDir, entry.Name())
		dp, ferr := discoverOne(dir, cfg.CoreVersion)
		if ferr != nil {
			failures = append(failures, *ferr)
			continue
		}
		found = append(found, dp)
	}
	return found, failures
}

func discoverOne(dir, coreVersion string) (*DiscoveredPlugin, *DiscoveryFailure) {
	manifest, _, err := loadManifest(dir)
	if err != nil {
		return nil, &DiscoveryFailure{Dir: dir, Err: err}
	}

	if err := versionGate(manifest, coreVersion); err != nil {
		return nil, &DiscoveryFailure{ProviderType: manifest.ProviderType, Dir: dir, Err: err}
	}

	if manifest.OS != "" && manifest.Arch != "" {
		if manifest.OS != runtime.GOOS || manifest.Arch != runtime.GOARCH {
			return nil, &DiscoveryFailure{
				ProviderType: manifest.ProviderType,
				Dir:          dir,
				Err: &ErrPluginIncompatible{
					ProviderType: manifest.ProviderType,
					Reason:       fmt.Sprintf("os/arch mismatch: manifest %s/%s, host %s/%s", manifest.OS, manifest.Arch, runtime.GOOS, runtime.GOARCH),
				},
			}
		}
	}

	binaryPath, err := resolveBinary(dir, manifest.Binary)
	if err != nil {
		return nil, &DiscoveryFailure{ProviderType: manifest.ProviderType, Dir: dir, Err: err}
	}

	if err := permissionGate(dir, binaryPath); err != nil {
		return nil, &DiscoveryFailure{ProviderType: manifest.ProviderType, Dir: dir, Err: err}
	}

	return &DiscoveredPlugin{Manifest: manifest, Dir: dir, BinaryPath: binaryPath}, nil
}

func versionGate(m *Manifest, coreVersion string) error {
	if m.ProtocolVersion != ProtocolVersion {
		return &ErrPluginIncompatible{
			ProviderType: m.ProviderType,
			Have:         m.ProtocolVersion,
			Want:         ProtocolVersion,
		}
	}
	if coreVersion == "" {
		return nil
	}
	if m.MinCoreVersion != "" && semverLess(coreVersion, m.MinCoreVersion) {
		return &ErrPluginIncompatible{ProviderType: m.ProviderType, Reason: fmt.Sprintf("core %s below min %s", coreVersion, m.MinCoreVersion)}
	}
	if m.MaxCoreVersion != "" && semverLess(m.MaxCoreVersion, coreVersion) {
		return &ErrPluginIncompatible{ProviderType: m.ProviderType, Reason: fmt.Sprintf("core %s above max %s", coreVersion, m.MaxCoreVersion)}
	}
	return nil
}

func resolveBinary(dir, name string) (string, error) {
	candidate := name
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(dir, name)
	}
	abs, err := filepath.Abs(candidate)
	if err != nil {
		return "", fmt.Errorf("plugin: resolve binary path: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("plugin: binary not found %q: %w", abs, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("plugin: binary path %q is a directory", abs)
	}
	return abs, nil
}

func permissionGate(dir, binaryPath string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	for _, path := range []string{dir, binaryPath} {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("plugin: stat %q: %w", path, err)
		}
		mode := info.Mode().Perm()
		if mode&(writeGroup|writeOther) != 0 {
			return &ErrPluginConfig{
				ProviderType: filepath.Base(dir),
				Inner:        fmt.Errorf("refusing group/world-writable path %q (mode %o)", path, mode),
			}
		}
	}
	return nil
}

const (
	writeGroup = 0o020
	writeOther = 0o002
)
