package plugin

import (
	"errors"
	"fmt"
)

type ErrPluginNotFound struct {
	ProviderType string
}

func (e *ErrPluginNotFound) Error() string {
	return fmt.Sprintf("plugin: provider %q is not installed", e.ProviderType)
}

type ErrPluginIncompatible struct {
	ProviderType string
	Have         uint32
	Want         uint32
	Reason       string
}

func (e *ErrPluginIncompatible) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("plugin: provider %q incompatible: %s", e.ProviderType, e.Reason)
	}
	return fmt.Sprintf("plugin: provider %q incompatible protocol version: have %d, want %d", e.ProviderType, e.Have, e.Want)
}

type ErrPluginCrashed struct {
	ProviderType string
	StderrTail   string
	Inner        error
}

func (e *ErrPluginCrashed) Error() string {
	inner := ""
	if e.Inner != nil {
		inner = e.Inner.Error()
	}
	if e.StderrTail != "" {
		return fmt.Sprintf("plugin: provider %q crashed: %s\nstderr tail:\n%s", e.ProviderType, inner, e.StderrTail)
	}
	return fmt.Sprintf("plugin: provider %q crashed: %s", e.ProviderType, inner)
}

func (e *ErrPluginCrashed) Unwrap() error { return e.Inner }

type ErrPluginChecksum struct {
	ProviderType string
	Have         string
	Want         string
}

func (e *ErrPluginChecksum) Error() string {
	return fmt.Sprintf("plugin: provider %q checksum mismatch: have %s, want %s (advisory)", e.ProviderType, e.Have, e.Want)
}

type ErrPluginConfig struct {
	ProviderType string
	Inner        error
}

func (e *ErrPluginConfig) Error() string {
	return fmt.Sprintf("plugin: provider %q misconfigured: %v", e.ProviderType, e.Inner)
}

func (e *ErrPluginConfig) Unwrap() error { return e.Inner }

func AsPluginError(err error) error {
	if err == nil {
		return nil
	}
	var nf *ErrPluginNotFound
	var inc *ErrPluginIncompatible
	var crash *ErrPluginCrashed
	if errors.As(err, &nf) || errors.As(err, &inc) || errors.As(err, &crash) {
		return err
	}
	return err
}

func IsPluginNotFound(err error) bool {
	var e *ErrPluginNotFound
	return errors.As(err, &e)
}

func IsPluginIncompatible(err error) bool {
	var e *ErrPluginIncompatible
	return errors.As(err, &e)
}

func IsPluginCrashed(err error) bool {
	var e *ErrPluginCrashed
	return errors.As(err, &e)
}
