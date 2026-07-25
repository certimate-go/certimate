package providerschema

import (
	"fmt"
	"sort"
	"sync"
)

type Registry[T comparable] interface {
	Register(name T, schema *Schema) error
	MustRegister(name T, schema *Schema)
	Get(name T) (*Schema, error)
	Has(name T) bool
	List() []T
	Unregister(T)
}

type registry[T comparable] struct {
	mu      sync.RWMutex
	schemas map[T]*Schema
}

func (r *registry[T]) Register(name T, schema *Schema) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if schema == nil {
		return fmt.Errorf("providerschema: schema must not be nil")
	}
	if _, exists := r.schemas[name]; exists {
		return fmt.Errorf("providerschema: schema for %v already registered", name)
	}
	r.schemas[name] = schema
	return nil
}

func (r *registry[T]) MustRegister(name T, schema *Schema) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.schemas[name]; exists {
		panic(fmt.Sprintf("providerschema: schema for %v already registered", name))
	}
	r.schemas[name] = schema
}

func (r *registry[T]) Get(name T) (*Schema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if schema, exists := r.schemas[name]; exists {
		return schema, nil
	}
	return nil, fmt.Errorf("providerschema: schema for %v not registered", name)
}

func (r *registry[T]) Has(name T) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.schemas[name]
	return exists
}

func (r *registry[T]) List() []T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys := make([]T, 0, len(r.schemas))
	for k := range r.schemas {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return fmt.Sprint(keys[i]) < fmt.Sprint(keys[j])
	})
	return keys
}

func (r *registry[T]) Unregister(name T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.schemas, name)
}

func newRegistry[T comparable]() Registry[T] {
	return &registry[T]{schemas: make(map[T]*Schema)}
}

var Registries = newRegistry[string]()

type EnvelopeRegistry interface {
	RegisterEnvelope(name string, env *Envelope) error
	GetEnvelope(name string) (*Envelope, bool)
	ListEnvelopes() []*Envelope
	UnregisterEnvelope(name string)
}

type envelopeRegistry struct {
	mu        sync.RWMutex
	envelopes map[string]*Envelope
}

func (r *envelopeRegistry) RegisterEnvelope(name string, env *Envelope) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if env == nil {
		return fmt.Errorf("providerschema: envelope must not be nil")
	}
	if _, exists := r.envelopes[name]; exists {
		return fmt.Errorf("providerschema: envelope for %q already registered", name)
	}
	r.envelopes[name] = env
	return nil
}

func (r *envelopeRegistry) GetEnvelope(name string) (*Envelope, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	env, ok := r.envelopes[name]
	return env, ok
}

func (r *envelopeRegistry) ListEnvelopes() []*Envelope {
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys := make([]string, 0, len(r.envelopes))
	for k := range r.envelopes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]*Envelope, 0, len(keys))
	for _, k := range keys {
		out = append(out, r.envelopes[k])
	}
	return out
}

func (r *envelopeRegistry) UnregisterEnvelope(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.envelopes, name)
}

func newEnvelopeRegistry() EnvelopeRegistry {
	return &envelopeRegistry{envelopes: make(map[string]*Envelope)}
}

var Envelopes = newEnvelopeRegistry()
