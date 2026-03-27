# Progress Summary: workflow-alerts.md

## Overview
Real-time alert delivery system. Delivers workflow alerts to users and roles via WebSocket with automatic role-change synchronization.

## Pipeline

```
create_alert ActionHandler
  ↓
RabbitMQ (QueueTypeAlert)
  ↓
AlertConsumer.handleAlert()
  ↓
AlertHub.BroadcastToUser/Role/All()
  ↓
foundation/websocket.Hub
  ↓
WebSocket clients
```

### Key Facts
- **Real-time delivery** via WebSocket
- **User- or role-scoped targeting** using string ID prefixes ("user:{uuid}" / "role:{uuid}")
- **Alert consumer still uses RabbitMQ** (QueueTypeAlert) — separate queue from main workflow engine

## AlertHub [api] — `api/domain/http/workflow/alertws/alerthub.go`

**Responsibility:** Manage alert delivery to users and roles via WebSocket.

### Struct
```go
type AlertHub struct {
    hub        *websocket.Hub
    userRoleBus *userrolebus.Business
    log        *logger.Logger
}
```

### Methods
- `NewAlertHub(hub, userRoleBus, log) *AlertHub`
- `Hub() *websocket.Hub` — access underlying hub
- `RegisterClient(ctx, client, userID uuid.UUID) error` — register user with roles
- `BroadcastToUser(userID uuid.UUID, message []byte) int` — returns delivery count
- `BroadcastToRole(roleID uuid.UUID, message []byte) int` — returns delivery count
- `BroadcastAll(message []byte) int` — returns delivery count
- `ConnectedUserIDs() []uuid.UUID` — get connected users
- `RefreshUserRoles(ctx, userID uuid.UUID) error` — update user's roles
- `UpdateClientIDsForID(ctx, id string, newIDs []string)` — bulk update clients under ID

### Key Facts
- **RegisterClient** fetches user roles via userRoleBus.QueryByUserID
- **Builds ID list:** ["user:{userID}", "role:{role0.RoleID}", "role:{role1.RoleID}", ...]
- **Calls hub.Register(ctx, client, ids)** to register the client

## AlertConsumer [api] — `api/domain/http/workflow/alertws/consumer.go`

**Responsibility:** Consume RabbitMQ alert events and broadcast to connected clients.

### Struct
```go
type AlertConsumer struct {
    alertHub *AlertHub
    wq       *rabbitmq.WorkflowQueue   // RabbitMQ connection
    log      *logger.Logger
    consumer *rabbitmq.Consumer        // set by Start()
}
```

### Methods
- `Start(ctx)` — blocks until context cancelled; consumes RabbitMQ messages
- `QueueManager.HandleMessage(ctx, msg)` — alternative direct routing

### handleAlert Routing

Based on message content:
- **msg.UserID != uuid.Nil** → BroadcastToUser(msg.UserID, bytes)
- **msg.Payload["role_id"] present** → parse UUID → BroadcastToRole(roleID, bytes)
- **else** → BroadcastAll(bytes)

### Message Types
- **"alert"** — new alert
- **"alert_updated"** — status change

### Error Handling
- Marshal/targeting errors not retried — returns nil
- Fail-open: error in one delivery doesn't block others

## AlertHubDelegateHandler [api] — `api/domain/http/workflow/alertws/delegate.go`

**Responsibility:** Keep WebSocket registrations in sync with role changes.

### Registration
- **Domain:** userrolebus.DomainName
- **Actions:** userrolebus.ActionCreated, userrolebus.ActionDeleted

### Handler Flow
```go
handleRoleChange(ctx, data)
  → unmarshal RawParams → userrolebus.ActionCreatedParams
  → extract params.Entity.UserID
  → alertHub.RefreshUserRoles(ctx, userID)
  → hub.UpdateClientIDsForID(ctx, "user:"+userID, newIDs)
```

### Key Facts
- **When user's role changes**, update WebSocket connection's registered IDs
- **Ensures role-targeted broadcasts** reach/exclude them correctly
- **Failure handling:** role fetch error logged, returns nil (don't fail business operation)

## ApprovalRequestBus [bus] — `business/domain/workflow/approvalrequestbus/approvalrequestbus.go`

**Responsibility:** CRUD for approval requests (part of alert/approval workflow).

### Struct
```go
type Business struct {
    log    *logger.Logger
    del    *delegate.Delegate
    storer Storer
}
```

### Methods
- `NewBusiness(log, del *delegate.Delegate, storer) *Business`
- `NewWithTx(tx sqldb.CommitRollbacker) (*Business, error)` — transaction-scoped copy
- `Create(ctx, na NewApprovalRequest) (ApprovalRequest, error)`
- `QueryByID(ctx, id uuid.UUID) (ApprovalRequest, error)`
- `Query(ctx, filter, orderBy, pg) ([]ApprovalRequest, error)`
- `Count(ctx, filter) (int, error)`
- `Resolve(ctx, id, resolvedBy, status, reason) (ApprovalRequest, error)` — mark resolved
- `IsApprover(ctx, approvalID, userID uuid.UUID) (bool, error)` — check if user is approver
- `ClearTaskToken(ctx, id uuid.UUID) error` — clears task_token after Temporal CompleteActivity (prevents duplicate completions on retry)

### Sentinel Error
- **ErrAlreadyResolved** — returned by Resolve() when UPDATE WHERE status='pending' matches zero rows (approval already resolved)

### Data Source
- ⊗⊕ workflow.approval_requests

## WebSocketFoundation [sdk] — `foundation/websocket/`

**Responsibility:** Low-level WebSocket message routing by ID.

### Hub Struct
```go
type Hub struct {
    clients   map[string]map[*Client]bool  // id → {client → true}
    clientIDs map[*Client][]string         // client → registered IDs
    mu        sync.RWMutex
    log       *logger.Logger
}

type Client struct {
    hub  *Hub
    conn *websocket.Conn
    ids  []string                // IDs registered under
    idMu sync.RWMutex
    send chan []byte             // capacity 256
    done chan struct{}
    log  *logger.Logger
}
```

### Hub Methods
- `NewHub(log) *Hub` — create new hub
- `Run(ctx) error` — blocking metrics loop
- `Register(ctx, client, ids []string)` — register client with IDs
- `Unregister(ctx, client)` — remove client
- `BroadcastToID(id string, message []byte) int` — send to all clients under ID
- `BroadcastAll(message []byte) int` — send to all clients
- `UpdateClientIDs(ctx, client, newIDs []string)` — change client's IDs
- `UpdateClientIDsForID(ctx, id string, newIDs []string)` — bulk update all clients under id
- `CloseAll(ctx) error` — close all connections
- `ConnectionCount() int` — count connected clients
- `ConnectedIDs() []string` — list all registered IDs
- `ClientsForID(id string) int` — count clients under ID

## Change Patterns

### ⚠ Wiring a New WebSocket Event Type
Affects 4 areas:
1. `api/domain/http/workflow/alertws/consumer.go` — add new message type handling in handleAlert
2. `api/domain/http/workflow/alertws/alerthub.go` — new Broadcast* method if new targeting scope needed
3. The ActionHandler that enqueues the event — set msg.UserID or role_id payload key
4. **Frontend WebSocket client** — handle new message type string

### ⚠ Adding a New Delegate Subscriber to AlertHub
Affects 2 areas:
1. `api/domain/http/workflow/alertws/delegate.go` — Register() call for new domain/action pair
2. `api/cmd/services/ichor/build/all/all.go` — wire new delegate registration at startup

### ⚠ Changing WebSocket Client ID Scheme
Currently: "user:{uuid}" / "role:{uuid}"

Affects 4 areas:
1. `api/domain/http/workflow/alertws/alerthub.go` — RegisterClient + RefreshUserRoles ID construction
2. `api/domain/http/workflow/alertws/consumer.go` — handleAlert routing (BroadcastToID calls)
3. `foundation/websocket/hub.go` — Hub.BroadcastToID key lookup
4. **All callers** of BroadcastToUser/BroadcastToRole — rebuild ID strings

## Critical Points
- **ID prefix scheme** — "user:{uuid}" and "role:{uuid}" are convention, not enforced
- **AlertConsumer uses RabbitMQ** — separate from main workflow engine queue
- **Role changes sync automatically** — via delegate handler + RefreshUserRoles
- **Broadcast returns delivery count** — allows testing verification
- **Fail-open design** — individual client errors don't block other deliveries

## Notes for Future Development
Alert delivery is a well-separated concern from workflow execution. Most changes will be:
- Adding new message types (straightforward, just add case in handleAlert)
- Adding new broadcast scopes (moderate, add Broadcast* method + delegate handler)
- Changing ID scheme (risky, affects multiple layers)

The WebSocket Hub is generic and reusable for other real-time features beyond alerts.
