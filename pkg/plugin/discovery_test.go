package plugin

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func writePluginDir(t *testing.T, layout map[string]map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for subdir, files := range layout {
		dir := filepath.Join(root, subdir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		for name, content := range files {
			path := filepath.Join(dir, name)
			if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
				t.Fatal(err)
			}
		}
	}
	return root
}

func manifestJSON(overrides ...func(map[string]any)) string {
	m := map[string]any{
		"version":              "0.1.0",
		"provider_type":        "demo",
		"access_provider_type": "demo",
		"display_name_key":     "plugin.demo.name",
		"deploy_category":      "other",
		"protocol_version":     ProtocolVersion,
		"binary":               "demo-binary",
	}
	for _, o := range overrides {
		o(m)
	}
	return string(jsonMarshal(m))
}

func TestDiscovery_ParsesManifestAndResolvesAbsPath_NoExec(t *testing.T) {
	dir := writePluginDir(t, map[string]map[string]string{
		"demo": {
			"manifest.json": manifestJSON(),
			"demo-binary":   "#!/bin/sh\nexit 0\n",
		},
	})
	found, failures := Discover(testContext(t), PluginConfig{PluginDir: dir, CoreVersion: "0.1.0"})
	if len(failures) != 0 {
		t.Fatalf("unexpected failures: %+v", failures)
	}
	if len(found) != 1 || found[0].Manifest.ProviderType != "demo" {
		t.Fatalf("expected 1 demo plugin, got %+v", found)
	}
	abs := found[0].BinaryPath
	if !filepath.IsAbs(abs) {
		t.Fatalf("binary path not absolute: %q", abs)
	}
}

func TestDiscovery_VersionGate_RejectsIncompatible_NoExec(t *testing.T) {
	dir := writePluginDir(t, map[string]map[string]string{
		"demo": {
			"manifest.json": manifestJSON(func(m map[string]any) { m["protocol_version"] = 999 }),
			"demo-binary":   "",
		},
	})
	found, failures := Discover(testContext(t), PluginConfig{PluginDir: dir})
	if len(found) != 0 {
		t.Fatalf("expected no plugins discovered, got %+v", found)
	}
	if len(failures) != 1 {
		t.Fatalf("expected 1 failure, got %+v", failures)
	}
	var ie *ErrPluginIncompatible
	if !errors.As(failures[0].Err, &ie) {
		t.Fatalf("expected ErrPluginIncompatible, got %v", failures[0].Err)
	}
}

func TestDiscovery_MinMaxCoreGate(t *testing.T) {
	dir := writePluginDir(t, map[string]map[string]string{
		"demo": {
			"manifest.json": manifestJSON(func(m map[string]any) { m["min_core_version"] = "9.0.0" }),
			"demo-binary":   "",
		},
	})
	found, failures := Discover(testContext(t), PluginConfig{PluginDir: dir, CoreVersion: "0.1.0"})
	if len(found) != 0 || len(failures) != 1 {
		t.Fatalf("expected rejection for core below min, found=%+v failures=%+v", found, failures)
	}
	var ie *ErrPluginIncompatible
	if !errors.As(failures[0].Err, &ie) {
		t.Fatalf("expected ErrPluginIncompatible, got %v", failures[0].Err)
	}
}

func TestDiscovery_PermissionGate_RejectsWorldWritable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits not applicable on windows")
	}
	dir := writePluginDir(t, map[string]map[string]string{
		"demo": {
			"manifest.json": manifestJSON(),
			"demo-binary":   "x",
		},
	})
	if err := os.Chmod(filepath.Join(dir, "demo", "demo-binary"), 0o777); err != nil {
		t.Fatal(err)
	}
	found, failures := Discover(testContext(t), PluginConfig{PluginDir: dir, CoreVersion: "0.1.0"})
	if len(found) != 0 || len(failures) != 1 {
		t.Fatalf("expected rejection for world-writable binary, found=%+v failures=%+v", found, failures)
	}
}

func TestDiscovery_MissingBinary_Fails(t *testing.T) {
	dir := writePluginDir(t, map[string]map[string]string{
		"demo": {
			"manifest.json": manifestJSON(),
		},
	})
	found, failures := Discover(testContext(t), PluginConfig{PluginDir: dir, CoreVersion: "0.1.0"})
	if len(found) != 0 || len(failures) != 1 {
		t.Fatalf("expected failure for missing binary, found=%+v failures=%+v", found, failures)
	}
}

func TestDiscovery_EmptyDir_NoFailures(t *testing.T) {
	dir := t.TempDir()
	found, failures := Discover(testContext(t), PluginConfig{PluginDir: dir})
	if len(found) != 0 || len(failures) != 0 {
		t.Fatalf("expected nothing for empty dir, found=%+v failures=%+v", found, failures)
	}
}
