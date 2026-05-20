package scenarios_test

import (
	"net/http/httptest"
	"testing"

	authbuild "github.com/timmaaaz/ichor/api/cmd/services/auth/build/all"
	ichorbuild "github.com/timmaaaz/ichor/api/cmd/services/ichor/build/all"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/api/sdk/http/mux"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

// startScenarioTest constructs an apitest.Test with ScenariosEnabled: true.
// This is the load-bearing fix — the existing apitest.StartTest defaults to
// ScenariosEnabled: false, making ApplyScenarioFilter a no-op throughout
// existing integration tests. That gap hid GB-006, GB-014, GB-015 from CI
// until manual smoke-verify (see session-log 2026-05-19 line 651).
//
// ⚠ Do not regress this — if a future refactor loses ScenariosEnabled,
// the harness becomes silent theater and the bugs hide again.
func startScenarioTest(t *testing.T, testName string) *apitest.Test {
	t.Helper()

	db := dbtest.NewDatabase(t, testName)

	// -------------------------------------------------------------------------

	auth, err := auth.New(auth.Config{
		Log:       db.Log,
		DB:        db.DB,
		KeyLookup: &apitest.KeyStore{},
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

	appMux := mux.WebAPI(mux.Config{
		Log:              db.Log,
		AuthClient:       authClient,
		DB:               db.DB,
		ScenariosEnabled: true, // ⚠ load-bearing — see comment above
	}, ichorbuild.Routes())

	return apitest.New(db, auth, appMux)
}
