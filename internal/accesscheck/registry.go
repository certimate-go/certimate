package accesscheck

import (
	"fmt"
	"sync"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/pkg/core"
)

const (
	SourceCore   = "core"
	SourcePlugin = "plugin"
)

type Checker = core.AccessChecker

type Registration struct {
	Source            string
	OwnerProviderType string
	UnavailableReason dtos.AccessTestReasonCode
	Checker           Checker
}

type Registry struct {
	mu      sync.RWMutex
	entries map[string]*Registration
	known   func(providerType string) bool
}

func NewRegistry() *Registry {
	return &Registry{
		entries: make(map[string]*Registration),
		known: func(providerType string) bool {
			return domain.IsBuiltInAccessProviderType(domain.AccessProviderType(providerType))
		},
	}
}

var defaultRegistry = NewRegistry()

func DefaultRegistry() *Registry {
	return defaultRegistry
}

func (r *Registry) Register(entry *Registration) error {
	if entry == nil {
		return fmt.Errorf("accesscheck: registration is nil")
	}

	switch entry.Source {
	case SourceCore:
		if entry.OwnerProviderType != "" {
			return fmt.Errorf("accesscheck: core registration must not set an owner provider type")
		}
	case SourcePlugin:
		if entry.OwnerProviderType == "" {
			return fmt.Errorf("accesscheck: plugin registration requires an owner provider type")
		}
	default:
		return fmt.Errorf("accesscheck: unknown registration source %q", entry.Source)
	}

	if entry.Checker == nil && entry.UnavailableReason == "" {
		return fmt.Errorf("accesscheck: registration requires a checker or an unavailable reason")
	}

	providerType := ""
	if entry.Checker != nil {
		providerType = entry.Checker.ProviderType()
	} else {
		providerType = entry.OwnerProviderType
	}
	if providerType == "" {
		return fmt.Errorf("accesscheck: registration requires a provider type")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.entries[providerType]; ok && existing.Source != entry.Source {
		return fmt.Errorf("accesscheck: provider type %q is already registered by source %q", providerType, existing.Source)
	}
	r.entries[providerType] = entry
	return nil
}

func (r *Registry) Unregister(providerType string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, providerType)
}

func (r *Registry) Get(providerType string) (*Registration, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entry, ok := r.entries[providerType]
	return entry, ok
}

func (r *Registry) SetKnownProviderResolver(fn func(providerType string) bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if fn == nil {
		r.known = func(providerType string) bool {
			return domain.IsBuiltInAccessProviderType(domain.AccessProviderType(providerType))
		}
		return
	}
	r.known = fn
}

func (r *Registry) IsKnownProvider(providerType string) bool {
	if _, ok := r.Get(providerType); ok {
		return true
	}

	r.mu.RLock()
	fn := r.known
	r.mu.RUnlock()
	return fn(providerType)
}
