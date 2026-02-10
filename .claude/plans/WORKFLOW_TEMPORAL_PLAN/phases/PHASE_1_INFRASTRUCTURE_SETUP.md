# Phase 1: Infrastructure Setup

**Category**: infrastructure
**Status**: Pending
**Dependencies**: None (this phase is foundational - blocks ALL other phases)

---

## Overview

Set up Temporal server, workflow-worker service skeleton, test container infrastructure, and integrate everything with the existing Ardan Labs K8s local development workflow (`make dev-bounce`). This phase builds directly into the existing `zarf/` layer infrastructure, following the established patterns for RabbitMQ, database, auth, and the ichor service.

Developers should be able to `make dev-brew && make dev-docker && make dev-up && make dev-update-apply` with Temporal included, exactly as they do today with the existing services.

## Goals

1. **Deploy Temporal to the local K8s cluster** - Add Temporal server manifests to `zarf/k8s/dev/temporal/` using PostgreSQL (the project's existing database), integrated into `make dev-apply` and `make dev-bounce`
2. **Create workflow-worker service skeleton** - Build the base/dev K8s manifests, Dockerfile, and Makefile targets for a new `workflow-worker` service following the `ichor` service pattern (`zarf/k8s/base/`, `zarf/k8s/dev/`, `zarf/docker/`)
3. **Build test container infrastructure** - Create `foundation/temporal/` package following the `foundation/rabbitmq/rabbitmq.go` singleton container pattern with `StartTemporal()`, `GetTestContainer(t)`, and `waitForReady()`

## Prerequisites

- Docker installed with at least 4 CPUs
- KIND, kubectl, kustomize installed (`make dev-brew`)
- Existing dev cluster operational (`make dev-up` works)

---

## Critical Design Decisions

### 1. Temporal Storage Backend: PostgreSQL

**Decision**: Use the project's existing PostgreSQL instance for Temporal's persistence storage.

**Rationale**:
- The project already runs PostgreSQL 16.4 in the K8s dev cluster (`database-service`)
- Production will use PostgreSQL - dev-prod parity is important
- Temporal creates its own separate databases (`temporal` and `temporal_visibility`) - it does NOT touch application schemas (`core.*`, `workflow.*`, `sales.*`, etc.)
- The `temporalio/auto-setup` image handles schema creation automatically on first boot

**Image**: `temporalio/auto-setup:1.26.2` (pin to stable version; verify latest on [Docker Hub](https://hub.docker.com/r/temporalio/auto-setup/tags))

**Key environment variables**:
- `DB=postgres12` (NOT `postgresql` - the valid values are `postgres12` or `postgres12_pgx`)
- `DB_PORT=5432`
- `POSTGRES_USER=postgres`
- `POSTGRES_PWD=postgres`
- `POSTGRES_SEEDS=database-service` (K8s DNS name of existing PostgreSQL)

### 2. Temporal Web UI: Sidecar Container

**Decision**: Deploy `temporalio/ui` as a sidecar container in the same pod as the Temporal server.

**Rationale**:
- `temporalio/auto-setup` does NOT include the Web UI (unlike the `temporalio/temporal` dev server)
- Running UI as a sidecar in the same pod keeps the deployment simple (single YAML, similar to RabbitMQ pattern)
- UI connects to `localhost:7233` within the pod (no service discovery needed)
- Exposed on port 8080 (Temporal UI default)

**Image**: `temporalio/ui:2.34.0` (pin to stable version; verify latest on [Docker Hub](https://hub.docker.com/r/temporalio/ui/tags))

### 3. Test Container: SQLite Dev Server

**Decision**: Test containers (`foundation/temporal/`) use `temporalio/temporal` with `server start-dev` (embedded SQLite), NOT PostgreSQL.

**Rationale**:
- Test containers must be self-contained (no coordination with external database)
- Follows the `foundation/rabbitmq/rabbitmq.go` pattern where each test infrastructure is independent
- The SQLite dev server starts faster and is simpler for CI/CD
- The graph executor, workflow logic, and activities being tested don't depend on which database Temporal uses internally

---

## Task Breakdown

### Task 1: Create Temporal K8s Manifests

**Status**: Pending

**Description**: Create Temporal server deployment manifests following the RabbitMQ pattern in `zarf/k8s/dev/rabbitmq/`. Uses `temporalio/auto-setup` with PostgreSQL backend and `temporalio/ui` sidecar for the Web UI.

**Notes**:
- Must integrate with `make dev-bounce` workflow
- Add to `dev-apply` in Makefile
- Use `temporalio/auto-setup:1.26.2` pointed at existing `database-service` PostgreSQL
- Use `temporalio/ui:2.34.0` as sidecar container in same pod
- Expose Temporal gRPC port (7233) for worker connections
- Expose Temporal UI port (8080) for visibility
- Follow the RabbitMQ pattern: single deployment + service in one YAML
- **Must come after database in dev-apply** (Temporal needs PostgreSQL to be ready)
- Deployment (NOT StatefulSet) - Temporal's data lives in PostgreSQL, not local storage
- Auto-setup creates `temporal` and `temporal_visibility` databases on first boot

**Files**:
- `zarf/k8s/dev/temporal/kustomization.yaml`
- `zarf/k8s/dev/temporal/dev-temporal.yaml` (deployment + service)

**Implementation Guide**:

The kustomization follows the exact RabbitMQ pattern:
```yaml
# zarf/k8s/dev/temporal/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ./dev-temporal.yaml
```

The deployment mirrors `zarf/k8s/dev/rabbitmq/dev-rabbitmq.yaml` but with two containers (server + UI sidecar):
```yaml
# zarf/k8s/dev/temporal/dev-temporal.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: ichor-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: temporal
  namespace: ichor-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: temporal
  template:
    metadata:
      labels:
        app: temporal
    spec:
      terminationGracePeriodSeconds: 30
      containers:
        # Temporal Server with auto-setup (creates schemas on first boot)
        - name: temporal
          image: temporalio/auto-setup:1.26.2
          ports:
            - name: grpc
              containerPort: 7233
          resources:
            requests:
              cpu: 100m
            limits:
              cpu: 1000m
          env:
            - name: DB
              value: "postgres12"
            - name: DB_PORT
              value: "5432"
            - name: POSTGRES_USER
              value: "postgres"
            - name: POSTGRES_PWD
              value: "postgres"
            - name: POSTGRES_SEEDS
              value: "database-service"
          livenessProbe:
            tcpSocket:
              port: grpc
            initialDelaySeconds: 60
            periodSeconds: 30
            timeoutSeconds: 5
          readinessProbe:
            tcpSocket:
              port: grpc
            initialDelaySeconds: 30
            periodSeconds: 10
            timeoutSeconds: 5

        # Temporal Web UI (sidecar - connects to server via localhost)
        - name: temporal-ui
          image: temporalio/ui:2.34.0
          ports:
            - name: ui
              containerPort: 8080
          resources:
            requests:
              cpu: 50m
            limits:
              cpu: 200m
          env:
            - name: TEMPORAL_ADDRESS
              value: "localhost:7233"
            - name: TEMPORAL_CORS_ORIGINS
              value: "http://localhost:3000"
          livenessProbe:
            httpGet:
              path: /
              port: ui
            initialDelaySeconds: 30
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /
              port: ui
            initialDelaySeconds: 15
            periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: temporal-service
  namespace: ichor-system
spec:
  type: ClusterIP
  selector:
    app: temporal
  ports:
    - name: grpc
      port: 7233
      targetPort: grpc
    - name: ui
      port: 8080
      targetPort: ui
```

**Key design details**:
- **Two containers in one pod**: Temporal server + UI sidecar. UI connects to `localhost:7233` (same pod).
- **`DB=postgres12`**: The valid PostgreSQL driver value. NOT `postgresql` or `postgres` (those are invalid and will fail).
- **`POSTGRES_SEEDS=database-service`**: Points to existing PostgreSQL via K8s DNS.
- **Higher `initialDelaySeconds`**: Auto-setup runs schema migrations on first boot, which takes longer than a plain server start. Liveness: 60s, Readiness: 30s.
- **UI on port 8080**: Default Temporal UI port. Accessible via `temporal-service:8080` or `localhost:8080` with hostNetwork.
- **Service name: `temporal-service`**: Follows the `{app}-service` naming convention.

---

### Task 2: Create Workflow-Worker Base K8s Manifests

**Status**: Pending

**Description**: Create base K8s manifests for the workflow-worker service following the pattern in `zarf/k8s/base/ichor/`. The base manifests define the deployment structure that gets patched in dev/staging/prod overlays.

**Notes**:
- Follow `zarf/k8s/base/ichor/` pattern exactly
- The workflow-worker needs database access (for EdgeStore queries) but NOT the full ichor init container (no migrations/seeds from worker)
- No HTTP service needed (worker connects outbound to Temporal, not inbound)
- Needs configmap references for DB connection and Temporal host

**Files**:
- `zarf/k8s/base/workflow-worker/base-workflow-worker.yaml`
- `zarf/k8s/base/workflow-worker/kustomization.yaml`

**Implementation Guide**:

The base deployment mirrors `zarf/k8s/base/ichor/base-ichor.yaml` but:
- No init container (worker doesn't run migrations)
- No HTTP ports (worker is not an HTTP server)
- Single container: `workflow-worker`
- Image reference: `workflow-worker-image` (patched by kustomize)
- Environment variables: DB config from configmap + Temporal host + K8s metadata
- terminationGracePeriodSeconds: 60 (allow inflight workflows to complete)

```yaml
# zarf/k8s/base/workflow-worker/base-workflow-worker.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: ichor-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: workflow-worker
  namespace: ichor-system
spec:
  selector:
    matchLabels:
      app: workflow-worker
  template:
    metadata:
      labels:
        app: workflow-worker
    spec:
      terminationGracePeriodSeconds: 60
      containers:
        - name: workflow-worker
          image: workflow-worker-image
          env:
            - name: ICHOR_DB_USER
              valueFrom:
                configMapKeyRef:
                  name: app-config
                  key: db_user
                  optional: true
            - name: ICHOR_DB_PASSWORD
              valueFrom:
                configMapKeyRef:
                  name: app-config
                  key: db_password
                  optional: true
            - name: ICHOR_DB_HOST
              valueFrom:
                configMapKeyRef:
                  name: app-config
                  key: db_host
                  optional: true
            - name: ICHOR_DB_DISABLE_TLS
              valueFrom:
                configMapKeyRef:
                  name: app-config
                  key: db_disabletls
                  optional: true
            - name: ICHOR_TEMPORAL_HOST
              valueFrom:
                configMapKeyRef:
                  name: workflow-worker-config
                  key: temporal_host
                  optional: true
            - name: ICHOR_TEMPORAL_NAMESPACE
              valueFrom:
                configMapKeyRef:
                  name: workflow-worker-config
                  key: temporal_namespace
                  optional: true
            - name: KUBERNETES_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: KUBERNETES_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: KUBERNETES_POD_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
            - name: KUBERNETES_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
```

The kustomization.yaml:
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ./base-workflow-worker.yaml
```

---

### Task 3: Create Workflow-Worker Dev K8s Manifests

**Status**: Pending

**Description**: Create dev overlay manifests for the workflow-worker following the pattern in `zarf/k8s/dev/ichor/`. The dev overlay patches the base with dev-specific config (resource limits, host networking, configmap).

**Notes**:
- Follow `zarf/k8s/dev/ichor/` pattern
- Configmap should reference `temporal-service:7233` as Temporal host (K8s DNS)
- Use the same `app-config` configmap for database credentials (shared with ichor)
- Separate `workflow-worker-config` configmap for Temporal-specific settings
- hostNetwork: true for dev (same as ichor)
- dnsPolicy: ClusterFirstWithHostNet (same as ichor dev patch)

**Files**:
- `zarf/k8s/dev/workflow-worker/kustomization.yaml`
- `zarf/k8s/dev/workflow-worker/dev-workflow-worker-configmap.yaml`
- `zarf/k8s/dev/workflow-worker/dev-workflow-worker-patch-deploy.yaml`

**Implementation Guide**:

Kustomization references base and applies patches (mirrors `zarf/k8s/dev/ichor/kustomization.yaml`):
```yaml
# zarf/k8s/dev/workflow-worker/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../../base/workflow-worker/
  - ./dev-workflow-worker-configmap.yaml
patches:
  - path: ./dev-workflow-worker-patch-deploy.yaml
images:
  - name: workflow-worker-image
    newName: localhost/superior/workflow-worker
    newTag: 0.0.1
```

Configmap for Temporal-specific config:
```yaml
# zarf/k8s/dev/workflow-worker/dev-workflow-worker-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: workflow-worker-config
  namespace: ichor-system
data:
  temporal_host: "temporal-service:7233"
  temporal_namespace: "default"
```

Patch sets dev resource limits and networking (mirrors `zarf/k8s/dev/ichor/dev-ichor-patch-deploy.yaml`):
```yaml
# zarf/k8s/dev/workflow-worker/dev-workflow-worker-patch-deploy.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: workflow-worker
  namespace: ichor-system
spec:
  replicas: 1
  strategy:
    type: Recreate
  template:
    spec:
      dnsPolicy: ClusterFirstWithHostNet
      hostNetwork: true
      containers:
        - name: workflow-worker
          resources:
            requests:
              cpu: "250m"
              memory: "36Mi"
            limits:
              cpu: "250m"
              memory: "36Mi"
```

---

### Task 4: Create Workflow-Worker Dockerfile

**Status**: Pending

**Description**: Create the Dockerfile for the workflow-worker service following `zarf/docker/dockerfile.ichor` pattern. Multi-stage build: Go build stage -> Alpine runtime.

**Notes**:
- Follow `zarf/docker/dockerfile.ichor` pattern exactly
- Build path: `api/cmd/services/workflow-worker`
- No admin binary needed (worker doesn't run migrations)
- No RSA keys needed (worker doesn't issue tokens)
- Keep it minimal: just the worker binary

**Files**:
- `zarf/docker/dockerfile.workflow-worker`

**Implementation Guide**:

```dockerfile
# Build the Go Binary.
FROM golang:1.23 AS build_workflow_worker
ENV CGO_ENABLED=0
ARG BUILD_REF

COPY . /service

# Build the worker binary.
WORKDIR /service/api/cmd/services/workflow-worker
RUN go build -ldflags "-X main.build=${BUILD_REF}"

# Run the Go Binary in Alpine.
FROM alpine:3.20
ARG BUILD_DATE
ARG BUILD_REF
RUN addgroup -g 1000 -S ichor && \
    adduser -u 1000 -h /service -G ichor -S ichor
COPY --from=build_workflow_worker --chown=ichor:ichor /service/api/cmd/services/workflow-worker/workflow-worker /service/workflow-worker
WORKDIR /service
USER ichor
CMD ["./workflow-worker"]

LABEL org.opencontainers.image.created="${BUILD_DATE}" \
    org.opencontainers.image.title="workflow-worker" \
    org.opencontainers.image.authors="Superior ERP" \
    org.opencontainers.image.revision="${BUILD_REF}" \
    org.opencontainers.image.vendor="Ardan Labs"
```

---

### Task 5: Create Temporal Test Container Infrastructure

**Status**: Pending

**Description**: Create `foundation/temporal/` package following the `foundation/rabbitmq/rabbitmq.go` singleton container pattern. This provides test container management for integration tests throughout the codebase.

**Notes**:
- Follow pattern established in `foundation/rabbitmq/rabbitmq.go` **exactly**
- **Uses `temporalio/temporal` with `server start-dev` (SQLite)** - NOT the PostgreSQL auto-setup image
  - Test containers must be self-contained (no coordination with external database)
  - Follows the rabbitmq.go pattern where each test infrastructure is independent
  - SQLite dev server starts faster and is simpler for CI/CD
  - Workflow logic being tested doesn't depend on Temporal's internal storage backend
- `StartTemporal()` - starts container, returns Container with HostPort
- `GetTestContainer(t)` - returns shared container (singleton pattern with sync.Mutex)
- `NewTestClient(hostPort)` - creates a Temporal client for testing
- `Container` struct with `HostPort` field
- `waitForReady()` uses Temporal Go SDK `client.Dial` + `client.CheckHealth` to verify readiness
- Uses `foundation/docker` package for `StartContainer`/`StopContainer`
- Container name: `servicetest-temporal` (matches `servicetest-rabbit` pattern)
- Port: `7233` (Temporal gRPC)

**Files**:
- `foundation/temporal/temporal.go`
- `foundation/temporal/temporal_test.go`

**Implementation Guide**:

The package mirrors `foundation/rabbitmq/rabbitmq.go` structure exactly:

```go
package temporal

import (
    "context"
    "fmt"
    "strings"
    "sync"
    "testing"
    "time"

    "github.com/timmaaaz/ichor/foundation/docker"
    "go.temporal.io/sdk/client"
)

// Container represents a running Temporal container for testing.
type Container struct {
    docker.Container
    HostPort string
}

// StartTemporal starts a Temporal dev server container for running tests.
// Uses the temporalio/temporal CLI image with embedded SQLite for test
// isolation (no external database needed). The K8s dev cluster uses
// PostgreSQL via auto-setup, but tests use SQLite for simplicity.
func StartTemporal() (Container, error) {
    const (
        image = "temporalio/temporal:1.3.0"
        name  = "test-temporal"
        port  = "7233"
    )

    // Note: We do NOT specify -p port mappings here.
    // The docker.StartContainer uses -P to assign random host ports,
    // which allows tests to run alongside a dev server using port 7233.
    dockerArgs := []string{}

    // The temporalio/temporal image requires explicit command args
    // to run in dev server mode with all interfaces bound.
    appArgs := []string{"server", "start-dev", "--ip", "0.0.0.0"}

    c, err := docker.StartContainer(image, name, port, dockerArgs, appArgs)
    if err != nil {
        return Container{}, fmt.Errorf("starting temporal container: %w", err)
    }

    // Fix the host address if it's 0.0.0.0
    hostPort := c.HostPort
    if strings.HasPrefix(hostPort, "0.0.0.0:") {
        hostPort = strings.Replace(hostPort, "0.0.0.0:", "localhost:", 1)
    }

    container := Container{
        Container: c,
        HostPort:  hostPort,
    }

    // Wait for Temporal to be ready
    if err := waitForReady(hostPort); err != nil {
        docker.StopContainer(c.Name)
        return Container{}, fmt.Errorf("waiting for temporal to be ready: %w", err)
    }

    return container, nil
}

// StopTemporal stops and removes the Temporal container.
func StopTemporal(c Container) error {
    return docker.StopContainer(c.Name)
}

// DumpLogs outputs logs from the Temporal container.
func DumpLogs(c Container) []byte {
    return docker.DumpContainerLogs(c.Name)
}

// waitForReady waits for Temporal to accept gRPC connections using the
// Temporal Go SDK client.Dial + CheckHealth pattern.
func waitForReady(hostPort string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return fmt.Errorf("timeout waiting for temporal at %s", hostPort)
        case <-ticker.C:
            c, err := client.Dial(client.Options{
                HostPort: hostPort,
            })
            if err != nil {
                continue
            }
            // Verify the server is actually healthy, not just accepting TCP
            _, err = c.CheckHealth(ctx, &client.CheckHealthRequest{})
            c.Close()
            if err == nil {
                return nil
            }
        }
    }
}

// NewTestClient creates a new Temporal client connected to the given host.
// This bypasses the singleton pattern for test isolation.
func NewTestClient(hostPort string) (client.Client, error) {
    return client.Dial(client.Options{
        HostPort: hostPort,
    })
}

var (
    testContainer *Container
    testMu        sync.Mutex
    testStarted   bool
)

// GetTestContainer returns a shared Temporal container for tests.
func GetTestContainer(t *testing.T) Container {
    t.Helper()

    testMu.Lock()
    defer testMu.Unlock()

    if !testStarted {
        const image = "temporalio/temporal:1.3.0"
        const name = "servicetest-temporal"
        const port = "7233"

        // Clean up any existing container
        docker.StopContainer(name)

        dockerArgs := []string{
            // Note: We do NOT specify -p port mappings here.
            // The docker.StartContainer uses -P to assign random host ports,
            // which allows tests to run alongside a dev server using port 7233.
        }

        appArgs := []string{"server", "start-dev", "--ip", "0.0.0.0"}

        c, err := docker.StartContainer(image, name, port, dockerArgs, appArgs)
        if err != nil {
            t.Fatalf("starting temporal container: %s", err)
        }

        // Fix the host address if it's 0.0.0.0
        hostPort := c.HostPort
        if strings.HasPrefix(hostPort, "0.0.0.0:") {
            hostPort = strings.Replace(hostPort, "0.0.0.0:", "localhost:", 1)
        }

        container := Container{
            Container: c,
            HostPort:  hostPort,
        }

        if err := waitForReady(hostPort); err != nil {
            docker.StopContainer(c.Name)
            t.Fatalf("waiting for temporal: %s", err)
        }

        testContainer = &container
        testStarted = true

        t.Logf("Temporal Started: %s at %s", c.Name, hostPort)
    }

    return *testContainer
}
```

The test file should verify:
- `TestStartTemporal` - Container starts and returns valid HostPort
- `TestGetTestContainer` - Singleton returns same container on multiple calls
- `TestNewTestClient` - Client connects to test container and passes CheckHealth
- `TestSimpleWorkflow` - Register and execute a trivial workflow against test container

---

### Task 6: Update Makefile with Workflow-Worker Targets

**Status**: Pending

**Description**: Add all necessary Makefile targets for the workflow-worker service and Temporal infrastructure. This integrates the new services into the existing Ardan Labs development workflow.

**Notes**:
- Add `TEMPORAL` and `TEMPORAL_UI` image variables to dependencies section (line ~121, after RABBITMQ)
- Add `WORKFLOW_WORKER_APP` and `WORKFLOW_WORKER_IMAGE` variables (line ~131, after AUTH_IMAGE)
- Add `workflow-worker` to `.PHONY: build` target (line 174)
- Add `workflow-worker` build target (after `auth` target, line ~199)
- Add Temporal + UI images to `dev-up` (line ~218, alongside RABBITMQ load)
- Add `workflow-worker` image to `dev-load` (line ~244, alongside AUTH_IMAGE load)
- Add Temporal and workflow-worker to `dev-apply` (after database, before auth)
- Add `workflow-worker` to `dev-restart` (line ~268)
- Add `dev-logs-workflow-worker` target (after dev-logs-auth, line ~278)
- Add `dev-describe-workflow-worker` target (after dev-describe-auth, line ~298)
- Add `temporal-ui` target
- Update `dev-docker` to pull Temporal + UI images (line ~168)
- Update `help` target with new commands

**Files**:
- `Makefile`

**Implementation Guide**:

New variables (add to dependencies section, after `RABBITMQ` on line 121):
```makefile
TEMPORAL        := temporalio/auto-setup:1.26.2
TEMPORAL_UI     := temporalio/ui:2.34.0
```

New app/image variables (add after `AUTH_IMAGE` on line 131):
```makefile
WORKFLOW_WORKER_APP    := workflow-worker
WORKFLOW_WORKER_IMAGE  := $(BASE_IMAGE_NAME)/$(WORKFLOW_WORKER_APP):$(VERSION)
```

Update build target (line 174):
```makefile
.PHONY: build ichor metrics auth workflow-worker
build: ichor metrics auth workflow-worker
```

New build target (add after `auth` target around line 199):
```makefile
workflow-worker:
	docker build \
		-f zarf/docker/dockerfile.workflow-worker \
		-t $(WORKFLOW_WORKER_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
		.
```

Update `dev-docker` to pull Temporal + UI (add lines after `docker pull $(RABBITMQ) & \`):
```makefile
	docker pull $(TEMPORAL) & \
	docker pull $(TEMPORAL_UI) & \
```

Update `dev-up` to load Temporal + UI images (add lines after `kind load docker-image $(RABBITMQ)` on line 218):
```makefile
	kind load docker-image $(TEMPORAL) --name $(KIND_CLUSTER) & \
	kind load docker-image $(TEMPORAL_UI) --name $(KIND_CLUSTER) & \
```

Update `dev-load` (add line after AUTH_IMAGE load on line 244):
```makefile
	kind load docker-image $(WORKFLOW_WORKER_IMAGE) --name $(KIND_CLUSTER) & \
```

Update `dev-apply` (add Temporal after database rollout on line 255, add workflow-worker after ichor wait on line 264):
```makefile
	# After database rollout status (line 255) - Temporal needs PostgreSQL ready:
	kustomize build zarf/k8s/dev/temporal | kubectl apply -f -
	kubectl rollout status --namespace=$(NAMESPACE) --watch --timeout=120s deployment/temporal

	# ... existing rabbitmq, auth, and ichor apply ...

	# After ichor wait (line 264):
	kustomize build zarf/k8s/dev/workflow-worker | kubectl apply -f -
	kubectl wait pods --namespace=$(NAMESPACE) --selector app=$(WORKFLOW_WORKER_APP) --timeout=120s --for=condition=Ready
```

**dev-apply ordering** (full sequence for clarity):
```
grafana -> prometheus -> tempo -> loki -> promtail
  -> database (rollout status sts)
  -> temporal (rollout status deployment)     <-- NEW (needs database)
  -> rabbitmq (rollout status deployment)
  -> auth (wait pods Ready)
  -> ichor (wait pods Ready)
  -> workflow-worker (wait pods Ready)        <-- NEW (needs temporal)
```

Update `dev-restart` (line 266-268):
```makefile
dev-restart:
	kubectl rollout restart deployment $(AUTH_APP) --namespace=$(NAMESPACE)
	kubectl rollout restart deployment $(ICHOR_APP) --namespace=$(NAMESPACE)
	kubectl rollout restart deployment $(WORKFLOW_WORKER_APP) --namespace=$(NAMESPACE)
```

New targets (add after `dev-describe-auth` around line 298):
```makefile
dev-logs-workflow-worker:
	kubectl logs --namespace=$(NAMESPACE) -l app=$(WORKFLOW_WORKER_APP) --all-containers=true -f --tail=100

dev-describe-workflow-worker:
	kubectl describe pod --namespace=$(NAMESPACE) -l app=$(WORKFLOW_WORKER_APP)

temporal-ui:
	open http://localhost:8080
```

Update `help` target (around line 660):
```makefile
	@echo "  dev-logs-workflow-worker View workflow-worker logs"
	@echo "  dev-describe-workflow-worker Describe workflow-worker pods"
	@echo "  temporal-ui             Open Temporal Web UI"
```

**Note on `dev-update` / `dev-update-apply`**: These targets compose `build`, `dev-load`, and `dev-restart`/`dev-apply` respectively. Since we're updating `build`, `dev-load`, `dev-restart`, and `dev-apply`, these composite targets automatically include workflow-worker without any additional changes.

---

### Task 7: Verify dev-bounce Works with New Services

**Status**: Pending

**Description**: Run the full `make dev-bounce` cycle to verify Temporal and workflow-worker integrate correctly with the existing cluster. This validates both new services AND ensures no regressions on existing services.

**Notes**:
- Run `make dev-bounce`
- Verify Temporal UI accessible at localhost:8080
- Verify workflow-worker pod reaches Running state
- Verify Temporal auto-setup created databases in PostgreSQL (`temporal`, `temporal_visibility`)
- Verify Temporal "default" namespace exists and is healthy
- Check `make dev-status` shows ALL pods running (ichor, auth, rabbitmq, temporal, workflow-worker, database, grafana, etc.)
- Verify no errors in `make dev-logs-workflow-worker`
- Verify existing services still work: `make dev-logs` (ichor), `make dev-logs-auth` (auth)
- Run `make test-only` to verify no test regressions

**Files**: (none - validation only)

**Implementation Guide**:

```bash
# Full bounce test
make dev-bounce

# Verify ALL pods running (existing + new)
make dev-status
# Expected: database, temporal (2/2 containers), rabbitmq, auth, ichor, workflow-worker all Running

# Check Temporal UI
make temporal-ui
# or: open http://localhost:8080
# Verify: "default" namespace visible, no errors

# Verify Temporal created its databases in PostgreSQL
make pgcli
# Then: \l
# Expected: temporal and temporal_visibility databases exist alongside ichor

# Check workflow-worker logs
make dev-logs-workflow-worker
# Expected: startup log, connected to Temporal

# Verify existing services still work
make dev-logs        # ichor service
make dev-logs-auth   # auth service

# Verify no test regressions from go.mod changes
make test-only

# If hostNetwork is true, Temporal gRPC is directly accessible:
# No port-forward needed - worker connects via K8s DNS (temporal-service:7233)
```

---

### Task 8: Verify Test Container Infrastructure Works

**Status**: Pending

**Description**: Run the `foundation/temporal/` tests to verify the test container starts correctly and can execute workflows.

**Notes**:
- Run `go test ./foundation/temporal/...`
- Verify container starts and is reachable
- Verify tests can register and execute a simple workflow against test container
- Verify singleton pattern works (multiple calls to GetTestContainer return same container)
- Verify test container uses random port (doesn't conflict with dev cluster)
- Note: Test container uses SQLite dev server, NOT PostgreSQL (by design - see Critical Design Decisions)

**Files**: (none - validation only)

**Implementation Guide**:

```bash
# Run foundation temporal tests
go test -v ./foundation/temporal/...

# Verify container is running with random port (not 7233)
docker ps | grep servicetest-temporal

# Verify no conflict with dev cluster Temporal
# (test container uses random port via -P flag, dev uses 7233)
```

---

## Validation Criteria

- [ ] `make dev-bounce` completes successfully (all existing + new services start)
- [ ] Temporal UI accessible at http://localhost:8080
- [ ] workflow-worker pod reaches Running state (`kubectl get pods -n ichor-system`)
- [ ] Temporal pod shows 2/2 containers ready (server + UI)
- [ ] Temporal "default" namespace exists and is healthy (visible in Temporal UI)
- [ ] `temporal` and `temporal_visibility` databases exist in PostgreSQL (`make pgcli` -> `\l`)
- [ ] `go test ./foundation/temporal/...` passes
- [ ] Test container can start, accept connections (CheckHealth), and execute a simple workflow
- [ ] All existing tests still pass (`make test-only` - ensure no regressions from go.mod changes)
- [ ] `make dev-docker` pulls the Temporal + UI images
- [ ] `make build` builds the workflow-worker container
- [ ] `make dev-update` and `make dev-update-apply` include workflow-worker

---

## Deliverables

- `zarf/k8s/dev/temporal/kustomization.yaml` - Temporal K8s kustomization
- `zarf/k8s/dev/temporal/dev-temporal.yaml` - Temporal deployment (server + UI sidecar) + service
- `zarf/k8s/base/workflow-worker/base-workflow-worker.yaml` - Worker base deployment
- `zarf/k8s/base/workflow-worker/kustomization.yaml` - Worker base kustomization
- `zarf/k8s/dev/workflow-worker/kustomization.yaml` - Worker dev overlay
- `zarf/k8s/dev/workflow-worker/dev-workflow-worker-configmap.yaml` - Worker dev config
- `zarf/k8s/dev/workflow-worker/dev-workflow-worker-patch-deploy.yaml` - Worker dev patches
- `zarf/docker/dockerfile.workflow-worker` - Worker Dockerfile
- `foundation/temporal/temporal.go` - Test container infrastructure
- `foundation/temporal/temporal_test.go` - Test container tests
- Updated `Makefile` with all new targets

---

## Gotchas & Tips

### Common Pitfalls

- **`DB=postgres12` NOT `postgresql`**: The valid database driver values for `temporalio/auto-setup` are `postgres12` and `postgres12_pgx`. The value `postgresql` or `postgres` will silently fail or error. This was changed in Temporal 1.24.0+.
- **Auto-setup schema migration on first boot**: The first time Temporal starts, it creates `temporal` and `temporal_visibility` databases and runs schema migrations. This takes 30-60s. Subsequent restarts skip this. Set `initialDelaySeconds` high enough (60s for liveness).
- **Temporal UI is a separate image**: `temporalio/auto-setup` does NOT include the Web UI. Deploy `temporalio/ui` as a sidecar or separate pod. The UI connects to the server via `TEMPORAL_ADDRESS=localhost:7233` (same pod) or `TEMPORAL_ADDRESS=temporal-service:7233` (different pod).
- **UI port is 8080, not 8233**: The `temporalio/ui` image defaults to port 8080. Only the `temporalio/temporal` dev server uses 8233.
- **dev-apply ordering matters**: Temporal needs PostgreSQL ready before it can run auto-setup. Place Temporal after `database` rollout status in dev-apply. Workflow-worker needs Temporal ready. Place it after `temporal` rollout status.
- **Port conflicts**: Temporal gRPC uses port 7233, UI uses port 8080. With `hostNetwork: true` in dev, these are directly accessible on localhost. Make sure nothing else uses those ports.
- **Test container uses different image**: K8s uses `temporalio/auto-setup` (PostgreSQL), test containers use `temporalio/temporal` (SQLite dev server). This is intentional - test containers must be self-contained.
- **Host address fix in test containers**: After `docker.StartContainer`, the `HostPort` may start with `0.0.0.0:`. Must replace with `localhost:` for compatibility (same fix as rabbitmq.go line 179).
- **Container cleanup before start**: Always call `docker.StopContainer(name)` before `docker.StartContainer` in `GetTestContainer` to clean up orphaned containers (same pattern as rabbitmq.go line 162).
- **`t.Helper()` in GetTestContainer**: Always call `t.Helper()` at the top so test failure stack traces point to the calling test (same pattern as rabbitmq.go line 151).
- **gRPC vs HTTP for health checks**: Workers connect to Temporal via gRPC (port 7233). The `waitForReady` function should use `client.Dial` + `client.CheckHealth` (Temporal SDK). For K8s probes, use `tcpSocket` on the gRPC port. For the UI container, use `httpGet` on port 8080.
- **Go module dependencies**: Adding the Temporal SDK (`go.temporal.io/sdk`) will add significant dependencies to `go.mod`. Run `go mod tidy` after adding the import. If the project vendors: `go mod vendor` too.
- **Two Docker images to pull**: `dev-docker` and `dev-up` must pull/load both `TEMPORAL` (auto-setup) and `TEMPORAL_UI` images.

### Tips

- Look at `foundation/rabbitmq/rabbitmq.go` as the **definitive pattern** for test containers - match it line-for-line
- Look at `zarf/k8s/dev/rabbitmq/dev-rabbitmq.yaml` for the K8s deployment pattern
- Look at `zarf/k8s/base/ichor/` and `zarf/k8s/dev/ichor/` for the full base/dev overlay pattern
- Look at `zarf/docker/dockerfile.ichor` for the Dockerfile pattern
- The `foundation/docker/docker.go` package handles all Docker container lifecycle - use it (has built-in retry logic)
- The workflow-worker service entry point (`api/cmd/services/workflow-worker/main.go`) will be implemented in Phase 9 - Phase 1 only needs the K8s/Docker infrastructure. The initial main.go can be a simple placeholder that logs startup and sleeps.
- `docker.StartContainer` has built-in retry logic (2 attempts) and reuses existing containers - no need to handle this yourself
- Verify Temporal created its databases: `make pgcli` then `\l` should show `temporal` and `temporal_visibility`

---

## Testing Strategy

### Unit Tests

- `foundation/temporal/temporal_test.go`:
  - `TestStartTemporal` - Container starts and returns valid HostPort
  - `TestGetTestContainer` - Singleton returns same container on multiple calls
  - `TestNewTestClient` - Client connects to test container and passes `CheckHealth`
  - `TestSimpleWorkflow` - Register and execute a trivial workflow against test container

### Integration Tests

- Manual verification via `make dev-bounce`:
  - All pods reach Running state (including existing services - no regressions)
  - Temporal pod shows 2/2 containers (server + UI)
  - Temporal UI is accessible at http://localhost:8080
  - Temporal "default" namespace exists
  - `temporal` and `temporal_visibility` databases exist in PostgreSQL
  - No error logs in workflow-worker
  - `make test-only` passes (no regressions from go.mod changes)

---

## Existing Patterns Reference

### RabbitMQ K8s Pattern (`zarf/k8s/dev/rabbitmq/`)
- Single YAML with Namespace + Deployment + Service
- Kustomization just references the single YAML
- Deployment uses liveness/readiness probes (`rabbitmq-diagnostics check_port_connectivity`)
- Service uses ClusterIP type with `{app}-service` naming
- Resource limits: 100m request, 1000m limit CPU

### Ichor Service Pattern (`zarf/k8s/base/ichor/` + `zarf/k8s/dev/ichor/`)
- Base: Namespace + Deployment (with init container) + Service
- Dev overlay: Kustomization references base + configmap + patches
- Dev patches: replicas=1, strategy=Recreate, hostNetwork=true, dnsPolicy=ClusterFirstWithHostNet, resource limits (250m CPU, 36Mi memory)
- Image override via kustomize `images` field

### Test Container Pattern (`foundation/rabbitmq/rabbitmq.go`)
- `Start*()` function creates container via `foundation/docker`
- `GetTestContainer(t)` singleton with `sync.Mutex`, `t.Helper()`, cleanup before start
- Host address fix: `0.0.0.0:` -> `localhost:`
- `waitForReady()` polls with 500ms ticker until service accepts connections (30s timeout for RabbitMQ)
- Container struct embeds `docker.Container`, adds service-specific fields (URL/HostPort)
- `NewTestClient()` for isolated test clients
- Logging: `t.Logf` with container name and host port

### Dockerfile Pattern (`zarf/docker/dockerfile.ichor`)
- Multi-stage: `golang:1.23` build -> `alpine:3.20` runtime
- CGO_ENABLED=0
- Non-root user (ichor:ichor, uid/gid 1000)
- OCI labels with BUILD_DATE and BUILD_REF

---

## Commands Reference

```bash
# Start this phase
/workflow-temporal-next

# Validate this phase
/workflow-temporal-validate 1

# Review plan before implementing
/workflow-temporal-plan-review 1

# Review code after implementing
/workflow-temporal-review 1
```
