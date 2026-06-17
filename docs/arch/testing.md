# testing

[bus]=business [app]=application [api]=HTTP [db]=store [sdk]=shared [dbtest]=test-infra [apitest]=integration-test-infra
→=depends on ⊕=writes ⊗=reads ⚡=external [tx]=transaction [cache]=cached

---

## ContainerLifecycle [dbtest]

Three shared test containers. Each uses a distinct isolation strategy:

| Container | Name | Image | Isolation strategy |
|-----------|------|-------|-------------------|
| Postgres | `servicetest` | postgres:16.4 | shared container, per-test PID-tagged DB + orphan reaper |
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
  - container persists across test runs — the container itself is never stopped automatically, but leftover test *databases* are reclaimed each run (see the orphan reaper under `NewDatabase`)
  - `make test-down` explicitly removes the container when needed

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
0. orphanSweepOnce → reapOrphanedDatabases    → once per process: drop orphan DBs (dead owner), capped
1. docker.StartContainer("servicetest")       → reuse or create container
2. CREATE DATABASE ichortest_<pid>_<12 rand>  → e.g. "ichortest_50647_qz73wr3zjrbd" (retry on collision)
3. t.Cleanup(DROP DATABASE <name>)            → registered immediately, before migrate/seed (closes leak window)
4. open pool, SET TIME ZONE 'America/New_York'
5. migrate (run all migrations)
6. seed   (full seed chain via InsertSeedData)
```

key facts:
  - DB names embed the creating PID + 12 random chars (36^12 ≈ 4.7e18) — collisions are effectively impossible, even against a backlog of orphans; a retry-on-`already exists` loop is a deterministic backstop
  - ⚠ names are NO LONGER 4 random letters. The old 26^4 = 456K space + a shared container with no DB cleanup let orphans from killed/interrupted runs accumulate (observed: **717**) and collide → `database "xxxx" already exists`. This was the real cause of full-suite (`./...`) flakiness, not port exhaustion.
  - orphan reaper (`reapOrphanedDatabases`): runs once per process (`sync.Once`) BEFORE the per-test 10s timeout starts; drops test DBs whose owning PID is dead (`syscall.Kill(pid,0)`), handling both new-format and legacy 4-letter names. Container self-heals to 0 orphans each run.
  - ⚠ reaper is capped at `maxOrphanReap` (32) per run: `DROP DATABASE` forces a Postgres checkpoint, so dropping a large backlog at once triggers an I/O storm (observed: ~700 drops → a 261s checkpoint syncing 5334 files) that stalls concurrent migrations. The cap bounds the storm; any remainder is reclaimed over subsequent runs.
  - DROP cleanup is registered immediately after CREATE (not after migrate/seed), so a `t.Fatalf` mid-setup cannot orphan the DB; a drop hiccup logs instead of failing the test (reaper reclaims it next run)
  - each test gets its own database — fully isolated even within parallel runs
  - cleanup drops only the test's own DB; the `servicetest` container is never stopped by test code
  - `BusDomain` is fully wired: every bus package is instantiated against the test DB
  - `BusDomain.OutboxWriter` (*outbox.Writer) is built once in `newBusDomains` and injected via `.WithOutbox(...)` into the ~67 cascade buses (+ `.WithOutboxEmitter` on workflowBus), so the suite exercises the PRODUCTION outbox+relay cascade path (F8 parity), not the retired delegate path

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
                  → temporal.NewRelay(db.DB, workflowTrigger, RelayConfig{}) + go relay.Run(relayCtx)  ← cascade dispatcher (mirrors all.go)
                  → t.Cleanup: cancelRelay() + w.Stop() + tc.Close()
```

key facts:
  - task queue is unique per test (`t.Name()` includes subtest path) — prevents cross-test activity routing
  - `alertBus` and `approvalRequestBus` are constructed fresh per `InitWorkflowInfra` call, NOT from `db.BusDomain` — avoids accumulated state across tests that share a DB
  - NO `DelegateHandler` is wired — that type is DELETED (F2 / F7.1). The relay is the SOLE cascade dispatcher: a test just writes through `db.BusDomain` buses (which persist outbox rows in-tx) and the relay drains them into `WorkflowTrigger.OnEntityEvent`. Tests that previously called `wf.DelegateHandler.RegisterDomain(...)` no longer need any manual `Delegate.Register` for cascade triggering

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

> **Update 2026-06-09 — port saturation NOT reproduced; orphan DBs were the real flake.**
> The `floor/scenarios` walk (22 subtests) was re-run at the default `-parallel`
> (GOMAXPROCS, ~8–10 on macOS Apple Silicon) and passed clean in ~26s with **zero**
> `can't assign requested address` errors. The failures originally attributed to the
> port ceiling below were actually **DB-name collisions** from ~717 accumulated orphan
> databases — now fixed by the PID-tagged names + reaper in `NewDatabase`. Treat the
> `-parallel 2` ceiling below as **unverified on current hardware**: do NOT cap a fan-out
> harness on its account without first reproducing port exhaustion. The genuinely-true,
> still-current constraint is the reaper's `maxOrphanReap` cap (`DROP DATABASE` forces a
> checkpoint). The section below is retained for history and the CI-portability caveat.

Tests that fan out many subtests within a single package (e.g. `api/cmd/services/ichor/tests/floor/scenarios/` which walks 22 scenarios via `t.Run` + `t.Parallel`) were historically thought to need a `-parallel N` cap. Each subtest:

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

## ⚠ Scenario family registration (floor harness)

The floor scenario harness (`api/cmd/services/ichor/tests/floor/scenarios/`) discovers scenarios by enumerating `deployments/scenarios/` directories, then resolves each to a **family** (which selects the walk/handler) via a precedence chain: `deriveFamily(name)` → `familyOverrides` (`harness_test.go`) → `customRowOverrides`. Dropping a scenario folder on disk is **necessary but not sufficient** — a scenario whose name prefix matches no `deriveFamily` case must also be registered, or it has no family and dispatch fatals.

⚠ When adding a **lever-only** scenario (just `scenario.yaml`, no fixture subdirs, e.g. `e2e-pick-tote`), also add `"<name>": familyPick` to `familyOverrides` in `harness_test.go`, mirroring `e2e-pick-strict`. Otherwise `deriveFamily` returns `""` and dispatch fatals at runtime: `scenario "<name>" has empty family and no Custom handler`. The fatal message names `customRowOverrides`/`scenarios_test.go`, but lever-only pick scenarios belong in `familyOverrides`/`harness_test.go`.

---

## ⚠ Running tests

**Never run `go test ./...`** — hundreds of tests, many require live DB.
Always scope to packages you changed:

```bash
go test ./business/domain/inventory/inventorylocationbus/...
go test ./api/cmd/services/ichor/tests/inventory/putawaytaskapi/...
```

`make test-down` shuts down all test containers when done.
