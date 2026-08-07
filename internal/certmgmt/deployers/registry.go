package deployers

import (
	"fmt"
	"sync"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
)

type ProviderFactoryFunc func(options *ProviderFactoryOptions) (core.Deployer, error)

type ProviderFactoryOptions struct {
	ProviderAccessConfig   map[string]any
	ProviderExtendedConfig map[string]any
}

type Registry[T comparable] interface {
	Register(T, ProviderFactoryFunc) error
	MustRegister(T, ProviderFactoryFunc)
	Get(T) (ProviderFactoryFunc, error)
	Unregister(T)
}

type registry[T comparable] struct {
	mu        sync.RWMutex
	factories map[T]ProviderFactoryFunc
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

func (r *registry[T]) MustRegister(name T, factory ProviderFactoryFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.factories[name]; exists {
		panic(fmt.Sprintf("provider '%v' already registered", name))
	}
	r.factories[name] = factory
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
}

func newRegistry[T comparable]() Registry[T] {
	return &registry[T]{factories: make(map[T]ProviderFactoryFunc)}
}

var Registries = newRegistry[domain.DeploymentProviderType]()
