# User Preferences Domain Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a per-user key-value preference store (`userpreferencesbus`) following the 7-layer domain pattern, with upsert semantics, user-scoped authorization, frontend seed data, and integration tests.

**Architecture:** Composite PK `(user_id, key)` with `INSERT ... ON CONFLICT DO UPDATE` for upsert. No delegate events, no pagination, no QueryFilter. Authorization via `mid.AuthorizeUser` with `auth.RuleAdminOrSubject` — workers access their own prefs, admins access anyone's.

**Tech Stack:** Go 1.23, PostgreSQL 16.4, sqlx, Ardan Labs Service SDK

**Spec:** `docs/superpowers/specs/2026-04-10-userpreferencesbus-design.md`

---

## Phase 1: Database Migration

**Files:**
- Modify: `business/sdk/migrate/sql/migrate.sql` (append after line 2421)

- [ ] **Step 1: Append migration v2.28**

Add to the end of `business/sdk/migrate/sql/migrate.sql`:

```sql
-- Version: 2.28
-- Description: Add per-user key-value preference store for floor typography
--              scale picker and future user-scoped settings.

CREATE TABLE core.user_preferences (
    user_id      UUID         NOT NULL REFERENCES core.users(user_id) ON DELETE CASCADE,
    key          VARCHAR(100) NOT NULL,
    value        JSONB        NOT NULL,
    updated_date TIMESTAMP    NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, key)
);

CREATE INDEX idx_user_preferences_user_id ON core.user_preferences(user_id);
```

- [ ] **Step 2: Verify the migration compiles**

```bash
go build ./business/sdk/migrate/...
```

Expected: clean build, no errors.

- [ ] **Step 3: Commit**

```bash
git add business/sdk/migrate/sql/migrate.sql
git commit -m "feat(migrate): add core.user_preferences table (v2.28)"
```

---

## Phase 2: Business Layer (Bus + DB Store)

**Files:**
- Create: `business/domain/core/userpreferencesbus/userpreferencesbus.go`
- Create: `business/domain/core/userpreferencesbus/model.go`
- Create: `business/domain/core/userpreferencesbus/stores/userpreferencesdb/userpreferencesdb.go`
- Create: `business/domain/core/userpreferencesbus/stores/userpreferencesdb/model.go`

**Reference:** Read `business/domain/config/settingsbus/settingsbus.go` and `business/domain/config/settingsbus/stores/settingsdb/settingsdb.go` for the pattern. This domain is simpler — no delegate, no pagination, no QueryFilter, no ordering.

- [ ] **Step 1: Create `business/domain/core/userpreferencesbus/model.go`**

```go
package userpreferencesbus

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// UserPreference represents a single user preference key-value pair.
type UserPreference struct {
	UserID      uuid.UUID       `json:"user_id"`
	Key         string          `json:"key"`
	Value       json.RawMessage `json:"value"`
	UpdatedDate time.Time       `json:"updated_date"`
}

// NewUserPreference contains fields required to set a preference.
type NewUserPreference struct {
	UserID uuid.UUID       `json:"user_id"`
	Key    string          `json:"key"`
	Value  json.RawMessage `json:"value"`
}
```

- [ ] **Step 2: Create `business/domain/core/userpreferencesbus/userpreferencesbus.go`**

```go
package userpreferencesbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound = errors.New("user preference not found")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Upsert(ctx context.Context, pref UserPreference) error
	Delete(ctx context.Context, userID uuid.UUID, key string) error
	QueryByUser(ctx context.Context, userID uuid.UUID) ([]UserPreference, error)
	QueryByUserAndKey(ctx context.Context, userID uuid.UUID, key string) (UserPreference, error)
}

// Business manages the set of APIs for user preference access.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a user preferences business API for use.
func NewBusiness(log *logger.Logger, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
	}
}

// NewWithTx constructs a new Business value replacing the Storer
// value with a Storer value that is currently inside a transaction.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	return &Business{
		log:    b.log,
		storer: storer,
	}, nil
}

// Set upserts a single preference for a user.
func (b *Business) Set(ctx context.Context, np NewUserPreference) (UserPreference, error) {
	ctx, span := otel.AddSpan(ctx, "business.userpreferencesbus.set")
	defer span.End()

	now := time.Now()

	pref := UserPreference{
		UserID:      np.UserID,
		Key:         np.Key,
		Value:       np.Value,
		UpdatedDate: now,
	}

	if err := b.storer.Upsert(ctx, pref); err != nil {
		return UserPreference{}, fmt.Errorf("set: %w", err)
	}

	return pref, nil
}

// Get retrieves a single preference by user ID and key.
func (b *Business) Get(ctx context.Context, userID uuid.UUID, key string) (UserPreference, error) {
	ctx, span := otel.AddSpan(ctx, "business.userpreferencesbus.get")
	defer span.End()

	pref, err := b.storer.QueryByUserAndKey(ctx, userID, key)
	if err != nil {
		return UserPreference{}, fmt.Errorf("get: %w", err)
	}

	return pref, nil
}

// GetAll retrieves all preferences for a user.
func (b *Business) GetAll(ctx context.Context, userID uuid.UUID) ([]UserPreference, error) {
	ctx, span := otel.AddSpan(ctx, "business.userpreferencesbus.getall")
	defer span.End()

	prefs, err := b.storer.QueryByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getall: %w", err)
	}

	return prefs, nil
}

// Delete removes a single preference by user ID and key.
func (b *Business) Delete(ctx context.Context, userID uuid.UUID, key string) error {
	ctx, span := otel.AddSpan(ctx, "business.userpreferencesbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, userID, key); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}
```

- [ ] **Step 3: Create `business/domain/core/userpreferencesbus/stores/userpreferencesdb/model.go`**

```go
package userpreferencesdb

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus"
)

type userPreference struct {
	UserID      uuid.UUID       `db:"user_id"`
	Key         string          `db:"key"`
	Value       json.RawMessage `db:"value"`
	UpdatedDate time.Time       `db:"updated_date"`
}

func toBusUserPreference(db userPreference) userpreferencesbus.UserPreference {
	return userpreferencesbus.UserPreference{
		UserID:      db.UserID,
		Key:         db.Key,
		Value:       db.Value,
		UpdatedDate: db.UpdatedDate,
	}
}

func toBusUserPreferences(dbs []userPreference) []userpreferencesbus.UserPreference {
	bus := make([]userpreferencesbus.UserPreference, len(dbs))
	for i, db := range dbs {
		bus[i] = toBusUserPreference(db)
	}
	return bus
}

func toDBUserPreference(bus userpreferencesbus.UserPreference) userPreference {
	return userPreference{
		UserID:      bus.UserID,
		Key:         bus.Key,
		Value:       bus.Value,
		UpdatedDate: bus.UpdatedDate,
	}
}
```

- [ ] **Step 4: Create `business/domain/core/userpreferencesbus/stores/userpreferencesdb/userpreferencesdb.go`**

```go
package userpreferencesdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for user preference database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the API for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (userpreferencesbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	return &Store{
		log: s.log,
		db:  ec,
	}, nil
}

// Upsert inserts or updates a user preference.
func (s *Store) Upsert(ctx context.Context, pref userpreferencesbus.UserPreference) error {
	const q = `
    INSERT INTO core.user_preferences (
        user_id, key, value, updated_date
    ) VALUES (
        :user_id, :key, :value, :updated_date
    )
    ON CONFLICT (user_id, key) DO UPDATE SET
        value        = EXCLUDED.value,
        updated_date = EXCLUDED.updated_date`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUserPreference(pref)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a user preference by user ID and key.
func (s *Store) Delete(ctx context.Context, userID uuid.UUID, key string) error {
	data := struct {
		UserID uuid.UUID `db:"user_id"`
		Key    string    `db:"key"`
	}{
		UserID: userID,
		Key:    key,
	}

	const q = `DELETE FROM core.user_preferences WHERE user_id = :user_id AND key = :key`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryByUser retrieves all preferences for a given user.
func (s *Store) QueryByUser(ctx context.Context, userID uuid.UUID) ([]userpreferencesbus.UserPreference, error) {
	data := struct {
		UserID uuid.UUID `db:"user_id"`
	}{
		UserID: userID,
	}

	const q = `
    SELECT
        user_id, key, value, updated_date
    FROM
        core.user_preferences
    WHERE
        user_id = :user_id
    ORDER BY
        key`

	var dbPrefs []userPreference
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbPrefs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusUserPreferences(dbPrefs), nil
}

// QueryByUserAndKey retrieves a single preference by user ID and key.
func (s *Store) QueryByUserAndKey(ctx context.Context, userID uuid.UUID, key string) (userpreferencesbus.UserPreference, error) {
	data := struct {
		UserID uuid.UUID `db:"user_id"`
		Key    string    `db:"key"`
	}{
		UserID: userID,
		Key:    key,
	}

	const q = `
    SELECT
        user_id, key, value, updated_date
    FROM
        core.user_preferences
    WHERE
        user_id = :user_id AND key = :key`

	var dbPref userPreference
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbPref); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return userpreferencesbus.UserPreference{}, userpreferencesbus.ErrNotFound
		}
		return userpreferencesbus.UserPreference{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusUserPreference(dbPref), nil
}
```

- [ ] **Step 5: Verify build**

```bash
go build ./business/domain/core/userpreferencesbus/...
```

Expected: clean build.

- [ ] **Step 6: Commit**

```bash
git add business/domain/core/userpreferencesbus/
git commit -m "feat(userpreferencesbus): add business and DB store layers"
```

---

## Phase 3: App Layer

**Files:**
- Create: `app/domain/core/userpreferencesapp/userpreferencesapp.go`
- Create: `app/domain/core/userpreferencesapp/model.go`

**Reference:** Read `app/domain/config/settingsapp/settingsapp.go` and `app/domain/config/settingsapp/model.go` for the pattern. This domain is simpler — no QueryParams, no pagination, no filter/order parsing.

- [ ] **Step 1: Create `app/domain/core/userpreferencesapp/model.go`**

```go
package userpreferencesapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus"
)

// UserPreference represents a user preference for API responses.
type UserPreference struct {
	UserID      string          `json:"user_id"`
	Key         string          `json:"key"`
	Value       json.RawMessage `json:"value"`
	UpdatedDate string          `json:"updated_date"`
}

// Encode implements the web.Encoder interface.
func (app UserPreference) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppUserPreference converts a bus UserPreference to an app UserPreference.
func ToAppUserPreference(bus userpreferencesbus.UserPreference) UserPreference {
	return UserPreference{
		UserID:      bus.UserID.String(),
		Key:         bus.Key,
		Value:       bus.Value,
		UpdatedDate: bus.UpdatedDate.Format("2006-01-02T15:04:05Z"),
	}
}

// ToAppUserPreferences converts a slice of bus UserPreferences to app UserPreferences.
func ToAppUserPreferences(bus []userpreferencesbus.UserPreference) []UserPreference {
	app := make([]UserPreference, len(bus))
	for i, v := range bus {
		app[i] = ToAppUserPreference(v)
	}
	return app
}

// UserPreferences is a collection wrapper that implements the Encoder interface.
type UserPreferences struct {
	Items []UserPreference `json:"items"`
}

// Encode implements the web.Encoder interface.
func (app UserPreferences) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// NewUserPreference contains the value for setting a preference.
// The user_id and key come from path parameters, not the request body.
type NewUserPreference struct {
	Value json.RawMessage `json:"value" validate:"required"`
}

// Decode implements the web.Decoder interface.
func (app *NewUserPreference) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app NewUserPreference) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}
```

- [ ] **Step 2: Create `app/domain/core/userpreferencesapp/userpreferencesapp.go`**

```go
package userpreferencesapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus"
)

// App manages the set of app layer APIs for user preferences.
type App struct {
	userPreferencesBus *userpreferencesbus.Business
}

// NewApp constructs a user preferences app API for use.
func NewApp(userPreferencesBus *userpreferencesbus.Business) *App {
	return &App{
		userPreferencesBus: userPreferencesBus,
	}
}

// Set upserts a single preference for a user.
func (a *App) Set(ctx context.Context, userID uuid.UUID, key string, app NewUserPreference) (UserPreference, error) {
	np := userpreferencesbus.NewUserPreference{
		UserID: userID,
		Key:    key,
		Value:  app.Value,
	}

	pref, err := a.userPreferencesBus.Set(ctx, np)
	if err != nil {
		return UserPreference{}, fmt.Errorf("set: %w", err)
	}

	return ToAppUserPreference(pref), nil
}

// Get retrieves a single preference by user ID and key.
func (a *App) Get(ctx context.Context, userID uuid.UUID, key string) (UserPreference, error) {
	pref, err := a.userPreferencesBus.Get(ctx, userID, key)
	if err != nil {
		if errors.Is(err, userpreferencesbus.ErrNotFound) {
			return UserPreference{}, errs.New(errs.NotFound, err)
		}
		return UserPreference{}, fmt.Errorf("get: %w", err)
	}

	return ToAppUserPreference(pref), nil
}

// GetAll retrieves all preferences for a user.
func (a *App) GetAll(ctx context.Context, userID uuid.UUID) (UserPreferences, error) {
	prefs, err := a.userPreferencesBus.GetAll(ctx, userID)
	if err != nil {
		return UserPreferences{}, fmt.Errorf("getall: %w", err)
	}

	return UserPreferences{
		Items: ToAppUserPreferences(prefs),
	}, nil
}

// Delete removes a single preference by user ID and key.
func (a *App) Delete(ctx context.Context, userID uuid.UUID, key string) error {
	if err := a.userPreferencesBus.Delete(ctx, userID, key); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}
```

- [ ] **Step 3: Verify build**

```bash
go build ./app/domain/core/userpreferencesapp/...
```

Expected: clean build.

- [ ] **Step 4: Commit**

```bash
git add app/domain/core/userpreferencesapp/
git commit -m "feat(userpreferencesapp): add app layer for user preferences"
```

---

## Phase 4: API Layer + Build Wiring

**Files:**
- Create: `api/domain/http/core/userpreferencesapi/userpreferencesapi.go`
- Create: `api/domain/http/core/userpreferencesapi/routes.go`
- Modify: `api/cmd/services/ichor/build/all/all.go`
- Modify: `api/cmd/services/ichor/build/crud/crud.go`

**Reference:** Read `api/domain/http/config/settingsapi/settingsapi.go` and `api/domain/http/config/settingsapi/routes.go` for the handler/route pattern. Read `api/domain/http/core/userapi/route.go` lines 32-33 for the `AuthorizeUser` middleware usage. Read the imports and wiring sections in `all.go` and `crud.go` around where `settingsBus` is created and `settingsapi.Routes` is called.

- [ ] **Step 1: Create `api/domain/http/core/userpreferencesapi/userpreferencesapi.go`**

```go
package userpreferencesapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/core/userpreferencesapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	userPreferencesApp *userpreferencesapp.App
}

func newAPI(userPreferencesApp *userpreferencesapp.App) *api {
	return &api{
		userPreferencesApp: userPreferencesApp,
	}
}

func (api *api) set(ctx context.Context, r *http.Request) web.Encoder {
	var app userpreferencesapp.NewUserPreference
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := uuid.Parse(web.Param(r, "user_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	key := web.Param(r, "key")

	pref, err := api.userPreferencesApp.Set(ctx, userID, key, app)
	if err != nil {
		return errs.NewError(err)
	}

	return pref
}

func (api *api) get(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := uuid.Parse(web.Param(r, "user_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	key := web.Param(r, "key")

	pref, err := api.userPreferencesApp.Get(ctx, userID, key)
	if err != nil {
		return errs.NewError(err)
	}

	return pref
}

func (api *api) getAll(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := uuid.Parse(web.Param(r, "user_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	prefs, err := api.userPreferencesApp.GetAll(ctx, userID)
	if err != nil {
		return errs.NewError(err)
	}

	return prefs
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := uuid.Parse(web.Param(r, "user_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	key := web.Param(r, "key")

	if err := api.userPreferencesApp.Delete(ctx, userID, key); err != nil {
		return errs.NewError(err)
	}

	return nil
}
```

- [ ] **Step 2: Create `api/domain/http/core/userpreferencesapi/routes.go`**

```go
package userpreferencesapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/core/userpreferencesapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	UserPreferencesBus *userpreferencesbus.Business
	AuthClient         *authclient.Client
	UserBus            *userbus.Business
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAuthorizeUser := mid.AuthorizeUser(cfg.AuthClient, cfg.UserBus, auth.RuleAdminOrSubject)

	api := newAPI(userpreferencesapp.NewApp(cfg.UserPreferencesBus))

	app.HandlerFunc(http.MethodPut, version, "/users/{user_id}/preferences/{key}", api.set, authen, ruleAuthorizeUser)
	app.HandlerFunc(http.MethodGet, version, "/users/{user_id}/preferences/{key}", api.get, authen, ruleAuthorizeUser)
	app.HandlerFunc(http.MethodGet, version, "/users/{user_id}/preferences", api.getAll, authen, ruleAuthorizeUser)
	app.HandlerFunc(http.MethodDelete, version, "/users/{user_id}/preferences/{key}", api.delete, authen, ruleAuthorizeUser)
}
```

- [ ] **Step 3: Wire into `api/cmd/services/ichor/build/all/all.go`**

Add these imports (find the existing import block, add alphabetically within the relevant sections):

```go
// In the api imports section (around line 26):
"github.com/timmaaaz/ichor/api/domain/http/core/userpreferencesapi"

// In the business imports section (around line 186-187):
"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus"
"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus/stores/userpreferencesdb"
```

Add the bus creation near the other core domain bus creations (after `settingsBus` creation around line 476):

```go
userPreferencesBus := userpreferencesbus.NewBusiness(cfg.Log, userpreferencesdb.NewStore(cfg.Log, cfg.DB))
```

Add the route registration near the other core domain routes (after `settingsapi.Routes` around line 1312):

```go
userpreferencesapi.Routes(app, userpreferencesapi.Config{
    UserPreferencesBus: userPreferencesBus,
    AuthClient:         cfg.AuthClient,
    UserBus:            userBus,
})
```

- [ ] **Step 4: Wire into `api/cmd/services/ichor/build/crud/crud.go`**

Add the same imports as all.go (find the import block, add alphabetically):

```go
// In the api imports section:
"github.com/timmaaaz/ichor/api/domain/http/core/userpreferencesapi"

// In the business imports section:
"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus"
"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus/stores/userpreferencesdb"
```

Add the bus creation (after `settingsBus` creation around line 258):

```go
userPreferencesBus := userpreferencesbus.NewBusiness(cfg.Log, userpreferencesdb.NewStore(cfg.Log, cfg.DB))
```

Add the route registration (after `settingsapi.Routes` around line 613):

```go
userpreferencesapi.Routes(app, userpreferencesapi.Config{
    UserPreferencesBus: userPreferencesBus,
    AuthClient:         cfg.AuthClient,
    UserBus:            userBus,
})
```

- [ ] **Step 5: Verify full server build**

```bash
go build ./api/cmd/services/ichor/...
```

Expected: clean build. This proves all wiring compiles.

- [ ] **Step 6: Commit**

```bash
git add api/domain/http/core/userpreferencesapi/ api/cmd/services/ichor/build/all/all.go api/cmd/services/ichor/build/crud/crud.go
git commit -m "feat(userpreferencesapi): add API layer and wire into build"
```

---

## Phase 5: Seed Data + Frontend Seed

**Files:**
- Create: `business/domain/core/userpreferencesbus/testutil.go`
- Create: `business/sdk/dbtest/seed_userpreferences.go`
- Modify: `business/sdk/dbtest/dbtest.go` (add `UserPreferences` field to `BusDomain`, wire in `newBusDomains`)
- Modify: `business/sdk/dbtest/seedFrontend.go` (add `seedUserPreferences` call)

**Reference:** Read `business/domain/config/settingsbus/testutil.go` for the TestSeed pattern. Read `business/sdk/dbtest/dbtest.go` for the `BusDomain` struct (starts around line 185) and `newBusDomains` function (starts around line 302). Read `business/sdk/dbtest/seedFrontend.go` for the `InsertSeedData` function call order.

- [ ] **Step 1: Create `business/domain/core/userpreferencesbus/testutil.go`**

```go
package userpreferencesbus

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/google/uuid"
)

// TestSeedUserPreferences seeds user preferences for the provided user IDs.
// Each user gets a "floor.font_scale" preference set to "medium".
func TestSeedUserPreferences(ctx context.Context, userIDs uuid.UUIDs, api *Business) ([]UserPreference, error) {
	prefs := make([]UserPreference, 0, len(userIDs))

	for _, userID := range userIDs {
		np := NewUserPreference{
			UserID: userID,
			Key:    "floor.font_scale",
			Value:  json.RawMessage(`"medium"`),
		}

		pref, err := api.Set(ctx, np)
		if err != nil {
			return nil, err
		}

		prefs = append(prefs, pref)
	}

	sort.Slice(prefs, func(i, j int) bool {
		if prefs[i].UserID.String() == prefs[j].UserID.String() {
			return prefs[i].Key < prefs[j].Key
		}
		return prefs[i].UserID.String() < prefs[j].UserID.String()
	})

	return prefs, nil
}
```

- [ ] **Step 2: Modify `business/sdk/dbtest/dbtest.go` — add `UserPreferences` to `BusDomain`**

Find the `BusDomain` struct definition (around line 185). Add this field in the Core/Users section alongside the other core domain fields:

```go
UserPreferences *userpreferencesbus.Business
```

Add the import:

```go
"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus"
"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus/stores/userpreferencesdb"
```

Find the `newBusDomains` function (around line 302). In the core section (near where `userBus` is created), add:

```go
userPreferencesBus := userpreferencesbus.NewBusiness(log, userpreferencesdb.NewStore(log, db))
```

Find where the `BusDomain` struct is populated in the return statement. Add:

```go
UserPreferences: userPreferencesBus,
```

- [ ] **Step 3: Create `business/sdk/dbtest/seed_userpreferences.go`**

```go
package dbtest

import (
	"context"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus"
)

func seedUserPreferences(ctx context.Context, busDomain BusDomain, userIDs uuid.UUIDs) error {
	_, err := userpreferencesbus.TestSeedUserPreferences(ctx, userIDs, busDomain.UserPreferences)
	if err != nil {
		return err
	}

	return nil
}
```

- [ ] **Step 4: Modify `business/sdk/dbtest/seedFrontend.go`**

Find the `InsertSeedData` function. After the `seedFoundation` call (around line 26-30), add the `seedUserPreferences` call. The foundation seed returns user IDs that we need. Look at how `seedFoundation` returns its data — it returns a `FoundationSeed` struct. Extract all user IDs from the admin, reporters, and bosses slices.

Add after `seedFoundation` and before `seedGeographyHR`:

```go
// Seed user preferences for all seeded users.
var allUserIDs uuid.UUIDs
for _, u := range foundationSeed.Admins {
    allUserIDs = append(allUserIDs, u.ID)
}
for _, u := range foundationSeed.Reporters {
    allUserIDs = append(allUserIDs, u.ID)
}
for _, u := range foundationSeed.Bosses {
    allUserIDs = append(allUserIDs, u.ID)
}

if err := seedUserPreferences(ctx, busDomain, allUserIDs); err != nil {
    return fmt.Errorf("seeding user preferences: %w", err)
}
```

**Important:** Read `seedFrontend.go` first to confirm the exact variable names for the foundation seed return value and the user slice field names. The above code assumes `foundationSeed` with `.Admins`, `.Reporters`, `.Bosses` — verify and adjust field names if different.

- [ ] **Step 5: Verify build**

```bash
go build ./business/sdk/dbtest/... ./business/domain/core/userpreferencesbus/...
```

Expected: clean build.

- [ ] **Step 6: Commit**

```bash
git add business/domain/core/userpreferencesbus/testutil.go business/sdk/dbtest/seed_userpreferences.go business/sdk/dbtest/dbtest.go business/sdk/dbtest/seedFrontend.go
git commit -m "feat(seed): add user preferences seed data and frontend seed integration"
```

---

## Phase 6: Integration Tests

**Files:**
- Create: `api/cmd/services/ichor/tests/core/userpreferencesapi/userpreferencesapi_test.go`
- Create: `api/cmd/services/ichor/tests/core/userpreferencesapi/seed_test.go`
- Create: `api/cmd/services/ichor/tests/core/userpreferencesapi/set_test.go`
- Create: `api/cmd/services/ichor/tests/core/userpreferencesapi/get_test.go`
- Create: `api/cmd/services/ichor/tests/core/userpreferencesapi/delete_test.go`

**Reference:** Read `api/cmd/services/ichor/tests/core/currencyapi/currency_test.go`, `seed_test.go`, `create_test.go`, `query_test.go`, and `delete_test.go` for the exact test framework pattern. The pattern uses `apitest.StartTest`, `insertSeedData`, and `test.Run` with `apitest.Table` slices.

- [ ] **Step 1: Create `api/cmd/services/ichor/tests/core/userpreferencesapi/seed_test.go`**

```go
package userpreferencesapi_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/userpreferencesapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// Seed admin user
	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admin users: %w", err)
	}
	admin := apitest.User{
		User:  admins[0],
		Token: apitest.Token(db.BusDomain.User, ath, admins[0].Email.Address),
	}

	// Seed regular user
	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding regular users: %w", err)
	}
	user := apitest.User{
		User:  users[0],
		Token: apitest.Token(db.BusDomain.User, ath, users[0].Email.Address),
	}

	// =========================================================================
	// Permissions setup
	// =========================================================================
	roles, err := rolebus.TestSeedRoles(ctx, 2, busDomain.Role)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding roles: %w", err)
	}

	roleIDs := make(uuid.UUIDs, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	userIDs := uuid.UUIDs{admin.ID, user.ID}

	_, err = userrolebus.TestSeedUserRoles(ctx, userIDs, roleIDs, busDomain.UserRole)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user roles: %w", err)
	}

	_, err = tableaccessbus.TestSeedTableAccess(ctx, roleIDs, busDomain.TableAccess)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding table access: %w", err)
	}

	// =========================================================================
	// Seed user preferences
	// =========================================================================
	_, err = userpreferencesbus.TestSeedUserPreferences(ctx, uuid.UUIDs{admin.ID}, busDomain.UserPreferences)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user preferences: %w", err)
	}

	return apitest.SeedData{
		Admins: []apitest.User{admin},
		Users:  []apitest.User{user},
	}, nil
}

// toAppUserPreference is available if tests need to convert seeded bus
// preferences to app-layer expected values.
func toAppUserPreference(bus userpreferencesbus.UserPreference) userpreferencesapp.UserPreference {
	return userpreferencesapp.ToAppUserPreference(bus)
}
```

- [ ] **Step 2: Create `api/cmd/services/ichor/tests/core/userpreferencesapi/userpreferencesapi_test.go`**

```go
package userpreferencesapi_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_UserPreferencesAPI(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_UserPreferencesAPI")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("seeding error: %s", err)
	}

	test.Run(t, set200(sd), "set-200")
	test.Run(t, get200(sd), "get-200")
	test.Run(t, getAll200(sd), "getall-200")
	test.Run(t, delete200(sd), "delete-200")
}
```

- [ ] **Step 3: Create `api/cmd/services/ichor/tests/core/userpreferencesapi/set_test.go`**

```go
package userpreferencesapi_test

import (
	"encoding/json"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/userpreferencesapp"
)

func set200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "create-new",
			URL:        "/v1/users/" + sd.Admins[0].ID.String() + "/preferences/floor.theme",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPut,
			Input: &userpreferencesapp.NewUserPreference{
				Value: json.RawMessage(`"dark"`),
			},
			GotResp: &userpreferencesapp.UserPreference{},
			ExpResp: &userpreferencesapp.UserPreference{
				UserID: sd.Admins[0].ID.String(),
				Key:    "floor.theme",
				Value:  json.RawMessage(`"dark"`),
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*userpreferencesapp.UserPreference)
				expResp := exp.(*userpreferencesapp.UserPreference)

				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:       "upsert-existing",
			URL:        "/v1/users/" + sd.Admins[0].ID.String() + "/preferences/floor.font_scale",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPut,
			Input: &userpreferencesapp.NewUserPreference{
				Value: json.RawMessage(`"large"`),
			},
			GotResp: &userpreferencesapp.UserPreference{},
			ExpResp: &userpreferencesapp.UserPreference{
				UserID: sd.Admins[0].ID.String(),
				Key:    "floor.font_scale",
				Value:  json.RawMessage(`"large"`),
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*userpreferencesapp.UserPreference)
				expResp := exp.(*userpreferencesapp.UserPreference)

				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}
```

- [ ] **Step 4: Create `api/cmd/services/ichor/tests/core/userpreferencesapi/get_test.go`**

```go
package userpreferencesapi_test

import (
	"encoding/json"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/userpreferencesapp"
)

func get200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/users/" + sd.Admins[0].ID.String() + "/preferences/floor.font_scale",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &userpreferencesapp.UserPreference{},
			ExpResp: &userpreferencesapp.UserPreference{
				UserID: sd.Admins[0].ID.String(),
				Key:    "floor.font_scale",
				Value:  json.RawMessage(`"medium"`),
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*userpreferencesapp.UserPreference)
				expResp := exp.(*userpreferencesapp.UserPreference)

				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func getAll200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/users/" + sd.Admins[0].ID.String() + "/preferences",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &userpreferencesapp.UserPreferences{},
			ExpResp: &userpreferencesapp.UserPreferences{
				Items: []userpreferencesapp.UserPreference{
					{
						UserID: sd.Admins[0].ID.String(),
						Key:    "floor.font_scale",
						Value:  json.RawMessage(`"medium"`),
					},
				},
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*userpreferencesapp.UserPreferences)
				expResp := exp.(*userpreferencesapp.UserPreferences)

				for i := range expResp.Items {
					if i < len(gotResp.Items) {
						expResp.Items[i].UpdatedDate = gotResp.Items[i].UpdatedDate
					}
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}
```

- [ ] **Step 5: Create `api/cmd/services/ichor/tests/core/userpreferencesapi/delete_test.go`**

```go
package userpreferencesapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func delete200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/users/" + sd.Admins[0].ID.String() + "/preferences/floor.font_scale",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusNoContent,
			Method:     http.MethodDelete,
			GotResp:    nil,
			ExpResp:    nil,
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
```

- [ ] **Step 6: Verify build**

```bash
go build ./api/cmd/services/ichor/tests/core/userpreferencesapi/...
```

Expected: clean build (tests compile but don't run without a database).

- [ ] **Step 7: Run integration tests**

```bash
go test ./api/cmd/services/ichor/tests/core/userpreferencesapi/... -v -count=1
```

Expected: all tests PASS. If tests fail, debug the specific failure — common issues:
- JSON response shape mismatch (check Encode methods)
- Auth middleware rejecting requests (verify `AuthorizeUser` wiring)
- SQL errors (check column names match migration)

- [ ] **Step 8: Commit**

```bash
git add api/cmd/services/ichor/tests/core/userpreferencesapi/
git commit -m "test(userpreferencesapi): add integration tests for user preferences"
```

---

## Post-Implementation Checklist

After all 6 phases complete:

- [ ] `go build ./...` passes (full repo build)
- [ ] `go test ./api/cmd/services/ichor/tests/core/userpreferencesapi/... -v -count=1` passes
- [ ] `go build ./business/domain/core/userpreferencesbus/...` passes
- [ ] `go build ./app/domain/core/userpreferencesapp/...` passes
- [ ] `go build ./api/domain/http/core/userpreferencesapi/...` passes
