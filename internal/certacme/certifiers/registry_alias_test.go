package certifiers

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core"
)

func TestOwnedProviderAliasesFollowRegistration(t *testing.T) {
	registry := newRegistry[string]()
	factory := func(_ *ProviderFactoryOptions) (core.ACMEChallenger, error) { return nil, nil }
	registry.MustRegister("builtin-dns", factory)
	registry.MustRegisterAlias("builtin", "builtin-dns")
	const provider, owner = "plugin-dns", "test-owner"
	assertAlias := func(alias, want string) {
		t.Helper()
		canonical, _, err := registry.Lookup(alias)
		if want == "" {
			if err == nil {
				t.Fatalf("alias %q still resolves to %q", alias, canonical)
			}
			return
		}
		if err != nil || canonical != want {
			t.Fatalf("alias %q resolves to %q (%v), want %q", alias, canonical, err, want)
		}
	}
	if err := registry.SetOwned(provider, owner, factory, "legacy"); err != nil {
		t.Fatal(err)
	}
	if err := registry.SetOwned(provider, owner, factory, "changed", "builtin"); err == nil {
		t.Fatal("plugin replaced a built-in alias")
	}
	if err := registry.Register("legacy", factory); err == nil {
		t.Fatal("another provider took an active alias")
	}
	assertAlias("legacy", provider)
	assertAlias("changed", "")
	assertAlias("builtin", "builtin-dns")
	if err := registry.SetOwned(provider, owner, factory, "changed"); err != nil {
		t.Fatal(err)
	}
	assertAlias("legacy", "")
	assertAlias("changed", provider)
	registry.RemoveOwned(provider, owner)
	assertAlias("changed", "")
}
