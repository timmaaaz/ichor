# Worker Deployment

This document covers the `workflow-worker` service — a separate process that executes Temporal workflow tasks.

## Overview

The workflow-worker runs independently from the main `ichor` API service. This separation allows:
- **Independent scaling**: Add more workers without scaling the API
- **Resource isolation**: Worker CPU/memory usage doesn't affect API latency
- **Independent deployment**: Update worker code without restarting the API

## Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `ICHOR_TEMPORAL_HOSTPORT` | Temporal server address (e.g., `temporal:7233`) | Yes |
| `ICHOR_DB_HOST` | PostgreSQL host | Yes |
| `ICHOR_DB_USER` | Database user | Yes |
| `ICHOR_DB_PASSWORD` | Database password | Yes |
| `ICHOR_DB_NAME` | Database name (default: `ichor`) | No |
| `ICHOR_DB_DISABLE_TLS` | Disable TLS for DB connection | No |

The worker connects to the **same database** as the main service. No separate database is needed.

No RabbitMQ connection is required for the worker.

## K8s Deployment

### Manifest Locations

| Purpose | Path |
|---------|------|
| Base deployment + service | `zarf/k8s/base/workflow-worker/base-workflow-worker.yaml` |
| Base kustomization | `zarf/k8s/base/workflow-worker/kustomization.yaml` |
| Dev overlay kustomization | `zarf/k8s/dev/workflow-worker/kustomization.yaml` |
| Dev ConfigMap | `zarf/k8s/dev/workflow-worker/dev-workflow-worker-configmap.yaml` |
| Dev deployment patch | `zarf/k8s/dev/workflow-worker/dev-workflow-worker-patch-deploy.yaml` |

### Dev ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: workflow-worker-config
data:
  ICHOR_TEMPORAL_HOSTPORT: "temporal:7233"
  ICHOR_DB_HOST: "database-service.ichor-system.svc.cluster.local"
  ICHOR_DB_USER: "postgres"
  ICHOR_DB_PASSWORD: "postgres"
  ICHOR_DB_NAME: "ichor"
  ICHOR_DB_DISABLE_TLS: "true"
```

### Temporal Server Deployment

Temporal runs as a separate deployment with a UI sidecar:

| Path | Purpose |
|------|---------|
| `zarf/k8s/dev/temporal/dev-temporal.yaml` | Temporal server + UI sidecar |
| `zarf/k8s/dev/temporal/kustomization.yaml` | Kustomization |

The Temporal server uses `temporalio/auto-setup` with PostgreSQL backend (shared database).

## Action Handler Registration

The worker registers action handlers at startup:

### RegisterCoreActions (Current)

5 handlers that have no external dependencies beyond the database:

| Handler | Action Type |
|---------|-------------|
| EvaluateConditionHandler | `evaluate_condition` |
| UpdateFieldHandler | `update_field` |
| SeekApprovalHandler | `seek_approval` |
| SendEmailHandler | `send_email` |
| SendNotificationHandler | `send_notification` |

### RegisterAll (Future)

Full set of handlers including those requiring additional dependencies:

| Handler | Action Type | Extra Dependencies |
|---------|-------------|--------------------|
| CreateAlertHandler | `create_alert` | alertBus |
| AllocateInventoryHandler | `allocate_inventory` | inventoryBus, allocationBus |

These will be added when the worker gains access to the required business layer dependencies.

### AsyncRegistry

An empty `AsyncRegistry` is registered at startup. If an async workflow is dispatched, it will fail gracefully with a clear error message. Full async support requires `StartAsync` adapters for `SendEmailHandler` and `AllocateInventoryHandler`.

## Scaling

### Horizontal Scaling

Multiple worker instances can poll the same `ichor-workflow` task queue:

```yaml
spec:
  replicas: 3  # Scale up workers
```

Temporal automatically distributes tasks across available workers. No coordination needed.

### Considerations

- Each worker opens a database connection pool
- Each worker maintains a gRPC connection to Temporal
- Workers are stateless — safe to scale up/down at any time
- Temporal handles at-most-once delivery per task

## Monitoring

### Temporal UI

Access the Temporal Web UI to inspect workflows:

```bash
make temporal-ui
# Opens port-forward to http://localhost:8280
```

The UI provides:
- Running/completed/failed workflow list
- Workflow history timeline
- Activity input/output inspection
- Search by workflow ID prefix (e.g., `workflow-{ruleID}`)

### Worker Logs

```bash
# View worker logs
make dev-logs-workflow-worker

# Describe worker pods
make dev-describe-workflow-worker
```

### Pod Status

```bash
make dev-status
# Look for workflow-worker pods in the output
```

## Graceful Shutdown

The worker handles shutdown signals gracefully:

```go
// Signal handling
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()

// Worker uses InterruptCh for graceful drain
w.Run(worker.InterruptCh())
```

On shutdown:
1. Worker stops polling for new tasks
2. In-progress activities complete (with timeout)
3. Database connections are closed
4. Temporal client is closed

Temporal automatically reassigns any incomplete tasks to other workers.

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make dev-logs-workflow-worker` | Stream worker pod logs |
| `make dev-describe-workflow-worker` | Describe worker pods |
| `make temporal-ui` | Port-forward Temporal UI to localhost:8280 |
| `make dev-status` | Show all pod statuses (includes worker) |

## Dockerfile

**Location**: `zarf/docker/dockerfile.workflow-worker`

The worker uses the same base image pattern as the main service, building from the Go module root.

## Related Documentation

- [Architecture](architecture.md) — System overview, trigger and execution sides
- [Temporal Integration](temporal.md) — Temporal design, activities, error handling
- [Migration from RabbitMQ](migration-from-rabbitmq.md) — What changed from old engine
