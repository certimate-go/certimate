package certapply

import (
	"fmt"
	"regexp"
	"strings"

	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
	"golang.org/x/net/publicsuffix"
)

var dnsValidationLabel = regexp.MustCompile(`^[_a-z0-9]([_a-z0-9-]{0,61}[_a-z0-9])?$`)

func tencentCloudDNSRecords(detail *ssl.DvAuthDetail, domain string) ([]DNSRecord, error) {
	if detail == nil {
		return nil, nil
	}
	auths := detail.DvAuths
	if len(auths) == 0 {
		auths = []*ssl.DvAuths{{DvAuthKey: detail.DvAuthKey, DvAuthValue: detail.DvAuthValue, DvAuthDomain: detail.DvAuthDomain}}
	}
	base, err := publicsuffix.EffectiveTLDPlusOne(strings.ToLower(domain))
	if err != nil {
		return nil, fmt.Errorf("cannot determine certificate DNS zone: %w", err)
	}
	records := make([]DNSRecord, 0, len(auths))
	seen := make(map[DNSRecord]bool)
	for _, auth := range auths {
		if auth == nil || auth.DvAuthKey == nil || auth.DvAuthDomain == nil || auth.DvAuthValue == nil ||
			*auth.DvAuthKey == "" || *auth.DvAuthDomain == "" || *auth.DvAuthValue == "" {
			return nil, nil
		}
		if auth.DvAuthVerifyType != nil && *auth.DvAuthVerifyType != "" && *auth.DvAuthVerifyType != "TXT" {
			return nil, fmt.Errorf("unsupported Tencent Cloud DNS record type: %q", *auth.DvAuthVerifyType)
		}
		key := strings.ToLower(strings.TrimSuffix(*auth.DvAuthKey, "."))
		zone := strings.ToLower(strings.TrimSuffix(*auth.DvAuthDomain, "."))
		name := key
		switch {
		case key == "@":
			name = zone
		case key == zone || strings.HasSuffix(key, "."+zone):
		case strings.HasPrefix(zone, key+"."):
			name = zone
		default:
			name = key + "." + zone
		}
		if len(name) > 253 || (name != base && !strings.HasSuffix(name, "."+base)) {
			return nil, fmt.Errorf("Tencent Cloud DNS validation record is outside the certificate domain zone")
		}
		for _, label := range strings.Split(name, ".") {
			if !dnsValidationLabel.MatchString(label) {
				return nil, fmt.Errorf("invalid Tencent Cloud DNS validation record name")
			}
		}
		if len(*auth.DvAuthValue) > 255 {
			return nil, fmt.Errorf("Tencent Cloud DNS validation TXT value is too long")
		}
		record := DNSRecord{Name: name, Value: *auth.DvAuthValue}
		if !seen[record] {
			records = append(records, record)
			seen[record] = true
		}
	}
	return records, nil
}
