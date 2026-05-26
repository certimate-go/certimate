package matrix

import "testing"

func TestConfigFromMap_requiresHomeserver(t *testing.T) {
	_, err := ConfigFromMap(map[string]any{})
	if err == nil {
		t.Fatal("expected error for empty config")
	}
}

func TestStableDeviceID_deterministic(t *testing.T) {
	cfg := &NotifierConfig{
		HomeserverUrl: "https://matrix.example.org",
		UserId:        "@bot:matrix.example.org",
	}
	a := stableDeviceID(cfg)
	b := stableDeviceID(cfg)
	if a != b || a == "" {
		t.Fatalf("expected stable non-empty device id, got %q and %q", a, b)
	}
}
