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
		domain.AccessProviderTypeFTP,
		domain.AccessProviderTypeS3,
		domain.AccessProviderTypeWebhook,
		domain.AccessProviderTypeCPanel,
		domain.AccessProviderTypeArvanCloud,
		domain.AccessProviderTypeBunny,
		domain.AccessProviderTypeDeSEC,
		domain.AccessProviderTypeDigitalOcean,
		domain.AccessProviderTypeDynu,
		domain.AccessProviderTypeDynv6,
		domain.AccessProviderTypeGandinet,
		domain.AccessProviderTypeGcore,
		domain.AccessProviderTypeHetzner,
		domain.AccessProviderTypeHostingde,
		domain.AccessProviderTypeHostinger,
		domain.AccessProviderTypeInfomaniak,
		domain.AccessProviderTypeLinode,
		domain.AccessProviderTypeNameSilo,
		domain.AccessProviderTypeNetlify,
		domain.AccessProviderTypeNS1,
		domain.AccessProviderTypeRainYun,
		domain.AccessProviderTypeVultr,
		domain.AccessProviderTypeAliyun,
		domain.AccessProviderTypeTencentCloud,
		domain.AccessProviderTypeHuaweiCloud,
		domain.AccessProviderTypeAWS,
		domain.AccessProviderTypeAzure,
		domain.AccessProviderTypeGoogleCloud,
		domain.AccessProviderTypeVolcEngine,
		domain.AccessProviderTypeBytePlus,
		domain.AccessProviderTypeBaiduCloud,
		domain.AccessProviderType35cn,
		domain.AccessProviderType51DNScom,
		domain.AccessProviderTypeBeget,
		domain.AccessProviderTypeClouDNS,
		domain.AccessProviderTypeConstellix,
		domain.AccessProviderTypeDNSLA,
		domain.AccessProviderTypeDNSMadeEasy,
		domain.AccessProviderTypeDynadot,
		domain.AccessProviderTypeGname,
		domain.AccessProviderTypeGoDaddy,
		domain.AccessProviderTypeIONOS,
		domain.AccessProviderTypeLuaDNS,
		domain.AccessProviderTypeNamecheap,
		domain.AccessProviderTypeNameDotCom,
		domain.AccessProviderTypeNetcup,
		domain.AccessProviderTypePorkbun,
		domain.AccessProviderTypeRegru,
		domain.AccessProviderTypeRuCenter,
		domain.AccessProviderTypeSimplyCom,
		domain.AccessProviderTypeSpaceship,
		domain.AccessProviderTypeTodayNIC,
		domain.AccessProviderTypeVercel,
		domain.AccessProviderTypeWestcn,
		domain.AccessProviderTypeXinnet,
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
	domain.AccessProviderTypeDuckDNS:            "API only mutates DNS records; no read-only credential probe",
	domain.AccessProviderTypeDNSExit:            "API only mutates DNS or DDNS state; no read-only credential probe",
	domain.AccessProviderTypeBookMyName:         "DynDNS API only adds or removes records; no read-only credential probe",
}

var temporaryAccessCheckExemptions = map[domain.AccessProviderType]string{}

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
