# Progress Summary: delegate.md

## Overview
Architecture for Ichor's publish-subscribe event system. Used to fire domain events (Created/Updated/Deleted) that trigger workflows and other subscribers.

## Delegate [sdk] — `business/sdk/delegate/delegate.go`

**Responsibility:** Publish domain events to subscribers after database mutations.

### Struct
```go
type Delegate struct {
    log   *logger.Logger
    funcs map[string]map[string][]Func  // [domain][action][]handler
}

type Func func(context.Context, Data) error
```

### Key Facts
- **Thread-safe read after startup registration** — no lock on Call path (lock-free during request handling)
- **One instance shared across all domains** — wired in all.go
- **205 call sites** across 65 files in business/domain/ (verified 2026-03-09)
- **Subscribers register at startup** — via Register() calls in all.go
- **Current subscriber:** workflow DelegateHandler (business/sdk/workflow/temporal/delegatehandler.go)

### Methods
- `Register(domainType string, actionType string, fn Func)` — register subscriber at startup
- `Call(ctx context.Context, data Data) error` — fire event (205 call sites verified)

## Data [sdk] — `business/sdk/delegate/model.go`

**Event payload structure.**

```go
type Data struct {
    Domain    string
    Action    string
    RawParams []byte   // JSON-encoded event payload
}

type Func func(context.Context, Data) error
```

## StandardActionConstants [sdk]

Every domain defines three action constants:

```
ActionCreated  = "{entity}.created"
ActionUpdated  = "{entity}.updated"
ActionDeleted  = "{entity}.deleted"
```

### Event Payloads (encoded in RawParams as JSON)

- **ActionCreated** → `{ EntityID uuid, UserID uuid, Entity T }`
- **ActionUpdated** → `{ EntityID uuid, UserID uuid, Entity T, BeforeEntity T }`
- **ActionDeleted** → `{ EntityID uuid, UserID uuid, Entity T }`

### Calling Pattern

Every [bus] Create/Update/Delete calls:
```go
delegate.Call(ctx, ActionCreatedData(entity))   // after DB write succeeds
delegate.Call(ctx, ActionUpdatedData(before, after))
delegate.Call(ctx, ActionDeletedData(entity))
```

**Critical:** Events are fired AFTER the database write succeeds (ordering guarantee for workflows).

## Change Patterns

### ⚠ Adding a New Domain That Calls delegate.Call()
Affects 3 areas:
1. `business/domain/{area}/{entity}bus/{entity}bus.go` — call delegate.Call() in Create/Update/Delete
2. `business/domain/{area}/{entity}bus/{entity}bus.go` — define ActionCreated/Updated/Deleted consts + ActionCreatedData/etc helpers
3. `api/cmd/services/ichor/build/all/all.go` — pass *delegate.Delegate to NewBusiness

### ⚠ Registering a New Subscriber (Register() Call)
Affects 2 areas:
1. `api/cmd/services/ichor/build/all/all.go` — delegate.Register(domain, action, fn) for each domain/action pair
2. `{subscriber_package}/{subscriber}.go` — implement Func signature: `func(context.Context, Data) error`

### ⚠ Changing Data Struct Shape
High impact — affects 205 call sites:
1. `business/sdk/delegate/model.go` — Data struct definition
2. ALL 205 [bus] files that call delegate.Call() with ActionCreatedData/etc helpers
3. `business/sdk/workflow/temporal/delegatehandler.go` — decodes RawParams via reflection
4. **Verify first:** findReferences(business/sdk/delegate/delegate.go:48:21) — confirm exact call site count before mass edit

## Critical Points
- Delegate events are **ordered guarantees** — fired AFTER database writes (crucial for workflows)
- Payload is **JSON-encoded** — allows flexibility in subscriber implementations
- **No global locks on Call path** — read-only after startup registration (performance optimized)
- **Thread-safe maps** — registration happens once at startup, call path has no synchronization

## Notes for Future Development
The delegate pattern is the event bridge between domain mutations and workflows. Most changes will be adding new ActionCreated/Updated/Deleted calls in new domains (straightforward) rather than changing the Data struct shape (risky, 205 call sites).
