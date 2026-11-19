package pluginhost

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/challenge/dns01"
)

func TestMarshalJSON_EmptyMap(t *testing.T) {
	if got := marshalJSON(nil); got != "{}" {
		t.Fatalf("empty map should be {}, got %q", got)
	}
}

func TestPluginDNSPropagationTiming(t *testing.T) {
	for _, test := range []struct {
		config            string
		timeout, interval time.Duration
	}{
		{"", dns01.DefaultPropagationTimeout, dns01.DefaultPollingInterval},
		{`{"propagation_timeout_seconds":240,"propagation_polling_interval_seconds":10,"custom":{"mode":"auto"}}`, 4 * time.Minute, 10 * time.Second},
		{`{"propagation_timeout_seconds":300}`, 5 * time.Minute, dns01.DefaultPollingInterval},
	} {
		timeout, interval, err := dnsRuntimeTiming([]byte(test.config))
		if err != nil || timeout != test.timeout || interval != test.interval {
			t.Errorf("timing for %s = %s, %s (%v); want %s, %s", test.config, timeout, interval, err, test.timeout, test.interval)
		}
	}
	for _, config := range []string{`{"propagation_timeout_seconds":-1}`, `{"propagation_timeout_seconds":9223372037}`} {
		if _, _, err := dnsRuntimeTiming([]byte(config)); err == nil {
			t.Errorf("invalid timing accepted: %s", config)
		}
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
