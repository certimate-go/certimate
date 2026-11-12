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
	RegisterEnvelope(string, *Envelope) error
	GetEnvelope(string, ...Category) (*Envelope, bool)
	ListEnvelopes() []*Envelope
	UnregisterEnvelope(string)
	ReplaceOwner(string, []*Envelope)
}

type envelopeKey struct {
	name     string
	category Category
}

type envelopeRegistry struct {
	mu            sync.RWMutex
	envelopes     map[envelopeKey]*Envelope
	contributions map[string]map[envelopeKey]*Envelope
	owners        map[envelopeKey]string
}

func (r *envelopeRegistry) RegisterEnvelope(name string, env *Envelope) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if env == nil {
		return fmt.Errorf("providerschema: envelope must not be nil")
	}
	key := envelopeKey{name, env.Category}
	if _, exists := r.envelopes[key]; exists {
		return fmt.Errorf("providerschema: envelope for %q already registered", name)
	}
	r.envelopes[key] = env
	return nil
}

func (r *envelopeRegistry) GetEnvelope(name string, categories ...Category) (*Envelope, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(categories) > 0 && categories[0] != "" {
		env, ok := r.envelopes[envelopeKey{name, categories[0]}]
		return env, ok
	}
	for _, category := range []Category{CategoryDeploy, CategoryAccess, CategoryApply, CategoryNotify, ""} {
		if env, ok := r.envelopes[envelopeKey{name, category}]; ok {
			return env, true
		}
	}
	return nil, false
}

func (r *envelopeRegistry) ListEnvelopes() []*Envelope {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Envelope, 0, len(r.envelopes))
	for _, env := range r.envelopes {
		out = append(out, env)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Provider == out[j].Provider {
			return out[i].Category < out[j].Category
		}
		return out[i].Provider < out[j].Provider
	})
	return out
}

func (r *envelopeRegistry) UnregisterEnvelope(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for key := range r.envelopes {
		if key.name == name {
			delete(r.envelopes, key)
			delete(r.owners, key)
			for _, contribution := range r.contributions {
				delete(contribution, key)
			}
		}
	}
}

func (r *envelopeRegistry) ReplaceOwner(owner string, envs []*Envelope) {
	r.mu.Lock()
	defer r.mu.Unlock()
	next := make(map[envelopeKey]*Envelope, len(envs))
	for _, env := range envs {
		next[envelopeKey{env.Provider, env.Category}] = env
	}
	if len(next) == 0 {
		delete(r.contributions, owner)
	} else {
		r.contributions[owner] = next
	}
	for key, previousOwner := range r.owners {
		if env, ok := r.contributions[previousOwner][key]; ok {
			r.envelopes[key] = env
		} else {
			delete(r.envelopes, key)
			delete(r.owners, key)
		}
	}
	names := make([]string, 0, len(r.contributions))
	for name := range r.contributions {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		for key, env := range r.contributions[name] {
			if _, ok := r.envelopes[key]; !ok {
				r.envelopes[key] = env
				r.owners[key] = name
			}
		}
	}
}

func newEnvelopeRegistry() EnvelopeRegistry {
	return &envelopeRegistry{envelopes: make(map[envelopeKey]*Envelope), contributions: make(map[string]map[envelopeKey]*Envelope), owners: make(map[envelopeKey]string)}
}

var Envelopes = newEnvelopeRegistry()
