# User Preferences Domain Design

**Date:** 2026-04-10
**Status:** Approved
**Area:** `core`

## Purpose

Per-user key-value preference store. First consumer is the floor typography scale picker (`floor.font_scale`), but the domain is generic — future keys include theme, language, notification settings, etc.

## Architecture

Follows the 7-layer pattern under `core` (scoped to `core.users`).

| Layer | Package |
|---|---|
| Business | `business/domain/core/userpreferencesbus/` |
| DB Store | `business/domain/core/userpreferencesbus/stores/userpreferencesdb/` |
| App | `app/domain/core/userpreferencesapp/` |
| API | `api/domain/http/core/userpreferencesapi/` |
| Tests | `api/cmd/services/ichor/tests/core/userpreferencesapi/` |

## Database — Migration v2.28

```sql
CREATE TABLE core.user_preferences (
    user_id      UUID         NOT NULL REFERENCES core.users(user_id) ON DELETE CASCADE,
    key          VARCHAR(100) NOT NULL,
    value        JSONB        NOT NULL,
    updated_date TIMESTAMP    NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, key)
);

CREATE INDEX idx_user_preferences_user_id ON core.user_preferences(user_id);
```

- Composite PK `(user_id, key)` — one row per user per preference key.
- No `created_date` — upsert semantics mean creation and update are the same operation.
- `ON DELETE CASCADE` — deleting a user cleans up their preferences.
- `JSONB` value column — stores any JSON-serializable preference value.

## Business Layer

### Models

```go
type UserPreference struct {
    UserID      uuid.UUID       `json:"user_id"`
    Key         string          `json:"key"`
    Value       json.RawMessage `json:"value"`
    UpdatedDate time.Time       `json:"updated_date"`
}

type NewUserPreference struct {
    UserID uuid.UUID       `json:"user_id"`
    Key    string          `json:"key"`
    Value  json.RawMessage `json:"value"`
}

type UpdateUserPreference struct {
    Value json.RawMessage `json:"value"`
}
```

### Storer Interface

```go
type Storer interface {
    NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
    Upsert(ctx context.Context, pref UserPreference) error
    Delete(ctx context.Context, userID uuid.UUID, key string) error
    QueryByUser(ctx context.Context, userID uuid.UUID) ([]UserPreference, error)
    QueryByUserAndKey(ctx context.Context, userID uuid.UUID, key string) (UserPreference, error)
}
```

`Upsert` uses `INSERT ... ON CONFLICT (user_id, key) DO UPDATE SET value = EXCLUDED.value, updated_date = EXCLUDED.updated_date`.

### Business Methods

```go
func (b *Business) Set(ctx context.Context, np NewUserPreference) (UserPreference, error)
func (b *Business) Get(ctx context.Context, userID uuid.UUID, key string) (UserPreference, error)
func (b *Business) GetAll(ctx context.Context, userID uuid.UUID) ([]UserPreference, error)
func (b *Business) Delete(ctx context.Context, userID uuid.UUID, key string) error
```

- No `Query` with pagination — preferences are per-user and will never be large enough to paginate.
- No delegate events — preferences are low-stakes, no workflow triggers needed.

## App Layer

Thin translation layer between HTTP models and business models.

### App Models

```go
type UserPreference struct {
    UserID      string          `json:"user_id"`
    Key         string          `json:"key"`
    Value       json.RawMessage `json:"value"`
    UpdatedDate string          `json:"updated_date"`
}

type NewUserPreference struct {
    Value json.RawMessage `json:"value" validate:"required"`
}
```

- `Set` decodes `NewUserPreference` from request body, combines with `user_id` and `key` from path params.
- Error mapping: bus `ErrNotFound` -> `errs.NotFound`.
- Timestamps formatted as RFC3339 strings.

## API Layer — Routes

| Method | Path | Handler | Middleware |
|---|---|---|---|
| `PUT` | `/v1/users/{user_id}/preferences/{key}` | `set` | `Authenticate` + `AuthorizeUser(RuleAdminOrSubject)` |
| `GET` | `/v1/users/{user_id}/preferences/{key}` | `get` | `Authenticate` + `AuthorizeUser(RuleAdminOrSubject)` |
| `GET` | `/v1/users/{user_id}/preferences` | `getAll` | `Authenticate` + `AuthorizeUser(RuleAdminOrSubject)` |
| `DELETE` | `/v1/users/{user_id}/preferences/{key}` | `delete` | `Authenticate` + `AuthorizeUser(RuleAdminOrSubject)` |

- Uses `mid.AuthorizeUser` with `auth.RuleAdminOrSubject` — workers can only access their own preferences, admins can access any user's.
- No `mid.Authorize` with `PermissionsBus` — ownership check is sufficient, no table-level RBAC.

### Request/Response

**PUT request body:**
```json
{"value": "\"large\""}
```

**GET single response:**
```json
{
    "user_id": "uuid",
    "key": "floor.font_scale",
    "value": "\"large\"",
    "updated_date": "2026-04-10T..."
}
```

**GET all response:**
```json
{
    "items": [
        {"user_id": "uuid", "key": "floor.font_scale", "value": "\"large\"", "updated_date": "..."},
        {"user_id": "uuid", "key": "floor.theme", "value": "\"dark\"", "updated_date": "..."}
    ]
}
```

## Wiring

Register bus, app, and API in `api/cmd/services/ichor/build/all/all.go` and `crud/crud.go` following the `settingsbus` pattern:

```go
userPreferencesBus := userpreferencesbus.NewBusiness(cfg.Log, userpreferencesdb.NewStore(cfg.Log, cfg.DB))

userpreferencesapi.Routes(app, userpreferencesapi.Config{
    UserPreferencesBus: userPreferencesBus,
    AuthClient:         cfg.AuthClient,
    UserBus:            userBus,
})
```

Note: No `delegate` dependency (no events). Config includes `UserBus` for `mid.AuthorizeUser`.

## Seed Data

In `userpreferencesbus/testutil.go`:

- `TestSeedUserPreferences` — seeds `floor.font_scale = "medium"` for provided seed users.
- Called from integration test seed setup.

## Testing

Integration tests in `api/cmd/services/ichor/tests/core/userpreferencesapi/`:

| File | Purpose |
|---|---|
| `userpreferencesapi_test.go` | Full HTTP round-trip tests |
| `model_test.go` | Test helpers for generating request bodies |
| `seed_test.go` | Seed data setup |

### Test Cases

1. **Set (create)** — PUT a new preference, verify 200 + returned value
2. **Set (upsert)** — PUT an existing preference with new value, verify update
3. **Get** — GET a single preference by user+key
4. **GetAll** — GET all preferences for a user, verify items array
5. **Delete** — DELETE a preference, verify subsequent GET returns 404
6. **Ownership enforcement** — User A tries to GET user B's preferences, verify 403
7. **Admin cross-user access** — Admin GETs another user's preferences, verify 200

## What's NOT Built

- No QueryFilter / pagination / ordering
- No delegate events
- No cache layer
- No `created_date` column
