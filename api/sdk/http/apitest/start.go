package apitest

import (
	"net/http/httptest"
	"testing"

	authbuild "github.com/timmaaaz/ichor/api/cmd/services/auth/build/all"
	ichorbuild "github.com/timmaaaz/ichor/api/cmd/services/ichor/build/all"
	"github.com/timmaaaz/ichor/api/sdk/http/mux"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/approval"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/control"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/integration"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
	foundationtemporal "github.com/timmaaaz/ichor/foundation/temporal"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// StartTest initialized the system to run a test.
func StartTest(t *testing.T, testName string) *Test {
	db := dbtest.NewDatabase(t, testName)

	// -------------------------------------------------------------------------

	auth, err := auth.New(auth.Config{
		Log:       db.Log,
		DB:        db.DB,
		KeyLookup: &KeyStore{},
	})
	if err != nil {
		t.Fatal(err)
	}

	// -------------------------------------------------------------------------

	server := httptest.NewServer(mux.WebAPI(mux.Config{
		Log:  db.Log,
		Auth: auth,
		DB:   db.DB,
	}, authbuild.Routes()))

	authClient := authclient.New(db.Log, server.URL)

	// -------------------------------------------------------------------------

	mux := mux.WebAPI(mux.Config{
		Log:        db.Log,
		AuthClient: authClient,
		DB:         db.DB,
	}, ichorbuild.Routes())

	return New(db, auth, mux)
}

// StartTestWithTemporal initializes the system like StartTest but stands up a
// real Temporal client (and a worker polling the production task queue) and
// passes the client into the ichor mux. all.go's `if cfg.TemporalClient != nil`
// block then builds the WorkflowTrigger internally and wires it into
// executionapi.Config.Trigger, exercising the exact production composition-root
// path. Use this for HTTP integration tests of routes that depend on a live
// Temporal-backed trigger (e.g. POST /workflow/executions/{id}/rerun), where the
// nil-trigger path returned by StartTest would only surface an Internal error.
//
// Mirrors StartWSTestWithRabbitMQ's pattern of feeding an infra client into
// mux.Config so all.go does the wiring (rather than injecting an externally
// built trigger, which all.go has no hook for).
func StartTestWithTemporal(t *testing.T, testName string) *Test {
	t.Helper()

	// The 4 core sync handlers cover the seeded rules' actions for the rerun-200
	// dispatch assertion; async/inventory handlers are unnecessary because that
	// test only proves a fresh id is returned (it does not run the graph to
	// completion). For an e2e test that actually runs the granular inventory
	// pipeline, use StartTestWithTemporalGranular.
	return startTestWithTemporal(t, testName, registerCoreWorkerActions)
}

// StartTestWithTemporalGranular is StartTestWithTemporal but its production-queue
// worker registers the FULL granular inventory handler set (check_inventory,
// reserve_inventory, check_reorder_point, create_alert, ... + async seek_approval)
// in addition to the core handlers. This is what lets a dispatched (or re-run)
// Rule-5 "Granular Inventory Pipeline" workflow actually execute end-to-end —
// reserve_inventory must run against live stock for the over-order recovery e2e
// test (rerun_e2e_test.go), and the plain core worker has no reserve handler.
//
// It shares the exact composition-root path with StartTestWithTemporal (same
// real TemporalClient fed into mux.Config so all.go builds the production trigger
// + relay); only the worker's registered handler set differs. Existing callers of
// StartTestWithTemporal are untouched.
func StartTestWithTemporalGranular(t *testing.T, testName string) *Test {
	t.Helper()
	return startTestWithTemporal(t, testName, registerGranularWorkerActions)
}

// startTestWithTemporal is the shared body of the Temporal-backed integration
// entrypoints. registerActions installs the worker's handler set on the supplied
// registries (sync + async) and returns the execution-lifecycle store the worker's
// Activities should advance (nil = the MarkExecution* activities no-op, leaving the
// execution record at its created status). Everything else — DB, auth subserver,
// production-queue worker, real TemporalClient fed into the ichor mux — is identical
// across callers.
func startTestWithTemporal(
	t *testing.T,
	testName string,
	registerActions func(db *dbtest.Database, reg *workflow.ActionRegistry, asyncReg *temporal.AsyncRegistry) temporal.ExecutionLifecycleStore,
) *Test {
	t.Helper()

	db := dbtest.NewDatabase(t, testName)

	// -------------------------------------------------------------------------

	ath, err := auth.New(auth.Config{
		Log:       db.Log,
		DB:        db.DB,
		KeyLookup: &KeyStore{},
	})
	if err != nil {
		t.Fatal(err)
	}

	authServer := httptest.NewServer(mux.WebAPI(mux.Config{
		Log:  db.Log,
		Auth: ath,
		DB:   db.DB,
	}, authbuild.Routes()))

	authClient := authclient.New(db.Log, authServer.URL)

	// -------------------------------------------------------------------------
	// Stand up the Temporal test container + client. The client is what makes
	// all.go's trigger real: RerunExecution -> ExecuteWorkflow dispatches against
	// this server and returns a fresh execution id.

	container := foundationtemporal.GetTestContainer(t)
	tc, err := temporalclient.Dial(temporalclient.Options{
		HostPort: container.HostPort,
	})
	if err != nil {
		authServer.Close()
		t.Fatalf("connecting to temporal: %s", err)
	}

	// Register a worker on the PRODUCTION task queue (temporal.TaskQueue) — the
	// same queue all.go's internally-built trigger dispatches to — so a dispatched
	// rerun workflow is actually processable (not left orphaned at StatusPending).
	w := worker.New(tc, temporal.TaskQueue, worker.Options{})
	w.RegisterWorkflow(temporal.ExecuteGraphWorkflow)
	w.RegisterWorkflow(temporal.ExecuteBranchUntilConvergence)
	registry := workflow.NewActionRegistry()
	asyncRegistry := temporal.NewAsyncRegistry()
	execStore := registerActions(db, registry, asyncRegistry)
	activities := &temporal.Activities{
		Registry:       registry,
		AsyncRegistry:  asyncRegistry,
		ExecutionStore: execStore, // nil for the core path -> MarkExecution* no-op
	}
	w.RegisterActivity(activities)
	if err := w.Start(); err != nil {
		tc.Close()
		authServer.Close()
		t.Fatalf("starting temporal worker: %s", err)
	}

	// -------------------------------------------------------------------------

	ichorMux := mux.WebAPI(mux.Config{
		Log:            db.Log,
		AuthClient:     authClient,
		DB:             db.DB,
		TemporalClient: tc, // enables all.go's trigger -> executionapi.Config.Trigger wiring
	}, ichorbuild.Routes())

	t.Cleanup(func() {
		w.Stop()
		tc.Close()
		authServer.Close()
	})

	return New(db, ath, ichorMux)
}

// registerCoreWorkerActions installs the 4 core sync handlers used by
// StartTestWithTemporal's rerun-dispatch assertion. It returns a nil execution
// store: that test asserts only on the returned fresh id, not on the execution
// record's eventual status, so the MarkExecution* lifecycle activities can no-op.
func registerCoreWorkerActions(db *dbtest.Database, reg *workflow.ActionRegistry, _ *temporal.AsyncRegistry) temporal.ExecutionLifecycleStore {
	reg.Register(communication.NewSendEmailHandler(db.Log, db.DB, nil, ""))
	reg.Register(communication.NewSendNotificationHandler(db.Log, nil))
	reg.Register(control.NewEvaluateConditionHandler(db.Log))
	reg.Register(integration.NewCallWebhookHandler(db.Log))
	return nil
}

// registerGranularWorkerActions installs the core handlers plus the granular
// inventory pipeline handlers (Rule 5) so a dispatched/re-run Granular Inventory
// Pipeline workflow can run end-to-end. create_alert + seek_approval get the real
// alert/approval buses so the over_order alert and approval-hold records persist.
//
// It returns a real execution-lifecycle store (a workflowdb.Store) so the worker's
// MarkExecution* activities advance the automation_executions record
// pending -> running -> completed, mirroring the production worker
// (api/cmd/services/workflow-worker/main.go). Without it the record would stay at
// its created status even though the workflow completes — the over-order e2e test
// asserts the re-run execution reaches completed.
func registerGranularWorkerActions(db *dbtest.Database, reg *workflow.ActionRegistry, asyncReg *temporal.AsyncRegistry) temporal.ExecutionLifecycleStore {
	registerCoreWorkerActions(db, reg, asyncReg)

	// Granular inventory pipeline: check_inventory gates reserve_inventory; reserve
	// soft-fails to the over_order branch on a shortfall, succeeds against live stock
	// after restock.
	reg.Register(inventory.NewCheckInventoryHandler(db.Log, db.BusDomain.InventoryItem))
	reg.Register(inventory.NewCheckReorderPointHandler(db.Log, db.BusDomain.InventoryItem))
	reg.Register(inventory.NewReserveInventoryHandler(db.Log, db.DB, db.BusDomain.InventoryItem, db.BusDomain.Workflow))

	// create_alert with the real alert bus so the over_order alert row persists
	// (scoped/queryable by SourceRuleID).
	reg.Register(communication.NewCreateAlertHandler(db.Log, db.BusDomain.Alert, nil))

	// seek_approval is async; give it the real approval + alert buses so the hold
	// record is created when the over_order alert routes into it.
	asyncReg.Register("seek_approval", approval.NewSeekApprovalHandler(db.Log, db.DB, db.BusDomain.ApprovalRequest, db.BusDomain.Alert, nil))

	return workflowdb.NewStore(db.Log, db.DB)
}
