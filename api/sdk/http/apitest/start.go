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
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/control"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/integration"
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
	// The 4 core sync handlers cover the seeded rules' actions; async handlers are
	// unnecessary for the rerun path's assertions.
	w := worker.New(tc, temporal.TaskQueue, worker.Options{})
	w.RegisterWorkflow(temporal.ExecuteGraphWorkflow)
	w.RegisterWorkflow(temporal.ExecuteBranchUntilConvergence)
	registry := workflow.NewActionRegistry()
	registry.Register(communication.NewSendEmailHandler(db.Log, db.DB, nil, ""))
	registry.Register(communication.NewSendNotificationHandler(db.Log, nil))
	registry.Register(control.NewEvaluateConditionHandler(db.Log))
	registry.Register(integration.NewCallWebhookHandler(db.Log))
	activities := &temporal.Activities{
		Registry:      registry,
		AsyncRegistry: temporal.NewAsyncRegistry(),
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
