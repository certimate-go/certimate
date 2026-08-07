package pluginhost

import (
	"encoding/json"
	"testing"
)

func TestMarshalJSON_EmptyMap(t *testing.T) {
	if got := marshalJSON(nil); got != "{}" {
		t.Fatalf("empty map should be {}, got %q", got)
	}
}

func TestMarshalJSON_RoundTrips(t *testing.T) {
	in := map[string]any{"url": "https://example.com", "port": float64(8443)}
	s := marshalJSON(in)
	var out map[string]any
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		t.Fatalf("round-trip unmarshal: %v", err)
	}
	if out["url"] != "https://example.com" {
		t.Fatalf("url lost: %v", out)
	}
}
