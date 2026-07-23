package providerschema

import (
	"fmt"
	"sort"
)

type Registry[T comparable] interface {
	Register(name T, schema *Schema) error
	MustRegister(name T, schema *Schema)
	Get(name T) (*Schema, error)
	Has(name T) bool
	List() []T
}

type registry[T comparable] struct {
	schemas map[T]*Schema
}

func (r *registry[T]) Register(name T, schema *Schema) error {
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
	if err := r.Register(name, schema); err != nil {
		panic(err)
	}
}

func (r *registry[T]) Get(name T) (*Schema, error) {
	if schema, exists := r.schemas[name]; exists {
		return schema, nil
	}
	return nil, fmt.Errorf("providerschema: schema for %v not registered", name)
}

func (r *registry[T]) Has(name T) bool {
	_, exists := r.schemas[name]
	return exists
}

func (r *registry[T]) List() []T {
	keys := make([]T, 0, len(r.schemas))
	for k := range r.schemas {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return fmt.Sprint(keys[i]) < fmt.Sprint(keys[j])
	})
	return keys
}

func newRegistry[T comparable]() Registry[T] {
	return &registry[T]{schemas: make(map[T]*Schema)}
}

var Registries = newRegistry[string]()
