package plugin

import (
	"errors"
	"testing"
)

func TestTypedErrors_Distinguishable(t *testing.T) {
	notFound := &ErrPluginNotFound{ProviderType: "x"}
	incompat := &ErrPluginIncompatible{ProviderType: "x", Have: 1, Want: 2}
	crashed := &ErrPluginCrashed{ProviderType: "x", StderrTail: "boom", Inner: errors.New("conn closed")}

	if !IsPluginNotFound(notFound) || IsPluginNotFound(incompat) {
		t.Fatal("IsPluginNotFound misclassified")
	}
	if !IsPluginIncompatible(incompat) || IsPluginIncompatible(crashed) {
		t.Fatal("IsPluginIncompatible misclassified")
	}
	if !IsPluginCrashed(crashed) || IsPluginCrashed(notFound) {
		t.Fatal("IsPluginCrashed misclassified")
	}

	var target *ErrPluginCrashed
	if !errors.As(crashed, &target) {
		t.Fatal("errors.As failed for ErrPluginCrashed")
	}
	if target.StderrTail != "boom" {
		t.Fatalf("stderr tail not preserved: %q", target.StderrTail)
	}
}

func TestErrPluginIncompatible_WithReason(t *testing.T) {
	e := &ErrPluginIncompatible{ProviderType: "x", Reason: "core above max"}
	if e.Error() == "" {
		t.Fatal("empty message")
	}
	e2 := &ErrPluginIncompatible{ProviderType: "x", Have: 1, Want: 2}
	if e2.Error() == "" {
		t.Fatal("empty message without reason")
	}
}
