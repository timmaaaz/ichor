# Layer Patterns

This document details the implementation patterns for each layer in the Ardan Labs architecture.

## Business Layer (`*bus` packages)

The business layer contains ALL domain logic. It knows nothing about HTTP, JSON, or the API layer.

### Core Structure

```go
// Core structure
type Business struct {
    log      *logger.Logger
    delegate *delegate.Delegate  // Handles UUID generation, time
    storer   Storer              // Interface to database
}

// Always expose interface for storage
type Storer interface {
    Create(ctx context.Context, entity Entity) error
    QueryByID(ctx context.Context, id uuid.UUID) (Entity, error)
    // ... other methods
}
```

### Storer Interface Pattern

Every business package defines a `Storer` interface that the database store implements:

```go
type Storer interface {
    NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
    Create(ctx context.Context, entity Entity) error
    Update(ctx context.Context, entity Entity) error
    Delete(ctx context.Context, entity Entity) error
    Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Entity, error)
    Count(ctx context.Context, filter QueryFilter) (int, error)
    QueryByID(ctx context.Context, id uuid.UUID) (Entity, error)
}
```

### Model Definitions

```go
// Main entity - returned from queries
type Entity struct {
    ID        uuid.UUID
    Name      string
    CreatedAt time.Time
}

// Creation input - no ID, no timestamps
type NewEntity struct {
    Name string
}

// Update input - all fields are pointers (nil = no change)
type UpdateEntity struct {
    Name *string
}
```

## Application Layer (`*app` packages)

The application layer validates input, converts between API and business models, and handles error translation.

### Core Structure

```go
type App struct {
    business *entitybus.Business
}

func (a *App) Create(ctx context.Context, app NewEntity) (Entity, error) {
    // 1. Validate app model
    if err := app.Validate(); err != nil {
        return Entity{}, err
    }

    // 2. Convert app → bus
    bus := toBusNewEntity(app)

    // 3. Call business layer
    busEntity, err := a.business.Create(ctx, bus)
    if err != nil {
        return Entity{}, err
    }

    // 4. Convert bus → app
    return toAppEntity(busEntity), nil
}
```

## Encoder Interface (Response Types)

All response types in the app layer must implement the `Encoder` interface:

```go
type Encoder interface {
    Encode() ([]byte, string, error)
}
```

### Single Entity Response

```go
type Entity struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

func (app Entity) Encode() ([]byte, string, error) {
    data, err := json.Marshal(app)
    return data, "application/json", err
}
```

### Paginated Query Response

For standard paginated queries, use `query.Result[T]` which already implements `Encode()`:

```go
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Entity], error) {
    // ... filter, order, page parsing ...
    entities, err := a.business.Query(ctx, filter, orderBy, pg)
    total, err := a.business.Count(ctx, filter)
    return query.NewResult(ToAppEntities(entities), total, pg), nil
}
```

### Slice Response (QueryAll, QueryByIDs, etc.)

For methods returning plain slices, create a wrapper type that implements `Encode()`:

```go
// Entities is a collection wrapper that implements the Encoder interface.
type Entities []Entity

func (app Entities) Encode() ([]byte, string, error) {
    data, err := json.Marshal(app)
    return data, "application/json", err
}

// Then use it in app methods:
func (a *App) QueryAll(ctx context.Context) (Entities, error) {
    entities, err := a.business.QueryAll(ctx)
    if err != nil {
        return nil, errs.Newf(errs.Internal, "queryall: %s", err)
    }
    return Entities(ToAppEntities(entities)), nil
}
```

### API Layer Returns

The API layer simply returns the encodable type - no need for `web.NewSliceResponse()`:

```go
func (api *api) queryAll(ctx context.Context, r *http.Request) web.Encoder {
    entities, err := api.entityApp.QueryAll(ctx)
    if err != nil {
        return errs.NewError(err)
    }
    return entities  // Already implements Encoder
}
```

## Decoder Interface (Request Types)

All request types in the app layer must implement the `Decoder` interface:

```go
type Decoder interface {
    Decode(data []byte) error
}
```

### Standard Request Models

All `New*` and `Update*` types already implement `Decode()` and `Validate()`:

```go
type NewEntity struct {
    Name string `json:"name" validate:"required"`
}

func (app *NewEntity) Decode(data []byte) error {
    return json.Unmarshal(data, &app)
}

func (app NewEntity) Validate() error {
    if err := errs.Check(app); err != nil {
        return errs.Newf(errs.InvalidArgument, "validate: %s", err)
    }
    return nil
}
```

### Custom Request Models (e.g., Batch Operations)

For special endpoints like batch queries, create dedicated request types:

```go
// QueryByIDsRequest represents a batch query by IDs.
type QueryByIDsRequest struct {
    IDs []string `json:"ids" validate:"required,min=1"`
}

func (app *QueryByIDsRequest) Decode(data []byte) error {
    return json.Unmarshal(data, &app)
}

func (app QueryByIDsRequest) Validate() error {
    if err := errs.Check(app); err != nil {
        return errs.Newf(errs.InvalidArgument, "validate: %s", err)
    }
    return nil
}
```

## API Layer (`*api` packages)

The API layer handles HTTP concerns only: routing, middleware, request/response serialization.

### Handler Pattern

```go
// HTTP handlers ONLY
func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
    var app appModel.NewEntity
    if err := web.Decode(r, &app); err != nil {
        return errs.NewError(errs.InvalidArgument, err)
    }

    entity, err := api.entityApp.Create(ctx, app)
    if err != nil {
        return errs.NewError(errs.Internal, err)
    }

    return entity
}
```

### API Layer Usage Rules

**ALWAYS** decode directly into app layer models - **NEVER** create local request structs in the API layer:

```go
// CORRECT - Decode into app layer model
func (api *api) queryByIDs(ctx context.Context, r *http.Request) web.Encoder {
    var app entityapp.QueryByIDsRequest
    if err := web.Decode(r, &app); err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    entities, err := api.entityApp.QueryByIDs(ctx, app.IDs)
    if err != nil {
        return errs.NewError(err)
    }
    return entities
}

// WRONG - Don't create local request structs
func (api *api) queryByIDs(ctx context.Context, r *http.Request) web.Encoder {
    type request struct {  // ❌ NEVER DO THIS
        IDs []string `json:"ids"`
    }
    var req request  // ❌ WRONG
    // ...
}
```

## Model Conversion Functions

### Naming Convention

- `toBus*()` - Convert app model to business model (private)
- `toApp*()` or `ToApp*()` - Convert business model to app model (public if needed externally)

### Example Conversions

```go
// Private - only used within app package
func toBusNewEntity(app NewEntity) entitybus.NewEntity {
    return entitybus.NewEntity{
        Name: app.Name,
    }
}

// Public - may be used by other packages
func ToAppEntity(bus entitybus.Entity) Entity {
    return Entity{
        ID:   bus.ID.String(),
        Name: bus.Name,
    }
}

func ToAppEntities(bus []entitybus.Entity) []Entity {
    entities := make([]Entity, len(bus))
    for i, b := range bus {
        entities[i] = ToAppEntity(b)
    }
    return entities
}
```

## Summary

| Layer | Imports | Responsibility |
|-------|---------|----------------|
| API | app, business, foundation | HTTP routing, middleware, serialization |
| App | business, foundation | Validation, conversion, error translation |
| Business | foundation only | Domain logic, database interface |
| Foundation | nothing internal | Logging, tracing, web utilities |
