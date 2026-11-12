package pluginhost

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/certimate-go/certimate/internal/certacme/certifiers"
	"github.com/certimate-go/certimate/internal/certmgmt/deployers"
	"github.com/certimate-go/certimate/internal/notify/notifiers"
	"github.com/certimate-go/certimate/internal/providerschema"
	"github.com/certimate-go/certimate/pkg/plugin"
)

type registeredPlugin struct {
	source        *plugin.DiscoveredPlugin
	fingerprint   string
	runtimeDir    string
	capability    string
	accessType    string
	ownsAccess    bool
	factory       any
	newDeployer   deployers.ProviderFactoryFunc
	newChallenger certifiers.ProviderFactoryFunc
	newNotifier   notifiers.ProviderFactoryFunc
	envelopes     []*providerschema.Envelope
	entry         *CatalogEntry
}

func pluginFingerprint(dp *plugin.DiscoveredPlugin) (string, error) {
	hash := sha256.New()
	addFile := func(path string) error {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("plugin file %q is not a regular file", path)
		}
		_, err = io.Copy(hash, file)
		return err
	}
	err := filepath.WalkDir(dp.Dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dp.Dir, path)
		if err != nil {
			return err
		}
		fmt.Fprintf(hash, "%s:%d:%d:%d\n", rel, info.Size(), info.ModTime().UnixNano(), info.Mode())
		return addFile(path)
	})
	if err != nil {
		return "", err
	}
	info, err := os.Stat(dp.BinaryPath)
	if err != nil {
		return "", err
	}
	fmt.Fprintf(hash, "%s:%d:%d:%d", dp.BinaryPath, info.Size(), info.ModTime().UnixNano(), info.Mode())
	rel, err := filepath.Rel(dp.Dir, dp.BinaryPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		if err := addFile(dp.BinaryPath); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func snapshotPlugin(dp *plugin.DiscoveredPlugin, parent string) (*plugin.DiscoveredPlugin, string, error) {
	prefix := "certimate-plugin-"
	if parent != "" {
		prefix = ".certimate-runtime-"
	}
	dir, err := os.MkdirTemp(parent, prefix)
	if err != nil {
		return nil, "", err
	}
	err = filepath.WalkDir(dp.Dir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(dp.Dir, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dir, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		return copyPluginFile(path, target)
	})
	if err != nil {
		os.RemoveAll(dir)
		return nil, "", err
	}
	binaryRel, err := filepath.Rel(dp.Dir, dp.BinaryPath)
	if err != nil || binaryRel == ".." || strings.HasPrefix(binaryRel, ".."+string(filepath.Separator)) {
		binaryRel = ".certimate-executable"
		if err := copyPluginFile(dp.BinaryPath, filepath.Join(dir, binaryRel)); err != nil {
			os.RemoveAll(dir)
			return nil, "", err
		}
	}
	snapshot := *dp
	snapshot.Dir = dir
	snapshot.BinaryPath = filepath.Join(dir, binaryRel)
	return &snapshot, dir, nil
}

func copyPluginFile(source, target string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("plugin resource %q is not a regular file", source)
	}
	output, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm()&0o755)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
