// Package formdataregistry provides a registry for mapping entities to their CRUD operations.
package formdataregistry

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// EntityRegistration contains all operations for a single entity.
type EntityRegistration struct {
	Name string

	// CREATE operations
	DecodeNew   func(json.RawMessage) (interface{}, error)
	CreateFunc  func(context.Context, interface{}) (interface{}, error)
	CreateModel interface{} // Example instance for reflection (e.g., userapp.NewUser{})

	// UPDATE operations
	DecodeUpdate func(json.RawMessage) (interface{}, error)
	UpdateFunc   func(context.Context, uuid.UUID, interface{}) (interface{}, error)
	UpdateModel  interface{} // Example instance for reflection (e.g., userapp.UpdateUser{})
}

// Registry manages entity registrations with thread-safe lookup by name or UUID.
type Registry struct {
	mu         sync.RWMutex
	entities   map[string]*EntityRegistration
	entityByID map[uuid.UUID]*EntityRegistration
}

// New creates a new empty registry.
func New() *Registry {
	return &Registry{
		entities:   make(map[string]*EntityRegistration),
		entityByID: make(map[uuid.UUID]*EntityRegistration),
	}
}

// Register adds an entity registration by name.
func (r *Registry) Register(reg EntityRegistration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if reg.Name == "" {
		return fmt.Errorf("entity name required")
	}

	if _, exists := r.entities[reg.Name]; exists {
		return fmt.Errorf("entity %s already registered", reg.Name)
	}

	r.entities[reg.Name] = &reg
	return nil
}

// RegisterWithID registers an entity with both name and UUID lookup support.
func (r *Registry) RegisterWithID(entityID uuid.UUID, reg EntityRegistration) error {
	if err := r.Register(reg); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.entityByID[entityID]; exists {
		return fmt.Errorf("entity ID %s already registered", entityID)
	}

	r.entityByID[entityID] = &reg
	return nil
}

// Get retrieves a registration by entity name.
func (r *Registry) Get(name string) (*EntityRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reg, ok := r.entities[name]
	if !ok {
		return nil, fmt.Errorf("entity %s not registered", name)
	}
	return reg, nil
}

// GetByID retrieves a registration by entity UUID from workflow.entities.
func (r *Registry) GetByID(id uuid.UUID) (*EntityRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reg, ok := r.entityByID[id]
	if !ok {
		return nil, fmt.Errorf("entity ID %s not registered", id)
	}
	return reg, nil
}

// ListEntities returns all registered entity names.
func (r *Registry) ListEntities() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.entities))
	for name := range r.entities {
		names = append(names, name)
	}
	return names
}
