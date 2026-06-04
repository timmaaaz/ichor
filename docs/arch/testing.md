# testing

[bus]=business [app]=application [api]=HTTP [db]=store [sdk]=shared [dbtest]=test-infra [apitest]=integration-test-infra
→=depends on ⊕=writes ⊗=reads ⚡=external [tx]=transaction [cache]=cached

---

## ContainerLifecycle [dbtest]

Three shared test containers. Each uses a distinct isolation strategy:

| Container | Name | Image | Isolation strategy |
|-----------|------|-------|-------------------|
| Postgres | `servicetest` | postgres:16.4 | shared container, per-test random DB |
| Temporal | `servicetest-temporal` | temporalio/temporal:latest | process-local singleton (mutex+flag) |

### Postgres (the primary pattern)

file: business/sdk/dbtest/dbtest.go
file: foundation/docker/docker.go

```
NewDatabase() → docker.StartContainer("servicetest") → startContainer() → exists()
                                                                          ↓ container running?
                                                              yes → return existing HostPort
                                                               no → docker run -P (random host port)
```

key facts:
  - `exists()` calls `docker inspect` — returns running container if found, error otherwise
  - `startContainer()` checks `exists()` first; only creates if not running
  - retry loop in `StartContainer` (2 retries, 100ms/200ms sleep) handles the race where two processes try to create simultaneously
  - container persists across test runs — no cleanup happens automatically
  - `make test-down` explicitly removes it when needed

### Temporal (process-local singleton)

file: foundation/temporal/temporal.go
function: `GetTestContainer`

```go
var (
    testContainer *Container
    testMu        sync.Mutex
    testStarted   bool
)
```

key facts:
  - mutex + `testStarted` flag: within a single process, only the first caller creates the container; all others return the cached pointer
  - `docker.StopContainer(name)` runs once per process (inside `!testStarted` guard) — this is intentional: Temporal dev server uses SQLite in-memory, so a fresh container per test process is required for isolation
  - each test connects with its own `temporalclient.Client`; worker uses unique task queue per test (`"test-workflow-" + t.Name()`)

---

## NewDatabase [dbtest]

file: business/sdk/dbtest/dbtest.go

```go
func NewDatabase(t *testing.T, testName string) *Database
```

lifecycle per test:
```
1. docker.StartContainer("servicetest")   → reuse or create container
2. CREATE DATABASE <random 4-char name>   → e.g. "vjsb" (26^4 = 456K possibilities)
3. SET TIME ZONE 'America/New_York'
4. migrate (run all migrations)
5. seed   (full seed chain via InsertSeedData)
6. t.Cleanup → DROP DATABASE <name>       (never stops the container)
```

key facts:
  - each test gets its own database — fully isolated even within parallel runs
  - cleanup drops only the test's own DB; the `servicetest` container is never stopped by test code
  - `BusDomain` is fully wired: every bus package is instantiated against the test DB

---

## IntegrationTestEntry [apitest]

file: api/sdk/http/apitest/start.go
function: `StartTest(t, testName) *Test`

```
StartTest → NewDatabase → auth.New
                        → httptest.NewServer(authbuild.Routes())   [auth subserver only]
                        → mux.WebAPI(ichorbuild.Routes())          [ichor mux — raw http.Handler]
                        → apitest.New(db, auth, ichorMux)
```

key facts:
  - each integration test gets an isolated DB; the auth subserver runs in `httptest.NewServer`
  - the ichor mux is **not** wrapped in `httptest.NewServer` — tests drive it via `test.ServeHTTP(w, r)` (see `apitest.go:40`)
  - no inbound port allocation for the ichor mux — but each subtest still burns ephemeral source ports on outbound DB connections (see "Parallel ceiling for fan-out integration harnesses" below)
  - `Test` wraps `*dbtest.Database` + `*auth.Auth` + `http.Handler` for use in table-driven tests
  - ⚠ `StartTest` does NOT call `t.Cleanup(server.Close)` on the auth subserver — that subserver leaks per test. Harnesses that fan out many subtests should add the cleanup explicitly (see `api/cmd/services/ichor/tests/floor/scenarios/harness_test.go`)
  - `StartTest` defaults `ScenariosEnabled: false` — integration tests using it bypass `sqldb.ApplyScenarioFilter`. Harnesses that need scenario filtering must construct the mux directly with `ScenariosEnabled: true`

---

## WorkflowTestEntry [apitest]

file: api/sdk/http/apitest/workflow.go
function: `InitWorkflowInfra(t, db) *WorkflowInfra`

```
InitWorkflowInfra → GetTestContainer(t)        → shared Temporal container (singleton per process)
                  → temporalclient.Dial
                  → workflow.NewBusiness
                  → worker.New(tc, "test-workflow-"+t.Name())
                  → w.Start()
                  → TriggerProcessor.Initialize
                  → NewWorkflowTrigger.WithTaskQueue(taskQueue)
                  → t.Cleanup: w.Stop() + tc.Close()
```

key facts:
  - task queue is unique per test (`t.Name()` includes subtest path) — prevents cross-test activity routing
  - `alertBus` and `approvalRequestBus` are constructed fresh per `InitWorkflowInfra` call, NOT from `db.BusDomain` — avoids accumulated state across tests that share a DB
  - `DelegateHandler` is wired but NOT registered into `db.BusDomain.Delegate` automatically — test must call `db.BusDomain.Delegate.Register(...)` explicitly if event-driven triggering is needed

---

## ⚠ NEVER add `docker rm -f servicetest` before `docker.StartContainer`

`foundation/docker/docker.go:startContainer()` already checks `exists()` before attempting creation. Force-removing the container defeats this check and causes connection failures in parallel test processes.

The bug was present in `dbtest.go` and was removed 2026-03-13. Do not re-add it.

correct:
```go
c, err := docker.StartContainer(image, name, port, dockerArgs, appArgs)
```

wrong (causes parallel test failures):
```go
exec.Command("docker", "rm", "-f", name).Run()   // ← destroys other processes' container
c, err := docker.StartContainer(image, name, port, dockerArgs, appArgs)
```

`make test-down` is the correct mechanism for explicit container cleanup.

---

## ⚠ Adding a new test container

If a new shared container is needed (e.g., Redis, RabbitMQ):

1. Add a foundation package: `foundation/{service}/{service}.go`
2. Implement the same `StartContainer` → `exists()` pattern — do NOT force-remove before starting
3. If process-level isolation is required (stateful server), add a mutex+flag singleton like `GetTestContainer` in `foundation/temporal/temporal.go`
4. If cross-process reuse is safe (stateless), rely purely on the `exists()` guard in `docker.StartContainer`

---

## ⚠ Parallel test safety rules

  - business layer tests: safe to run in parallel across packages — each gets its own DB in the shared `servicetest` container
  - workflow integration tests: safe within a process (unique task queues); new Temporal container per process via `GetTestContainer` singleton
  - never call `docker.StopContainer("servicetest")` from test code — it terminates other concurrent test processes
  - `t.Cleanup` must drop the test DB only, never stop the container: see `NewDatabase` cleanup in dbtest.go

---

## ⚠ Parallel ceiling for fan-out integration harnesses

Tests that fan out many subtests within a single package (e.g. `api/cmd/services/ichor/tests/floor/scenarios/` which walks 22 scenarios via `t.Run` + `t.Parallel`) must cap parallelism via `-parallel N`. Each subtest:

  - opens pooled `*sqlx.DB` connections to the shared `servicetest` container (admin DB + test DB), each burning an ephemeral source port
  - issues many HTTP requests per walk to the in-process `httptest`/handler; the auth `httptest.NewServer` listens on loopback, but the Go test process still consumes ephemeral source ports for the outbound DB connections issued from inside handlers

On macOS the ephemeral range is 49152–65535 (~16k ports, TIME_WAIT held ~2min). At 22 scenarios × `-parallel 4` the floor scenario harness exhausts this range and fails with:

```
dial tcp 127.0.0.1:5432: connect: can't assign requested address
```

Safe ceilings observed (macOS Apple Silicon, postgres:16.4):

| Harness shape | Safe -parallel | Notes |
|---|---|---|
| 22-scenario floor walk | 2 | `-parallel 4` fails; `-race` adds ~5.7× wall-clock |
| Single-domain api tests (e.g. labels) | unbounded | small subtest count, no port pressure |

Rule of thumb: if a single package contains >10 parallel subtests **and** each subtest hits the HTTP mux >5 times, cap with `-parallel 2` on macOS. Linux CI with a larger ephemeral range (e.g. 32768–60999) may tolerate `-parallel 4`; verify on the actual CI host before raising.

The author of a fan-out harness should document the chosen ceiling in a comment on the table-driven entry test, with a one-line rationale (e.g. `scenarios_test.go:TestFloorScenarios`).

---

## ⚠ Running tests

**Never run `go test ./...`** — hundreds of tests, many require live DB.
Always scope to packages you changed:

```bash
go test ./business/domain/inventory/inventorylocationbus/...
go test ./api/cmd/services/ichor/tests/inventory/putawaytaskapi/...
```

`make test-down` shuts down all test containers when done.
