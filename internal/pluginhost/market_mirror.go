package pluginhost

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/plugin"
)

type marketSettings interface {
	GetByName(context.Context, string) (*domain.Settings, error)
	Save(context.Context, *domain.Settings) (*domain.Settings, error)
}

type MarketMirror struct {
	Mode    string `json:"mode"`
	Preset  string `json:"preset,omitempty"`
	BaseURL string `json:"baseUrl,omitempty"`
}

type MarketMirrorPreset struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	BaseURL string `json:"baseUrl"`
}

type MarketMirrorSettings struct {
	Mirror      MarketMirror         `json:"mirror"`
	DefaultBase string               `json:"defaultBase"`
	Presets     []MarketMirrorPreset `json:"presets"`
	Error       string               `json:"error,omitempty"`
}

var (
	mirrorDNSLabel    = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
	mirrorPathPart    = regexp.MustCompile(`^[a-zA-Z0-9._~-]+$`)
	nonPublicPrefixes = []netip.Prefix{
		netip.MustParsePrefix("0.0.0.0/8"), netip.MustParsePrefix("100.64.0.0/10"), netip.MustParsePrefix("192.0.0.0/24"),
		netip.MustParsePrefix("192.0.2.0/24"), netip.MustParsePrefix("192.88.99.0/24"), netip.MustParsePrefix("198.18.0.0/15"),
		netip.MustParsePrefix("198.51.100.0/24"), netip.MustParsePrefix("203.0.113.0/24"), netip.MustParsePrefix("240.0.0.0/4"),
		netip.MustParsePrefix("2001::/23"), netip.MustParsePrefix("2001:db8::/32"), netip.MustParsePrefix("2002::/16"),
		netip.MustParsePrefix("3fff::/20"),
	}
	globalIPv6 = netip.MustParsePrefix("2000::/3")
)

func (s *MarketService) OfficialMirror(ctx context.Context) MarketMirrorSettings {
	s.sourcesMu.Lock()
	defer s.sourcesMu.Unlock()
	s.loadMirrorLocked(ctx)
	return s.mirrorSettingsLocked()
}

func (s *MarketService) SetOfficialMirror(ctx context.Context, mirror MarketMirror) (MarketMirrorSettings, error) {
	s.sourcesMu.Lock()
	defer s.sourcesMu.Unlock()
	normalized, indexURL, err := s.normalizeMirror(mirror)
	if err != nil {
		return MarketMirrorSettings{}, domain.NewError(http.StatusBadRequest, err.Error())
	}
	if s.config.Settings == nil {
		return MarketMirrorSettings{}, fmt.Errorf("market: settings storage is unavailable")
	}
	content := domain.SettingsContent{"mode": normalized.Mode}
	if normalized.Preset != "" {
		content["preset"] = normalized.Preset
	}
	if normalized.BaseURL != "" {
		content["baseUrl"] = normalized.BaseURL
	}
	if _, err := s.config.Settings.Save(ctx, &domain.Settings{Name: domain.SettingsNamePluginMarket, Content: content}); err != nil {
		return MarketMirrorSettings{}, err
	}
	s.mirror = &normalized
	s.replaceOfficialLocked(indexURL, nil)
	current := s.official
	go func() { _, _ = current.marketSnapshot(context.Background(), false, true) }()
	return s.mirrorSettingsLocked(), nil
}

func (s *MarketService) loadMirrorLocked(ctx context.Context) {
	if s.mirror != nil {
		return
	}
	mirror := MarketMirror{Mode: "default"}
	if s.config.Settings == nil {
		s.mirror = &mirror
		return
	}
	settings, err := s.config.Settings.GetByName(ctx, domain.SettingsNamePluginMarket)
	if domain.IsRecordNotFoundError(err) {
		err = nil
	}
	if err == nil && settings != nil {
		for name, value := range settings.Content {
			text, ok := value.(string)
			if !ok {
				err = fmt.Errorf("market: invalid persisted mirror field %q", name)
				break
			}
			switch name {
			case "mode":
				mirror.Mode = text
			case "preset":
				mirror.Preset = text
			case "baseUrl":
				mirror.BaseURL = text
			default:
				err = fmt.Errorf("market: unexpected persisted mirror field %q", name)
			}
		}
	}
	indexURL := ""
	if err == nil {
		mirror, indexURL, err = s.normalizeMirror(mirror)
	}
	if err != nil {
		if previous := s.official.configErr; previous == nil || previous.Error() != err.Error() {
			s.replaceOfficialLocked("", err)
		}
		return
	}
	s.mirror = &mirror
	s.replaceOfficialLocked(indexURL, nil)
}

func (s *MarketService) mirrorSettingsLocked() MarketMirrorSettings {
	mirror := MarketMirror{Mode: "default"}
	if s.mirror != nil {
		mirror = *s.mirror
	}
	settings := MarketMirrorSettings{
		Mirror: mirror, DefaultBase: plugin.OfficialMarketDefaultBase,
		Presets: append([]MarketMirrorPreset{}, s.config.MirrorPresets...),
	}
	if s.official.configErr != nil {
		settings.Error = s.official.configErr.Error()
	}
	return settings
}

func (s *MarketService) replaceOfficialLocked(indexURL string, configErr error) {
	cfg := s.config
	if indexURL != "" {
		cfg.IndexURL = indexURL
	}
	if cfg.IndexURL == "" {
		cfg.IndexURL = s.source.IndexURL()
	}
	previous := s.official
	if previous.indexURL == cfg.IndexURL && previous.configErr == nil && configErr == nil {
		return
	}
	next := newMarketCatalog(cfg, s.source)
	next.configErr = configErr
	previous.mu.Lock()
	previous.retired = true
	next.loaded, next.cache, next.syncInfo = previous.loaded, previous.cache, previous.syncInfo
	previous.mu.Unlock()
	s.official = next
	if s.sources != nil {
		s.sources[s.source.ID] = next
	}
}

func (s *MarketService) normalizeMirror(mirror MarketMirror) (MarketMirror, string, error) {
	switch {
	case mirror.Mode == "default" && mirror.Preset == "" && mirror.BaseURL == "":
		return mirror, "", nil
	case mirror.Mode == "github" && mirror.Preset == "" && mirror.BaseURL == "":
		return mirror, plugin.OfficialMarketGitHubIndexURL, nil
	case mirror.Mode == "preset" && mirror.BaseURL == "":
		for _, preset := range s.config.MirrorPresets {
			if preset.ID == mirror.Preset {
				base, err := normalizeMirrorBase(preset.BaseURL)
				return mirror, base + plugin.OfficialMarketIndexPath, err
			}
		}
	case mirror.Mode == "custom" && mirror.Preset == "":
		base, err := normalizeMirrorBase(mirror.BaseURL)
		mirror.BaseURL = base
		return mirror, base + plugin.OfficialMarketIndexPath, err
	}
	return MarketMirror{}, "", fmt.Errorf("market: expected default, GitHub origin, an available mirror preset, or a custom HTTPS base")
}

func normalizeMirrorBase(input string) (string, error) {
	raw := strings.TrimSpace(input)
	base, err := url.Parse(raw)
	if err != nil || len(raw) > 1024 || base.Scheme != "https" || base.Hostname() == "" || base.Opaque != "" ||
		base.User != nil || base.ForceQuery || strings.ContainsAny(raw, "#?%\\") || strings.HasSuffix(base.Host, ":") {
		return "", fmt.Errorf("market: HTTPS mirror must not contain credentials, escapes, query or fragment")
	}
	host := strings.ToLower(base.Hostname())
	if addr, err := netip.ParseAddr(host); err == nil {
		if !publicMirrorIP(addr) {
			return "", fmt.Errorf("market: mirror destination must be public")
		}
	} else {
		labels := strings.Split(host, ".")
		if len(host) > 253 || len(labels) < 2 || slices.ContainsFunc(labels, func(label string) bool {
			return !mirrorDNSLabel.MatchString(label)
		}) || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".home.arpa") {
			return "", fmt.Errorf("market: mirror must use a public hostname")
		}
	}
	if port := base.Port(); port != "" {
		number, err := strconv.ParseUint(port, 10, 16)
		if err != nil || number == 0 {
			return "", fmt.Errorf("market: invalid HTTPS mirror port")
		}
		if number != 443 {
			host = net.JoinHostPort(host, strconv.FormatUint(number, 10))
		} else if strings.Contains(host, ":") {
			host = "[" + host + "]"
		}
	} else if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	prefix := strings.TrimSuffix(base.Path, "/")
	if prefix != "" {
		for segment := range strings.SplitSeq(strings.TrimPrefix(prefix, "/"), "/") {
			if !mirrorPathPart.MatchString(segment) || segment == "." || segment == ".." {
				return "", fmt.Errorf("market: invalid HTTPS mirror path prefix")
			}
		}
	}
	base.Host, base.Path = host, prefix
	return base.String(), nil
}

func publicMirrorIP(addr netip.Addr) bool {
	addr = addr.Unmap()
	return addr.IsGlobalUnicast() && !addr.IsPrivate() && (!addr.Is6() || globalIPv6.Contains(addr)) &&
		!slices.ContainsFunc(nonPublicPrefixes, func(prefix netip.Prefix) bool { return prefix.Contains(addr) })
}

type (
	lookupMirrorIP func(context.Context, string, string) ([]netip.Addr, error)
	dialMirrorIP   func(context.Context, string, string) (net.Conn, error)
)

func dialPublicMirror(ctx context.Context, network, address string, lookup lookupMirrorIP, dial dialMirrorIP) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	var addresses []netip.Addr
	if addr, err := netip.ParseAddr(host); err == nil {
		addresses = []netip.Addr{addr}
	} else {
		addresses, err = lookup(ctx, "ip", host)
		if err != nil {
			return nil, err
		}
	}
	if len(addresses) == 0 || slices.ContainsFunc(addresses, func(addr netip.Addr) bool { return !publicMirrorIP(addr) }) {
		return nil, fmt.Errorf("market: mirror DNS must resolve only to public IP addresses")
	}
	var dialErr error
	for _, addr := range addresses {
		connection, err := dial(ctx, network, net.JoinHostPort(addr.Unmap().String(), port))
		if err == nil {
			return connection, nil
		}
		dialErr = err
	}
	return nil, dialErr
}

func newOfficialMirrorTransport() *http.Transport {
	dialer := &net.Dialer{Timeout: 15 * time.Second, KeepAlive: 30 * time.Second}
	return &http.Transport{
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
			defer cancel()
			return dialPublicMirror(ctx, network, address, net.DefaultResolver.LookupNetIP, dialer.DialContext)
		},
		ForceAttemptHTTP2: true, MaxIdleConns: 32, IdleConnTimeout: 90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second, ResponseHeaderTimeout: 30 * time.Second,
	}
}
