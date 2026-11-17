package certifiers

import (
	"fmt"
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
	Register(T, ProviderFactoryFunc) error
	RegisterAlias(T, T) error
	MustRegister(T, ProviderFactoryFunc)
	MustRegisterAlias(T, T)
	Get(T) (ProviderFactoryFunc, error)
	Unregister(T)
	SetOwned(T, string, ProviderFactoryFunc) error
	RemoveOwned(T, string)
}

type registry[T comparable] struct {
	mu        sync.RWMutex
	factories map[T]ProviderFactoryFunc
	owners    map[T]string
}

func (r *registry[T]) Register(name T, factory ProviderFactoryFunc) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("provider '%v' already registered", name)
	}

	r.factories[name] = factory
	return nil
}

func (r *registry[T]) RegisterAlias(name T, alias T) error {
	factory, err := r.Get(alias)
	if err != nil {
		return err
	}

	err = r.Register(name, factory)
	if err != nil {
		return err
	}

	return nil
}

func (r *registry[T]) MustRegister(name T, factory ProviderFactoryFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.factories[name]; exists {
		panic(fmt.Sprintf("provider '%v' already registered", name))
	}
	r.factories[name] = factory
}

func (r *registry[T]) MustRegisterAlias(name T, alias T) {
	if err := r.RegisterAlias(name, alias); err != nil {
		panic(err)
	}
}

func (r *registry[T]) Get(name T) (ProviderFactoryFunc, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if factory, exists := r.factories[name]; exists {
		return factory, nil
	}

	return nil, fmt.Errorf("provider '%v' not registered", name)
}

func (r *registry[T]) Unregister(name T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.factories, name)
	delete(r.owners, name)
}

func (r *registry[T]) SetOwned(name T, owner string, factory ProviderFactoryFunc) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.factories[name]; exists && r.owners[name] != owner {
		return fmt.Errorf("provider '%v' already registered", name)
	}
	r.factories[name] = factory
	r.owners[name] = owner
	return nil
}

func (r *registry[T]) RemoveOwned(name T, owner string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.owners[name] == owner {
		delete(r.factories, name)
		delete(r.owners, name)
	}
}

func newRegistry[T comparable]() Registry[T] {
	return &registry[T]{factories: make(map[T]ProviderFactoryFunc), owners: make(map[T]string)}
}

var (
	ACMEDns01Registries  = newRegistry[domain.ACMEDns01ProviderType]()
	ACMEHttp01Registries = newRegistry[domain.ACMEHttp01ProviderType]()
)
