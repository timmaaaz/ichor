# delegate

[bus]=business [app]=application [api]=HTTP [db]=store [sdk]=shared
→=depends on ⊕=writes ⊗=reads ⚡=external [tx]=transaction [cache]=cached

---

## Delegate [sdk]

file: business/sdk/delegate/delegate.go
key facts:
  - Delegate struct: log *logger.Logger, funcs map[string]map[string][]Func
  - Thread-safe read after startup registration (no lock on Call path)
  - One Delegate instance shared across all domains (wired in all.go)
  - 205 call sites across 65 files in business/domain/ (verified 2026-03-09)
  - Subscribers register at startup in all.go via Register()
  - Current subscriber: workflow DelegateHandler (business/sdk/workflow/temporal/delegatehandler.go)

  // ✓ verified 2026-03-09
  Register(domainType string, actionType string, fn Func)
<!-- lsp:hover:48:21 -->
```go
func (d *Delegate) Call(ctx context.Context, data Data) error
```
<!-- lsp:refs:48:21 --> count=205 (excl. test mocks)

---

## Data [sdk]

file: business/sdk/delegate/model.go
```go
type Data struct {
    Domain    string
    Action    string
    RawParams []byte   // JSON-encoded event payload
}

type Func func(context.Context, Data) error
```

---

## StandardActionConstants [sdk]

Every domain package defines three action constants:
  ActionCreated  = "{entity}.created"
  ActionUpdated  = "{entity}.updated"
  ActionDeleted  = "{entity}.deleted"

Payloads (encoded in RawParams):
  ActionCreated  → { EntityID uuid, UserID uuid, Entity T }
  ActionUpdated  → { EntityID uuid, UserID uuid, Entity T, BeforeEntity T }
  ActionDeleted  → { EntityID uuid, UserID uuid, Entity T }

Every [bus] Create/Update/Delete calls:
  delegate.Call(ctx, ActionCreatedData(entity))   // after DB write succeeds
  delegate.Call(ctx, ActionUpdatedData(before, after))
  delegate.Call(ctx, ActionDeletedData(entity))

---

## ⚠ Adding a new domain that calls delegate.Call()

  business/domain/{area}/{entity}bus/{entity}bus.go               (call delegate.Call() in Create/Update/Delete)
  business/domain/{area}/{entity}bus/{entity}bus.go               (define ActionCreated/Updated/Deleted consts + ActionCreatedData/etc helpers)
  api/cmd/services/ichor/build/all/all.go                         (pass *delegate.Delegate to NewBusiness)

## ⚠ Registering a new subscriber (Register() call)

  api/cmd/services/ichor/build/all/all.go                         (delegate.Register(domain, action, fn) for each domain/action pair)
  {subscriber_package}/{subscriber}.go                             (implement Func signature: func(context.Context, Data) error)

## ⚠ Changing Data struct shape

  business/sdk/delegate/model.go                                   (Data struct)
  ALL 205 [bus] files that call delegate.Call() with ActionCreatedData/etc helpers
  business/sdk/workflow/temporal/delegatehandler.go                (decodes RawParams via reflection)
  verify: findReferences(business/sdk/delegate/delegate.go:48:21) — confirm exact call site count before mass edit

