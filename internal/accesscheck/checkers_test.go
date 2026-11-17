package accesscheck

import (
	"testing"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
)

func TestCoreCheckersRegistered(t *testing.T) {
	providerTypes := []domain.AccessProviderType{
		domain.AccessProviderTypeSSH,
		domain.AccessProviderTypeCloudflare,
	}

	for _, providerType := range providerTypes {
		t.Run(string(providerType), func(t *testing.T) {
			entry, ok := DefaultRegistry().Get(string(providerType))
			if !ok {
				t.Fatalf("expected %q to be registered in the default registry", providerType)
			}
			if entry.Source != SourceCore {
				t.Fatalf("source = %q, want %q", entry.Source, SourceCore)
			}
			if entry.OwnerProviderType != "" {
				t.Fatalf("owner provider type = %q, want empty for a core checker", entry.OwnerProviderType)
			}
			if got := entry.Checker.ProviderType(); got != string(providerType) {
				t.Fatalf("provider type = %q, want %q", got, providerType)
			}
			scope := entry.Checker.Scope()
			if len(scope) != 2 || scope[0] != core.AccessCheckItemConnect || scope[1] != core.AccessCheckItemAuth {
				t.Fatalf("scope = %v, want [%s %s]", scope, core.AccessCheckItemConnect, core.AccessCheckItemAuth)
			}
		})
	}
}

const awaitingCheckerReason = "awaiting checker"

var permanentAccessCheckExemptions = map[domain.AccessProviderType]string{
	domain.AccessProviderTypeLocal:              "no access form / not selectable",
	domain.AccessProviderTypeLetsEncrypt:        "no access form / not selectable",
	domain.AccessProviderTypeLetsEncryptStaging: "no access form / not selectable",
	domain.AccessProviderTypeCUCCCloud:          "reserved, no config yet",
	domain.AccessProviderTypeFastly:             "reserved, no config yet",
	domain.AccessProviderTypeEmail:              "notification path uses Test notify, not connectivity API",
	domain.AccessProviderTypePowerDNS:           "DNS UI migrated to plugin",
	domain.AccessProviderTypeRFC2136:            "DNS UI migrated to plugin",
}

var temporaryAccessCheckExemptions = map[domain.AccessProviderType]string{
	domain.AccessProviderType35cn:                awaitingCheckerReason,
	domain.AccessProviderType51DNScom:            awaitingCheckerReason,
	domain.AccessProviderTypeACMECA:              awaitingCheckerReason,
	domain.AccessProviderTypeACMEDNS:             awaitingCheckerReason,
	domain.AccessProviderTypeACMEHttpReq:         awaitingCheckerReason,
	domain.AccessProviderTypeActalisSSL:          awaitingCheckerReason,
	domain.AccessProviderTypeAkamai:              awaitingCheckerReason,
	domain.AccessProviderTypeAliyun:              awaitingCheckerReason,
	domain.AccessProviderTypeArvanCloud:          awaitingCheckerReason,
	domain.AccessProviderTypeAWS:                 awaitingCheckerReason,
	domain.AccessProviderTypeAzure:               awaitingCheckerReason,
	domain.AccessProviderTypeBaiduCloud:          awaitingCheckerReason,
	domain.AccessProviderTypeBeget:               awaitingCheckerReason,
	domain.AccessProviderTypeBookMyName:          awaitingCheckerReason,
	domain.AccessProviderTypeBunny:               awaitingCheckerReason,
	domain.AccessProviderTypeBytePlus:            awaitingCheckerReason,
	domain.AccessProviderTypeClouDNS:             awaitingCheckerReason,
	domain.AccessProviderTypeCMCCCloud:           awaitingCheckerReason,
	domain.AccessProviderTypeConoHaVPS:           awaitingCheckerReason,
	domain.AccessProviderTypeConstellix:          awaitingCheckerReason,
	domain.AccessProviderTypeCPanel:              awaitingCheckerReason,
	domain.AccessProviderTypeCTCCCloud:           awaitingCheckerReason,
	domain.AccessProviderTypeDeSEC:               awaitingCheckerReason,
	domain.AccessProviderTypeDigiCert:            awaitingCheckerReason,
	domain.AccessProviderTypeDigitalOcean:        awaitingCheckerReason,
	domain.AccessProviderTypeDNSExit:             awaitingCheckerReason,
	domain.AccessProviderTypeDNSLA:               awaitingCheckerReason,
	domain.AccessProviderTypeDNSMadeEasy:         awaitingCheckerReason,
	domain.AccessProviderTypeDuckDNS:             awaitingCheckerReason,
	domain.AccessProviderTypeDynadot:             awaitingCheckerReason,
	domain.AccessProviderTypeDynu:                awaitingCheckerReason,
	domain.AccessProviderTypeDynv6:               awaitingCheckerReason,
	domain.AccessProviderTypeFTP:                 awaitingCheckerReason,
	domain.AccessProviderTypeGandinet:            awaitingCheckerReason,
	domain.AccessProviderTypeGcore:               awaitingCheckerReason,
	domain.AccessProviderTypeGlobalSignAtlas:     awaitingCheckerReason,
	domain.AccessProviderTypeGname:               awaitingCheckerReason,
	domain.AccessProviderTypeGoDaddy:             awaitingCheckerReason,
	domain.AccessProviderTypeGoogleCloud:         awaitingCheckerReason,
	domain.AccessProviderTypeGoogleTrustServices: awaitingCheckerReason,
	domain.AccessProviderTypeHetzner:             awaitingCheckerReason,
	domain.AccessProviderTypeHostingde:           awaitingCheckerReason,
	domain.AccessProviderTypeHostinger:           awaitingCheckerReason,
	domain.AccessProviderTypeHuaweiCloud:         awaitingCheckerReason,
	domain.AccessProviderTypeInfomaniak:          awaitingCheckerReason,
	domain.AccessProviderTypeIONOS:               awaitingCheckerReason,
	domain.AccessProviderTypeJDCloud:             awaitingCheckerReason,
	domain.AccessProviderTypeLinode:              awaitingCheckerReason,
	domain.AccessProviderTypeLiteSSL:             awaitingCheckerReason,
	domain.AccessProviderTypeLuaDNS:              awaitingCheckerReason,
	domain.AccessProviderTypeNamecheap:           awaitingCheckerReason,
	domain.AccessProviderTypeNameDotCom:          awaitingCheckerReason,
	domain.AccessProviderTypeNameSilo:            awaitingCheckerReason,
	domain.AccessProviderTypeNetcup:              awaitingCheckerReason,
	domain.AccessProviderTypeNetlify:             awaitingCheckerReason,
	domain.AccessProviderTypeNS1:                 awaitingCheckerReason,
	domain.AccessProviderTypeOracleCloud:         awaitingCheckerReason,
	domain.AccessProviderTypeOVHcloud:            awaitingCheckerReason,
	domain.AccessProviderTypePorkbun:             awaitingCheckerReason,
	domain.AccessProviderTypeQingCloud:           awaitingCheckerReason,
	domain.AccessProviderTypeRainYun:             awaitingCheckerReason,
	domain.AccessProviderTypeRegru:               awaitingCheckerReason,
	domain.AccessProviderTypeRuCenter:            awaitingCheckerReason,
	domain.AccessProviderTypeS3:                  awaitingCheckerReason,
	domain.AccessProviderTypeSectigo:             awaitingCheckerReason,
	domain.AccessProviderTypeSimplyCom:           awaitingCheckerReason,
	domain.AccessProviderTypeSpaceship:           awaitingCheckerReason,
	domain.AccessProviderTypeSSLCOM:              awaitingCheckerReason,
	domain.AccessProviderTypeTechnitiumDNS:       awaitingCheckerReason,
	domain.AccessProviderTypeTencentCloud:        awaitingCheckerReason,
	domain.AccessProviderTypeTodayNIC:            awaitingCheckerReason,
	domain.AccessProviderTypeUCloud:              awaitingCheckerReason,
	domain.AccessProviderTypeVercel:              awaitingCheckerReason,
	domain.AccessProviderTypeVolcEngine:          awaitingCheckerReason,
	domain.AccessProviderTypeVultr:               awaitingCheckerReason,
	domain.AccessProviderTypeWebhook:             awaitingCheckerReason,
	domain.AccessProviderTypeWestcn:              awaitingCheckerReason,
	domain.AccessProviderTypeXinnet:              awaitingCheckerReason,
	domain.AccessProviderTypeYandexCloud:         awaitingCheckerReason,
	domain.AccessProviderTypeZeroSSL:             awaitingCheckerReason,
}

func TestBuiltinAccessCheckCoverage(t *testing.T) {
	builtins := domain.AllBuiltInAccessProviderTypes()
	builtinSet := make(map[domain.AccessProviderType]struct{}, len(builtins))
	for _, providerType := range builtins {
		if _, exists := builtinSet[providerType]; exists {
			t.Errorf("duplicate built-in access provider %q", providerType)
		}
		builtinSet[providerType] = struct{}{}
	}

	for providerType, reason := range permanentAccessCheckExemptions {
		if _, ok := builtinSet[providerType]; !ok {
			t.Errorf("permanent exemption %q is not a built-in access provider", providerType)
		}
		if reason == "" {
			t.Errorf("permanent exemption %q has an empty reason", providerType)
		}
		if _, also := temporaryAccessCheckExemptions[providerType]; also {
			t.Errorf("%q is listed in both exemption maps", providerType)
		}
	}
	for providerType, reason := range temporaryAccessCheckExemptions {
		if _, ok := builtinSet[providerType]; !ok {
			t.Errorf("temporary exemption %q is not a built-in access provider", providerType)
		}
		if reason != "awaiting checker" {
			t.Errorf("temporary exemption %q reason = %q, want %q", providerType, reason, "awaiting checker")
		}
	}

	for _, providerType := range builtins {
		_, permanent := permanentAccessCheckExemptions[providerType]
		_, temporary := temporaryAccessCheckExemptions[providerType]
		hasCore, hasChecker := registeredChecker(string(providerType))
		if hasChecker && (permanent || temporary) {
			t.Errorf("stale exemption for %q: a checker is registered", providerType)
		}
		if !hasCore && !permanent && !temporary {
			t.Errorf("%q has no core checker and is not exempt", providerType)
		}
	}
}

func registeredChecker(providerType string) (hasCore bool, hasChecker bool) {
	entry, ok := DefaultRegistry().Get(providerType)
	if !ok || entry == nil || entry.Checker == nil {
		return false, false
	}
	return entry.Source == SourceCore, true
}
