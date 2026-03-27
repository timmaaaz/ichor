# Progress Summary: testing.md

## Overview
Comprehensive test infrastructure design. Three shared test containers with distinct isolation strategies. Emphasis on parallel safety and per-test isolation.

## ContainerLifecycle [dbtest]

### Postgres Container (Primary Pattern)

**File:** `business/sdk/dbtest/dbtest.go`, `foundation/docker/docker.go`

```
NewDatabase()
  → docker.StartContainer("servicetest")
    → exists() checks if container running
      → yes: return existing HostPort
      → no: docker run -P (random host port)
```

#### Key Facts
- **exists() mechanism** — calls `docker inspect`; returns running container if found
- **startContainer()** — checks exists() first; only creates if not running
- **Retry loop** — 2 retries with 100ms/200ms sleep handles race condition where two processes try to create simultaneously
- **Container persistence** — survives across test runs
- **Cleanup mechanism** — `make test-down` explicitly removes when needed

#### Per-Test Isolation
```
1. docker.StartContainer("servicetest") → reuse or create container
2. CREATE DATABASE <random 4-char name> → e.g. "vjsb" (26^4 = 456K possibilities)
3. SET TIME ZONE 'America/New_York'
4. migrate (run all migrations)
5. seed (full seed chain via InsertSeedData)
6. t.Cleanup → DROP DATABASE <name> (never stops the container)
```

### Temporal Container (Process-Local Singleton)

**File:** `foundation/temporal/temporal.go`, function: `GetTestContainer`

```go
var (
    testContainer *Container
    testMu        sync.Mutex
    testStarted   bool
)
```

#### Key Facts
- **Mutex + flag guard** — within a single process, only the first caller creates the container
- **All subsequent callers** return the cached pointer
- **docker.StopContainer(name)** runs once per process inside `!testStarted` guard (intentional)
- **Why fresh container per process** — Temporal dev server uses SQLite in-memory; requires isolation between test processes
- **Per-test unique task queue** — "test-workflow-" + t.Name() prevents cross-test activity routing

#### WorkflowTestEntry Pattern
```
InitWorkflowInfra(t, db)
  → GetTestContainer(t)
  → temporalclient.Dial
  → workflow.NewBusiness
  → worker.New(tc, "test-workflow-"+t.Name())
  → w.Start()
  → TriggerProcessor.Initialize
  → NewWorkflowTrigger.WithTaskQueue(taskQueue)
  → t.Cleanup: w.Stop() + tc.Close()
```

## NewDatabase [dbtest] — `business/sdk/dbtest/dbtest.go`

**Responsibility:** Per-test database setup and teardown.

### Function Signature
```go
func NewDatabase(t *testing.T, testName string) *Database
```

### Lifecycle Per Test
1. **docker.StartContainer("servicetest")** — reuse or create container
2. **CREATE DATABASE <random 4-char name>** — e.g. "vjsb" (26^4 = 456K possibilities)
3. **SET TIME ZONE 'America/New_York'**
4. **migrate** — run all migrations
5. **seed** — full seed chain via InsertSeedData
6. **t.Cleanup → DROP DATABASE <name>** — never stops the container

### Key Facts
- **Each test gets its own database** — fully isolated even within parallel runs
- **Cleanup drops only test's own DB** — the `servicetest` container is never stopped by test code
- **BusDomain is fully wired** — every bus package is instantiated against the test DB

## IntegrationTestEntry [apitest] — `api/sdk/http/apitest/start.go`

**Responsibility:** Full HTTP server setup for integration tests.

### Function Signature
```go
func StartTest(t, testName) *Test
```

### Setup Flow
```
StartTest
  → NewDatabase
  → auth.New
  → httptest.NewServer(authbuild.Routes())
  → httptest.NewServer(ichorbuild.Routes()) [main mux]
```

### Key Facts
- **Each integration test** gets an isolated DB + full HTTP server stack
- **httptest.NewServer** — no network port allocation; server listens on loopback
- **Test struct** wraps `*dbtest.Database` + `*auth.Auth` + `http.Handler` for table-driven tests

## WorkflowTestEntry [apitest] — `api/sdk/http/apitest/workflow.go`

**Responsibility:** Temporal infrastructure setup for workflow integration tests.

### Function Signature
```go
func InitWorkflowInfra(t, db) *WorkflowInfra
```

### WorkflowInfra Struct
Contains:
- WorkflowBus
- TemporalClient
- WorkflowTrigger
- DelegateHandler
- TriggerProcessor
- Worker

### Key Facts
- **Task queue unique per test** — `"test-workflow-" + t.Name()` prevents cross-test routing
- **alertBus and approvalRequestBus** — constructed fresh per call, NOT from `db.BusDomain` (avoids accumulated state)
- **DelegateHandler wired but not registered** — test must call `db.BusDomain.Delegate.Register(...)` explicitly if event-driven triggering needed
- **Cleanup** — `w.Stop() + tc.Close()` via t.Cleanup

## Critical Warnings

### ⚠ NEVER add `docker rm -f servicetest` before `docker.StartContainer`

**Why:** foundation/docker/docker.go:startContainer() already checks exists() before attempting creation. Force-removing defeats this and causes connection failures in parallel test processes.

**Bug removed 2026-03-13** — do not re-add.

#### Wrong Pattern (causes parallel test failures):
```go
exec.Command("docker", "rm", "-f", "servicetest").Run()   // ← destroys other processes' container
c, err := docker.StartContainer(image, name, port, dockerArgs, appArgs)
```

#### Correct Pattern:
```go
c, err := docker.StartContainer(image, name, port, dockerArgs, appArgs)
```

#### Cleanup Mechanism:
`make test-down` is the correct mechanism for explicit container cleanup.

## Adding a New Test Container

If a new shared container is needed (e.g., Redis, RabbitMQ):

1. **Add a foundation package:** `foundation/{service}/{service}.go`
2. **Implement pattern:** StartContainer → exists() — do NOT force-remove before starting
3. **If process-level isolation required** (stateful server): add mutex+flag singleton like GetTestContainer
4. **If cross-process reuse safe** (stateless): rely purely on exists() guard in docker.StartContainer

## Parallel Test Safety Rules

- **Business layer tests** — safe to run in parallel across packages; each gets its own DB
- **Workflow integration tests** — safe within a process (unique task queues); new Temporal container per process
- **Never call** `docker.StopContainer("servicetest")` from test code — terminates other concurrent processes
- **t.Cleanup** must drop test DB only, never stop container

## Running Tests

### Never run `go test ./...`
- Hundreds of tests
- Many require live DB
- Scope to packages you changed

### Correct Approach:
```bash
go test ./business/domain/inventory/inventorylocationbus/...
go test ./api/cmd/services/ichor/tests/inventory/putawaytaskapi/...
```

### Cleanup:
```bash
make test-down
```

Shuts down all test containers when done.

## Critical Points
- **Per-test DB isolation** — each test gets fresh, isolated database
- **Process-level Temporal isolation** — unique task queues prevent cross-test routing
- **No force-remove pattern** — relies on exists() guard for concurrency safety
- **Cleanup is automatic** — t.Cleanup handles per-test cleanup; container persists
- **Parallel-safe by design** — multiple test processes won't conflict

## Notes for Future Development
The test infrastructure is well-designed for:
- **Parallel execution** — different test processes use different DBs
- **Isolated container lifecycle** — containers are created once and reused
- **Per-process Temporal isolation** — avoids Temporal state leakage

Most changes should be:
- Scoping test runs to changed packages (low-risk)
- Using make test-down for cleanup (low-risk)
- Adding new containers only if absolutely needed (moderate, requires pattern understanding)
