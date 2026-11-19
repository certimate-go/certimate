package certifiers

import (
	"fmt"
	"maps"
	"sync"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
)

type ProviderFactoryFunc func(options *ProviderFactoryOptions) (core.ACMEChallenger, error)

type ProviderFactoryOptions struct {
	ProviderAccessConfig   map[string]any
	ProviderExtendedConfig map[string]any
	DnsPropagationTimeout  int
	DnsTTL                 int
	DnsNameservers         []string
	DisableFollowCNAME     bool
}

type Registry[T comparable] interface {
	Register(T, ProviderFactoryFunc, ...T) error
	RegisterAlias(T, T) error
	MustRegister(T, ProviderFactoryFunc)
	MustRegisterAlias(T, T)
	Get(T) (ProviderFactoryFunc, error)
	Lookup(T) (T, ProviderFactoryFunc, error)
	Unregister(T)
	SetOwned(T, string, ProviderFactoryFunc, ...T) error
	RemoveOwned(T, string)
}

type registry[T comparable] struct {
	mu        sync.RWMutex
	factories map[T]ProviderFactoryFunc
	owners    map[T]string
	aliases   map[T]T
}

func (r *registry[T]) Register(name T, factory ProviderFactoryFunc, aliases ...T) error {
	return r.SetOwned(name, "", factory, aliases...)
}

func (r *registry[T]) RegisterAlias(alias, target T) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if canonical, ok := r.aliases[target]; ok {
		target = canonical
	}
	if _, exists := r.factories[target]; !exists {
		return fmt.Errorf("provider '%v' not registered", target)
	}
	if _, exists := r.factories[alias]; exists {
		return fmt.Errorf("provider '%v' already registered", alias)
	}
	if _, exists := r.aliases[alias]; exists {
		return fmt.Errorf("provider '%v' already registered", alias)
	}
	r.aliases[alias] = target
	return nil
}

func (r *registry[T]) MustRegister(name T, factory ProviderFactoryFunc) {
	if err := r.Register(name, factory); err != nil {
		panic(err)
	}
}

func (r *registry[T]) MustRegisterAlias(alias, target T) {
	if err := r.RegisterAlias(alias, target); err != nil {
		panic(err)
	}
}

func (r *registry[T]) Get(name T) (ProviderFactoryFunc, error) {
	_, factory, err := r.Lookup(name)
	return factory, err
}

func (r *registry[T]) Lookup(name T) (T, ProviderFactoryFunc, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if canonical, ok := r.aliases[name]; ok {
		name = canonical
	}
	if factory, exists := r.factories[name]; exists {
		return name, factory, nil
	}
	return name, nil, fmt.Errorf("provider '%v' not registered", name)
}

func (r *registry[T]) Unregister(name T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.aliases, name)
	r.removeLocked(name)
}

func (r *registry[T]) SetOwned(name T, owner string, factory ProviderFactoryFunc, aliases ...T) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.factories[name]; exists && (owner == "" || r.owners[name] != owner) {
		return fmt.Errorf("provider '%v' already registered", name)
	}
	if _, exists := r.aliases[name]; exists {
		return fmt.Errorf("provider '%v' already registered as an alias", name)
	}
	for _, alias := range aliases {
		_, usedByProvider := r.factories[alias]
		canonical, usedByAlias := r.aliases[alias]
		if alias == name || usedByProvider || (usedByAlias && canonical != name) {
			return fmt.Errorf("cannot register alias '%v' for provider '%v'", alias, name)
		}
	}
	maps.DeleteFunc(r.aliases, func(_, canonical T) bool { return canonical == name })
	for _, alias := range aliases {
		r.aliases[alias] = name
	}
	r.factories[name] = factory
	r.owners[name] = owner
	return nil
}

func (r *registry[T]) RemoveOwned(name T, owner string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if owner != "" && r.owners[name] == owner {
		r.removeLocked(name)
	}
}

func (r *registry[T]) removeLocked(name T) {
	delete(r.factories, name)
	delete(r.owners, name)
	maps.DeleteFunc(r.aliases, func(_, canonical T) bool { return canonical == name })
}

func newRegistry[T comparable]() Registry[T] {
	return &registry[T]{factories: make(map[T]ProviderFactoryFunc), owners: make(map[T]string), aliases: make(map[T]T)}
}

var (
	ACMEDns01Registries  = newRegistry[domain.ACMEDns01ProviderType]()
	ACMEHttp01Registries = newRegistry[domain.ACMEHttp01ProviderType]()
)
