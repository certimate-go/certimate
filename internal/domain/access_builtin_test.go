package domain

import (
	"slices"
	"testing"
)

func TestBuiltInAccessProviderTypes(t *testing.T) {
	got := AllBuiltInAccessProviderTypes()

	want := make([]AccessProviderType, 0, len(builtinAccessProviderTypes))
	for providerType := range builtinAccessProviderTypes {
		want = append(want, providerType)
	}
	slices.Sort(want)

	if !slices.Equal(got, want) {
		t.Fatalf("AllBuiltInAccessProviderTypes() = %v, want sorted built-in types %v", got, want)
	}
	if !IsBuiltInAccessProviderType(AccessProviderType35cn) {
		t.Fatal("35cn is not a built-in access provider")
	}
}
