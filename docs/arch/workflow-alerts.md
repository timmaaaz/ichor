# workflow-alerts

[bus]=business [app]=application [api]=HTTP [db]=store [sdk]=shared
→=depends on ⊕=writes ⊗=reads ⚡=external [tx]=transaction [cache]=cached

---

## Pipeline

create_alert ActionHandler → RabbitMQ (QueueTypeAlert) → AlertConsumer.handleAlert()
      → AlertHub.BroadcastToUser/Role/All() → foundation/websocket.Hub → WebSocket clients

key facts:
  - Real-time alert delivery via WebSocket
  - Targeting is user- or role-scoped using string ID prefixes ("user:{uuid}" / "role:{uuid}")
  - Alert consumer still uses RabbitMQ (QueueTypeAlert) — separate from main workflow engine queue

---

## AlertHub [api]

file: api/domain/http/workflow/alertws/alerthub.go
```go
type AlertHub struct {
    hub        *websocket.Hub
    userRoleBus *userrolebus.Business
    log        *logger.Logger
}
```

  NewAlertHub(hub, userRoleBus, log) *AlertHub
  Hub() *websocket.Hub
  RegisterClient(ctx, client, userID uuid.UUID) error
  BroadcastToUser(userID uuid.UUID, message []byte) int   // returns delivery count
  BroadcastToRole(roleID uuid.UUID, message []byte) int   // returns delivery count
  BroadcastAll(message []byte) int                        // returns delivery count
  ConnectedUserIDs() []uuid.UUID
  RefreshUserRoles(ctx, userID uuid.UUID) error

key facts:
  - RegisterClient fetches user roles via userRoleBus.QueryByUserID
  - Builds ID list: ["user:{userID}", "role:{role0.RoleID}", "role:{role1.RoleID}", ...]
  - Calls hub.Register(ctx, client, ids)

---

## AlertConsumer [api]

file: api/domain/http/workflow/alertws/consumer.go
```go
type AlertConsumer struct {
    alertHub *AlertHub
    wq       *rabbitmq.WorkflowQueue   // RabbitMQ connection — alert queue path still uses MQ
    log      *logger.Logger
    consumer *rabbitmq.Consumer        // set by Start()
}
```

key facts:
  - event mechanism: RabbitMQ pull via wq.Consume(ctx, rabbitmq.QueueTypeAlert, handleAlert)
  - activation: Start(ctx) blocks until context cancelled; or QueueManager.HandleMessage(ctx, msg) for direct routing

handleAlert routing:
  msg.UserID != uuid.Nil            → BroadcastToUser(msg.UserID, bytes)
  msg.Payload["role_id"] present    → parse UUID → BroadcastToRole(roleID, bytes)
  else                              → BroadcastAll(bytes)

message types: "alert" (new alert), "alert_updated" (status change)
error handling: marshal/targeting errors not retried — returns nil

---

## AlertHubDelegateHandler [api]

file: api/domain/http/workflow/alertws/delegate.go
registers for:
  domain: userrolebus.DomainName
  actions: userrolebus.ActionCreated, userrolebus.ActionDeleted

handler: handleRoleChange(ctx, data)
  → unmarshal RawParams → userrolebus.ActionCreatedParms
  → extract params.Entity.UserID
  → alertHub.RefreshUserRoles(ctx, userID)
  → hub.UpdateClientIDsForID(ctx, "user:"+userID, newIDs)

key facts:
  - When a user's role changes, update their WebSocket connection's registered IDs
  - Ensures role-targeted broadcasts reach/exclude them correctly
  - failure: role fetch error → log error, return nil (don't fail business operation)

---

## ApprovalRequestBus [bus]

file: business/domain/workflow/approvalrequestbus/approvalrequestbus.go
```go
type Business struct {
    log    *logger.Logger
    del    *delegate.Delegate
    storer Storer
}
```

  NewBusiness(log, del *delegate.Delegate, storer) *Business
  NewWithTx(tx sqldb.CommitRollbacker) (*Business, error)
  Create(ctx, na NewApprovalRequest) (ApprovalRequest, error)
  QueryByID(ctx, id uuid.UUID) (ApprovalRequest, error)
  Query(ctx, filter, orderBy, pg) ([]ApprovalRequest, error)
  Count(ctx, filter) (int, error)
  Resolve(ctx, id, resolvedBy, status, reason) (ApprovalRequest, error)
  IsApprover(ctx, approvalID, userID uuid.UUID) (bool, error)
  ClearTaskToken(ctx, id uuid.UUID) error  — clears task_token after successful Temporal CompleteActivity (prevents duplicate completions on retry)

sentinel error:
  ErrAlreadyResolved — returned by Resolve() when UPDATE WHERE status='pending' matches
  zero rows (approval already resolved); Resolve() wraps and propagates unchanged

⊗⊕ workflow.approval_requests

---

## WebSocketFoundation [sdk]

file: foundation/websocket/
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
    ids  []string         // IDs registered under (protected by idMu)
    idMu sync.RWMutex
    send chan []byte       // capacity 256; drops on buffer full or closing
    done chan struct{}
    log  *logger.Logger
}
```

Hub api:
  NewHub(log) *Hub
  Run(ctx) error                                    // blocking metrics loop
  Register(ctx, client, ids []string)
  Unregister(ctx, client)
  BroadcastToID(id string, message []byte) int
  BroadcastAll(message []byte) int
  UpdateClientIDs(ctx, client, newIDs []string)
  UpdateClientIDsForID(ctx, id string, newIDs []string)  // bulk update all clients under id
  CloseAll(ctx) error
  ConnectionCount() int
  ConnectedIDs() []string
  ClientsForID(id string) int

---

## ⚠ Wiring a new WebSocket event type

  api/domain/http/workflow/alertws/consumer.go       (add new message type handling in handleAlert)
  api/domain/http/workflow/alertws/alerthub.go       (new Broadcast* method if new targeting scope needed)
  The ActionHandler that enqueues the event           (set msg.UserID or role_id payload key)
  frontend WebSocket client                           (handle new message type string)

## ⚠ Adding a new delegate subscriber to AlertHub

  api/domain/http/workflow/alertws/delegate.go       (Register() call for new domain/action pair)
  api/cmd/services/ichor/build/all/all.go            (wire new delegate registration at startup)

## ⚠ Changing WebSocket Client ID scheme (currently "user:{uuid}" / "role:{uuid}")

  api/domain/http/workflow/alertws/alerthub.go       (RegisterClient + RefreshUserRoles ID construction)
  api/domain/http/workflow/alertws/consumer.go       (handleAlert routing — BroadcastToID calls)
  foundation/websocket/hub.go                        (Hub.BroadcastToID key lookup)
  All callers of BroadcastToUser/BroadcastToRole     (rebuild ID strings)
