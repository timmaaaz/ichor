# FormData Registry Package

## Overview

The `formdataregistry` package provides a thread-safe registry for mapping entities to their CRUD operations. It enables dynamic form data handling while maintaining Ardan Labs architecture principles by avoiding reflection and preserving domain boundaries.

## Table of Contents

- [Architecture](#architecture)
- [Key Principles](#key-principles)
- [Usage](#usage)
- [API Reference](#api-reference)
- [Examples](#examples)
- [Best Practices](#best-practices)

## Architecture

### Problem Statement

When building dynamic forms that span multiple database tables, we need a way to:
1. Map entity names (e.g., "users") to their domain operations
2. Avoid reflection for type safety and performance
3. Preserve existing validation logic in app layers
4. Support transactions across multiple entities

### Solution: Registry Pattern

The registry pattern provides a centralized mapping of entity names to their operations:

```
┌─────────────────────────────────────────────────────────────┐
│                         Registry                            │
│                                                             │
│  "users" ──► { DecodeNew,  CreateFunc,                     │
│                DecodeUpdate, UpdateFunc }                   │
│                                                             │
│  "assets" ─► { DecodeNew,  CreateFunc,                     │
│                DecodeUpdate, UpdateFunc }                   │
│                                                             │
│  entity_uuid ──► Registration (for form field lookups)     │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

```
Frontend Request
    │
    ▼
FormData App Layer
    │
    ├─► Registry.Get("users")
    │   └─► EntityRegistration{DecodeNew, CreateFunc, ...}
    │
    ├─► DecodeNew(jsonData)
    │   └─► userapp.NewUser{} (validated)
    │
    └─► CreateFunc(ctx, model)
        └─► userapp.Create() → userbus.Create()
```

## Key Principles

### 1. No Reflection

All type conversions are explicit via registered functions:

```go
// ✅ GOOD: Explicit type conversion
DecodeNew: func(data json.RawMessage) (interface{}, error) {
    var app userapp.NewUser
    if err := json.Unmarshal(data, &app); err != nil {
        return nil, err
    }
    if err := app.Validate(); err != nil {
        return nil, err
    }
    return app, nil
}

// ❌ BAD: Using reflection
DecodeNew: func(data json.RawMessage) (interface{}, error) {
    modelType := reflect.TypeOf(userapp.NewUser{})
    model := reflect.New(modelType).Interface()
    // ... reflection-based unmarshaling
}
```

### 2. Domain Isolation

Each entity's operations remain in their domain packages. The registry only stores function pointers:

```go
// Registration happens in build.go (wiring layer)
registry.Register(EntityRegistration{
    Name: "users",
    CreateFunc: userAppInstance.Create,  // Points to existing method
    // ...
})
```

### 3. Validation Preserved

Uses existing app layer `Validate()` methods without duplication:

```go
DecodeNew: func(data json.RawMessage) (interface{}, error) {
    var app userapp.NewUser
    json.Unmarshal(data, &app)

    // Leverage existing validation
    if err := app.Validate(); err != nil {
        return nil, err
    }

    return app, nil
}
```

### 4. Transaction Safe

Compatible with Ardan Labs `NewWithTx()` pattern:

```go
// The app instance used in CreateFunc can be transactional
txUserApp, _ := userApp.NewWithTx(tx)

registry.Register(EntityRegistration{
    Name: "users",
    CreateFunc: txUserApp.Create,  // Uses transactional instance
})
```

## Usage

### Basic Registration

```go
package main

import (
    "github.com/timmaaaz/ichor/app/sdk/formdataregistry"
    "github.com/timmaaaz/ichor/app/domain/core/userapp"
)

func main() {
    // Create registry
    registry := formdataregistry.New()

    // Register users entity
    err := registry.Register(formdataregistry.EntityRegistration{
        Name: "users",

        // Decode and validate for CREATE
        DecodeNew: func(data json.RawMessage) (interface{}, error) {
            var app userapp.NewUser
            if err := json.Unmarshal(data, &app); err != nil {
                return nil, err
            }
            if err := app.Validate(); err != nil {
                return nil, err
            }
            return app, nil
        },

        // Execute CREATE
        CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
            return userAppInstance.Create(ctx, model.(userapp.NewUser))
        },

        // Decode and validate for UPDATE
        DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
            var app userapp.UpdateUser
            if err := json.Unmarshal(data, &app); err != nil {
                return nil, err
            }
            if err := app.Validate(); err != nil {
                return nil, err
            }
            return app, nil
        },

        // Execute UPDATE
        UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
            return userAppInstance.Update(ctx, model.(userapp.UpdateUser), id)
        },
    })

    if err != nil {
        log.Fatal("registration failed", err)
    }
}
```

### Registration with UUID Support

If you need to look up entities by their `workflow.entities.id` UUID:

```go
// Query workflow.entities for the UUID
usersEntityID := getEntityIDByName("users")  // Returns UUID

// Register with both name and ID
err := registry.RegisterWithID(usersEntityID, formdataregistry.EntityRegistration{
    Name: "users",
    // ... same as above
})
```

This enables lookups like:

```go
// By name (from frontend payload)
reg, _ := registry.Get("users")

// By UUID (from form field config)
reg, _ := registry.GetByID(formField.EntityID)
```

### Using Registered Operations

```go
// Look up entity
reg, err := registry.Get("users")
if err != nil {
    return fmt.Errorf("entity not found: %w", err)
}

// CREATE operation
jsonData := []byte(`{"first_name":"John","last_name":"Doe","email":"john@example.com"}`)

model, err := reg.DecodeNew(jsonData)
if err != nil {
    return fmt.Errorf("decode failed: %w", err)
}

result, err := reg.CreateFunc(ctx, model)
if err != nil {
    return fmt.Errorf("create failed: %w", err)
}

// UPDATE operation
updateData := []byte(`{"id":"uuid-here","first_name":"Jane"}`)

updateModel, err := reg.DecodeUpdate(updateData)
if err != nil {
    return fmt.Errorf("decode failed: %w", err)
}

id := uuid.MustParse("uuid-here")
result, err := reg.UpdateFunc(ctx, id, updateModel)
if err != nil {
    return fmt.Errorf("update failed: %w", err)
}
```

## API Reference

### Types

#### `EntityRegistration`

```go
type EntityRegistration struct {
    Name         string
    DecodeNew    func(json.RawMessage) (interface{}, error)
    CreateFunc   func(context.Context, interface{}) (interface{}, error)
    DecodeUpdate func(json.RawMessage) (interface{}, error)
    UpdateFunc   func(context.Context, uuid.UUID, interface{}) (interface{}, error)
}
```

**Fields:**

- `Name`: Entity name matching `workflow.entities.name`
- `DecodeNew`: Decodes and validates JSON for CREATE operations
- `CreateFunc`: Executes CREATE operation on the entity
- `DecodeUpdate`: Decodes and validates JSON for UPDATE operations
- `UpdateFunc`: Executes UPDATE operation on the entity

#### `EntityOperation`

```go
type EntityOperation string

const (
    OperationCreate EntityOperation = "create"
    OperationUpdate EntityOperation = "update"
)
```

**Methods:**

- `String() string`: Returns string representation
- `IsValid() bool`: Returns true if operation is recognized

### Registry Methods

#### `New() *Registry`

Creates a new empty registry.

**Example:**

```go
registry := formdataregistry.New()
```

#### `Register(reg EntityRegistration) error`

Registers an entity by name only.

**Parameters:**
- `reg`: The entity registration

**Returns:**
- `error`: If name is empty or already registered

**Example:**

```go
err := registry.Register(formdataregistry.EntityRegistration{
    Name: "users",
    DecodeNew: decodeNewUser,
    CreateFunc: createUser,
    DecodeUpdate: decodeUpdateUser,
    UpdateFunc: updateUser,
})
```

#### `RegisterWithID(entityID uuid.UUID, reg EntityRegistration) error`

Registers an entity with both name and UUID lookup.

**Parameters:**
- `entityID`: UUID from `workflow.entities.id`
- `reg`: The entity registration

**Returns:**
- `error`: If name/ID is empty or already registered

**Example:**

```go
entityID := uuid.MustParse("...")
err := registry.RegisterWithID(entityID, registration)
```

#### `Get(name string) (*EntityRegistration, error)`

Looks up a registration by entity name.

**Parameters:**
- `name`: Entity name (e.g., "users")

**Returns:**
- `*EntityRegistration`: The registration
- `error`: If not found

**Example:**

```go
reg, err := registry.Get("users")
```

#### `GetByID(id uuid.UUID) (*EntityRegistration, error)`

Looks up a registration by entity UUID.

**Parameters:**
- `id`: UUID from `workflow.entities.id`

**Returns:**
- `*EntityRegistration`: The registration
- `error`: If not found

**Example:**

```go
reg, err := registry.GetByID(entityUUID)
```

#### `ListEntities() []string`

Returns all registered entity names.

**Returns:**
- `[]string`: List of entity names

**Example:**

```go
entities := registry.ListEntities()
fmt.Println("Registered:", entities)  // ["users", "assets", "products"]
```

## Examples

### Complete Registration Example

```go
package build

import (
    "context"
    "encoding/json"

    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/app/sdk/formdataregistry"
    "github.com/timmaaaz/ichor/app/domain/core/userapp"
    "github.com/timmaaaz/ichor/app/domain/assets/assetapp"
)

func BuildFormDataRegistry(
    userApp *userapp.App,
    assetApp *assetapp.App,
) (*formdataregistry.Registry, error) {
    registry := formdataregistry.New()

    // Register users
    if err := registry.Register(formdataregistry.EntityRegistration{
        Name: "users",
        DecodeNew: func(data json.RawMessage) (interface{}, error) {
            var app userapp.NewUser
            if err := json.Unmarshal(data, &app); err != nil {
                return nil, err
            }
            if err := app.Validate(); err != nil {
                return nil, err
            }
            return app, nil
        },
        CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
            return userApp.Create(ctx, model.(userapp.NewUser))
        },
        DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
            var app userapp.UpdateUser
            if err := json.Unmarshal(data, &app); err != nil {
                return nil, err
            }
            if err := app.Validate(); err != nil {
                return nil, err
            }
            return app, nil
        },
        UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
            return userApp.Update(ctx, model.(userapp.UpdateUser), id)
        },
    }); err != nil {
        return nil, err
    }

    // Register assets
    if err := registry.Register(formdataregistry.EntityRegistration{
        Name: "assets",
        DecodeNew: func(data json.RawMessage) (interface{}, error) {
            var app assetapp.NewAsset
            if err := json.Unmarshal(data, &app); err != nil {
                return nil, err
            }
            if err := app.Validate(); err != nil {
                return nil, err
            }
            return app, nil
        },
        CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
            return assetApp.Create(ctx, model.(assetapp.NewAsset))
        },
        DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
            var app assetapp.UpdateAsset
            if err := json.Unmarshal(data, &app); err != nil {
                return nil, err
            }
            if err := app.Validate(); err != nil {
                return nil, err
            }
            return app, nil
        },
        UpdateFunc: func(ctx context.Context, id uuid.UUID, model interface{}) (interface{}, error) {
            return assetApp.Update(ctx, model.(assetapp.UpdateAsset), id)
        },
    }); err != nil {
        return nil, err
    }

    return registry, nil
}
```

### Error Handling Example

```go
reg, err := registry.Get("users")
if err != nil {
    // Entity not registered - this is a programming error
    // Should be caught during startup/initialization
    log.Fatal("Critical: users entity not registered", err)
    return
}

model, err := reg.DecodeNew(jsonData)
if err != nil {
    // Invalid JSON or validation failed - this is a user error
    return errs.New(errs.InvalidArgument, err)
}

result, err := reg.CreateFunc(ctx, model)
if err != nil {
    // Business logic error (e.g., duplicate email)
    return errs.NewError(err)
}
```

## Best Practices

### 1. Register at Startup

Register all entities during application startup in `build.go`:

```go
// ✅ GOOD: Register during wiring
func BuildApp() (*App, error) {
    registry := formdataregistry.New()

    // Register all entities
    registry.Register(...)
    registry.Register(...)

    // Build FormData app with registry
    formDataApp := formdataapp.NewApp(registry, ...)

    return app, nil
}

// ❌ BAD: Lazy registration during request handling
func Handler(ctx context.Context, r *http.Request) {
    registry.Register(...)  // Too late!
}
```

### 2. Validate on Registration

Ensure all required functions are provided:

```go
func RegisterEntity(registry *Registry, name string, app AppInterface) error {
    if app == nil {
        return fmt.Errorf("app instance required for %s", name)
    }

    return registry.Register(EntityRegistration{
        Name: name,
        DecodeNew: func(data json.RawMessage) (interface{}, error) {
            // ... must not be nil
        },
        CreateFunc: app.Create,
        DecodeUpdate: func(data json.RawMessage) (interface{}, error) {
            // ... must not be nil
        },
        UpdateFunc: app.Update,
    })
}
```

### 3. Use Type Assertions Safely

Always check type assertions in production code:

```go
// ✅ GOOD: Safe type assertion
CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
    newUser, ok := model.(userapp.NewUser)
    if !ok {
        return nil, fmt.Errorf("expected userapp.NewUser, got %T", model)
    }
    return userApp.Create(ctx, newUser)
}

// ⚠️  ACCEPTABLE: In well-tested registration code
CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
    return userApp.Create(ctx, model.(userapp.NewUser))
}
```

### 4. Document Entity Dependencies

Add comments explaining dependencies between entities:

```go
// Register entities in dependency order for documentation
// 1. Independent entities first
registry.Register("users", ...)      // No dependencies
registry.Register("categories", ...) // No dependencies

// 2. Dependent entities second
registry.Register("products", ...)   // Depends on categories
registry.Register("addresses", ...)  // Depends on users
```

### 5. Log Registration

Log successful registrations for debugging:

```go
entities := []string{"users", "assets", "products"}

for _, name := range entities {
    if err := registerEntity(registry, name); err != nil {
        return fmt.Errorf("register %s: %w", name, err)
    }
    log.Info("Registered entity", "name", name)
}
```

## Thread Safety

All registry methods are thread-safe:

```go
// Safe: Concurrent registration during startup
var wg sync.WaitGroup
for _, entity := range entities {
    wg.Add(1)
    go func(e Entity) {
        defer wg.Done()
        registry.Register(e.ToRegistration())
    }(entity)
}
wg.Wait()

// Safe: Concurrent lookups during request handling
go func() {
    reg, _ := registry.Get("users")
    // use reg...
}()

go func() {
    reg, _ := registry.Get("assets")
    // use reg...
}()
```

However, registration should typically happen during startup, not concurrently during request handling.

## Testing

### Unit Test Example

```go
func TestRegistry_Registration(t *testing.T) {
    registry := formdataregistry.New()

    reg := formdataregistry.EntityRegistration{
        Name: "users",
        DecodeNew: func(data json.RawMessage) (interface{}, error) {
            return "decoded", nil
        },
        CreateFunc: func(ctx context.Context, model interface{}) (interface{}, error) {
            return "created", nil
        },
    }

    // Test registration
    if err := registry.Register(reg); err != nil {
        t.Fatal(err)
    }

    // Test lookup
    found, err := registry.Get("users")
    if err != nil {
        t.Fatal(err)
    }

    if found.Name != "users" {
        t.Errorf("expected users, got %s", found.Name)
    }

    // Test duplicate registration
    if err := registry.Register(reg); err == nil {
        t.Error("expected error for duplicate registration")
    }
}
```

## Troubleshooting

### "entity X not registered"

**Cause:** Entity was not registered during startup or registration failed.

**Solution:** Check `build.go` for registration logic and ensure no errors are swallowed:

```go
if err := registry.Register(...); err != nil {
    return nil, fmt.Errorf("register users: %w", err)  // Don't ignore!
}
```

### Type Assertion Panics

**Cause:** Wrong type passed to `CreateFunc` or `UpdateFunc`.

**Solution:** Add type checking in decode functions:

```go
DecodeNew: func(data json.RawMessage) (interface{}, error) {
    var app userapp.NewUser
    if err := json.Unmarshal(data, &app); err != nil {
        return nil, fmt.Errorf("unmarshal: %w", err)
    }
    // Type is guaranteed to be userapp.NewUser here
    return app, nil
}
```

### Validation Not Running

**Cause:** Forgot to call `Validate()` in decode function.

**Solution:** Always validate before returning:

```go
DecodeNew: func(data json.RawMessage) (interface{}, error) {
    var app userapp.NewUser
    json.Unmarshal(data, &app)

    // MUST validate!
    if err := app.Validate(); err != nil {
        return nil, err
    }

    return app, nil
}
```
