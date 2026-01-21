# Plan: Add QueryByCode to Currency Domain for FK Resolution

## Summary

Add `QueryByCode` method to `currencyapp` to enable FK default resolution where form fields specify default values by code (e.g., "USD") that need to be resolved to UUIDs. Register this method in the formdata registry and add comprehensive tests.

## Files to Modify

| File | Change |
|------|--------|
| [currencyapp.go](app/domain/core/currencyapp/currencyapp.go) | Add `QueryByCode` method |
| [formdata_registry.go](api/cmd/services/ichor/build/all/formdata_registry.go) | Add `QueryByNameFunc` to `core.currencies` registration |
| [currencyapp_test.go](app/domain/core/currencyapp/currencyapp_test.go) | **Create new file** - Unit tests for `QueryByCode` |
| [fkresolution_test.go](app/domain/formdata/formdataapp/fkresolution_test.go) | Add integration test for currency FK resolution |

---

## Implementation Steps

### Step 1: Add QueryByCode Method to currencyapp.go

**File**: `app/domain/core/currencyapp/currencyapp.go`

Add after `QueryAll` method (around line 137):

```go
// QueryByCode finds a currency by its code and returns the ID.
// This is used for FK default resolution where form fields specify default values
// by code (e.g., "USD") that need to be resolved to UUIDs.
func (a *App) QueryByCode(ctx context.Context, code string) (uuid.UUID, error) {
	filter := currencybus.QueryFilter{
		Code: &code,
	}

	currencies, err := a.currencybus.Query(ctx, filter, currencybus.DefaultOrderBy, page.MustParse("1", "1"))
	if err != nil {
		return uuid.Nil, fmt.Errorf("query by code: %w", err)
	}

	if len(currencies) == 0 {
		return uuid.Nil, errs.Newf(errs.NotFound, "currency with code %q not found", code)
	}

	return currencies[0].ID, nil
}
```

**Required import additions**:
- `"fmt"` (if not already present)

Note: `currencybus.DefaultOrderBy` is already available - this follows the pattern used in `lineitemfulfillmentstatusapp.QueryByName`.

---

### Step 2: Update formdata_registry.go

**File**: `api/cmd/services/ichor/build/all/formdata_registry.go`

Modify the `core.currencies` registration (lines 437-470). Add `QueryByNameFunc` after line 467 (before the closing `}`):

```go
		UpdateModel: currencyapp.UpdateCurrency{},
		QueryByNameFunc: func(ctx context.Context, name string) (uuid.UUID, error) {
			return currencyApp.QueryByCode(ctx, name)
		},
```

Note: The parameter name is `name` (not `code`) because the formdata registry interface uses `QueryByNameFunc` generically across all entities. For currencies, we interpret "name" as the currency code.

---

### Step 3: Create Unit Tests for currencyapp

**File**: `app/domain/core/currencyapp/currencyapp_test.go` (new file)

Create unit tests following the pattern from `fkresolution_test.go`:

```go
package currencyapp_test

import (
	"context"
	"testing"

	"github.com/timmaaaz/ichor/app/domain/core/currencyapp"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func Test_QueryByCode(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_CurrencyApp_QueryByCode")
	defer db.Teardown()

	ctx := context.Background()

	// Seed currencies with known codes
	currencies, err := currencybus.TestSeedCurrencies(ctx, 3, db.BusDomain.Currency)
	if err != nil {
		t.Fatalf("seeding currencies: %s", err)
	}

	app := currencyapp.NewApp(db.BusDomain.Currency)

	// Test: QueryByCode - success case
	t.Run("queryByCode-success", func(t *testing.T) {
		expectedCurrency := currencies[0]

		id, err := app.QueryByCode(ctx, expectedCurrency.Code)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if id != expectedCurrency.ID {
			t.Errorf("expected ID %v, got %v", expectedCurrency.ID, id)
		}
	})

	// Test: QueryByCode - not found case
	t.Run("queryByCode-notFound", func(t *testing.T) {
		_, err := app.QueryByCode(ctx, "NONEXISTENT")
		if err == nil {
			t.Error("expected error for non-existent code, got nil")
		}
	})
}
```

---

### Step 4: Add Integration Test for Currency FK Resolution

**File**: `app/domain/formdata/formdataapp/fkresolution_test.go`

Add after existing tests (around line 541):

```go
// TestResolveFKByName_CurrencyResolution verifies that currency FK resolution
// works correctly using QueryByCode.
func TestResolveFKByName_CurrencyResolution(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "trace" })

	expectedCurrencyID := uuid.New()

	// Create registry with currency entity that has QueryByNameFunc
	registry := formdataregistry.New()
	registry.Register(formdataregistry.EntityRegistration{
		Name: "core.currencies",
		QueryByNameFunc: func(ctx context.Context, name string) (uuid.UUID, error) {
			if name == "USD" {
				return expectedCurrencyID, nil
			}
			return uuid.Nil, fmt.Errorf("currency with code %q not found", name)
		},
	})

	app := &App{
		log:      log,
		registry: registry,
	}

	ctx := context.Background()

	dropdownConfig := &formfieldbus.DropdownConfig{
		Entity:      "core.currencies",
		LabelColumn: "name",
		ValueColumn: "id",
	}

	// Should successfully resolve "USD" to the expected UUID
	result, err := app.resolveFKByName(ctx, dropdownConfig, "USD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != expectedCurrencyID {
		t.Errorf("expected resolved ID %v, got %v", expectedCurrencyID, result)
	}
}

// TestMergeFieldDefaults_CurrencyFKResolutionSuccess verifies that mergeFieldDefaults
// successfully resolves currency FK defaults when the entity has QueryByNameFunc.
func TestMergeFieldDefaults_CurrencyFKResolutionSuccess(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "trace" })

	expectedCurrencyID := uuid.New()

	// Create registry with currency entity that has QueryByNameFunc
	registry := formdataregistry.New()
	registry.Register(formdataregistry.EntityRegistration{
		Name: "core.currencies",
		QueryByNameFunc: func(ctx context.Context, name string) (uuid.UUID, error) {
			if name == "USD" {
				return expectedCurrencyID, nil
			}
			return uuid.Nil, fmt.Errorf("currency with code %q not found", name)
		},
	})

	app := &App{
		log:      log,
		registry: registry,
	}

	ctx := context.Background()

	// Field WITH currency entity config and QueryByNameFunc registered
	fieldWithCurrency := formfieldbus.FormField{
		Name:   "currency_id",
		Config: []byte(`{"entity": "core.currencies", "label_column": "name", "value_column": "id", "default_value_create": "USD"}`),
	}

	fields := []formfieldbus.FormField{fieldWithCurrency}
	inputData := []byte(`{}`)

	result, injected, err := app.mergeFieldDefaults(ctx, inputData, fields, formdataregistry.OperationCreate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse result
	var resultMap map[string]interface{}
	if err := json.Unmarshal(result, &resultMap); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Field should be injected with the resolved UUID
	if currencyID, ok := resultMap["currency_id"]; !ok || currencyID != expectedCurrencyID.String() {
		t.Errorf("expected currency_id to be %s, got %v", expectedCurrencyID, resultMap["currency_id"])
	}

	// Verify it was tracked as injected
	if injected.InjectedFields["currency_id"] != expectedCurrencyID.String() {
		t.Errorf("expected currency_id in injected fields with value %s, got: %v", expectedCurrencyID, injected.InjectedFields)
	}
}
```

**Required import additions** (if not already present):
- `"fmt"`

---

## Verification

### Run Tests

```bash
# Run currency app unit tests
go test -v ./app/domain/core/currencyapp/...

# Run formdata FK resolution tests
go test -v ./app/domain/formdata/formdataapp/... -run TestResolveFKByName_Currency
go test -v ./app/domain/formdata/formdataapp/... -run TestMergeFieldDefaults_CurrencyFKResolution

# Run full currency API integration tests
go test -v ./api/cmd/services/ichor/tests/core/currencyapi/...

# Run all tests to ensure no regressions
make test
```

### Build Verification

```bash
# Verify compilation
go build ./api/cmd/services/ichor/...
go build ./app/domain/core/currencyapp/...
```

---

## Key Patterns Followed

1. **QueryByCode signature** matches `QueryByName` from `lineitemfulfillmentstatusapp.go:138-153`
2. **Error handling** uses `errs.Newf(errs.NotFound, ...)` for not-found cases
3. **Unit test pattern** follows existing `fkresolution_test.go` style
4. **Registry integration** follows the same pattern as `sales.line_item_fulfillment_statuses`
