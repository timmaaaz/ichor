package apitest

import (
	"net/http/httptest"
	"testing"

	authbuild "bitbucket.org/superiortechnologies/ichor/api/cmd/services/auth/build/all"
	ichorbuild "bitbucket.org/superiortechnologies/ichor/api/cmd/services/ichor/build/all"
	"bitbucket.org/superiortechnologies/ichor/api/sdk/http/mux"
	"bitbucket.org/superiortechnologies/ichor/app/sdk/auth"
	"bitbucket.org/superiortechnologies/ichor/app/sdk/authclient"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/dbtest"
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
