package certifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
)

type ProviderFactoryFunc func(options *ProviderFactoryOptions) (core.ACMEChallenger, error)

type ProviderFactoryOptions struct {
	ProviderAccessConfig   map[string]any
	ProviderExtendedConfig map[string]any
	DnsPropagationTimeout  int
	DnsTTL                 int
}

type Registry[T comparable] interface {
	Register(T, ProviderFactoryFunc) error
	RegisterAlias(T, T) error
	MustRegister(T, ProviderFactoryFunc)
	MustRegisterAlias(T, T)
	Get(T) (ProviderFactoryFunc, error)
}

type registry[T comparable] struct {
	factories map[T]ProviderFactoryFunc
}

func (r *registry[T]) Register(name T, factory ProviderFactoryFunc) error {
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
	if err := r.Register(name, factory); err != nil {
		panic(err)
	}
}

func (r *registry[T]) MustRegisterAlias(name T, alias T) {
	if err := r.RegisterAlias(name, alias); err != nil {
		panic(err)
	}
}

func (r *registry[T]) Get(name T) (ProviderFactoryFunc, error) {
	if factory, exists := r.factories[name]; exists {
		return factory, nil
	}

	return nil, fmt.Errorf("provider '%v' not registered", name)
}

func newRegistry[T comparable]() Registry[T] {
	return &registry[T]{factories: make(map[T]ProviderFactoryFunc)}
}

var (
	ACMEDns01Registries  = newRegistry[domain.ACMEDns01ProviderType]()
	ACMEHttp01Registries = newRegistry[domain.ACMEHttp01ProviderType]()
)
