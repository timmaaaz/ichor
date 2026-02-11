package data_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/data"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

func Test_LookupEntityAction(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_LookupEntityAction")

	sd, err := insertLookupSeedData(t, db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	sd.LookupHandler = data.NewLookupEntityHandler(log, db.DB)

	unitest.Run(t, lookupEntityActionTests(sd), "lookupEntityAction")
}

// =============================================================================

type lookupSeedData struct {
	unitest.SeedData
	Customer      customersbus.Customers
	LookupHandler *data.LookupEntityHandler
}

func insertLookupSeedData(t *testing.T, busDomain dbtest.BusDomain) (lookupSeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return lookupSeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	adminUser := admins[0]

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return lookupSeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, 1, regionIDs, busDomain.City)
	if err != nil {
		return lookupSeedData{}, fmt.Errorf("seeding cities : %w", err)
	}

	cityIDs := make([]uuid.UUID, 0, len(cities))
	for _, c := range cities {
		cityIDs = append(cityIDs, c.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 1, cityIDs, busDomain.Street)
	if err != nil {
		return lookupSeedData{}, fmt.Errorf("seeding streets : %w", err)
	}

	streetIDs := make([]uuid.UUID, 0, len(streets))
	for _, s := range streets {
		streetIDs = append(streetIDs, s.ID)
	}

	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return lookupSeedData{}, fmt.Errorf("querying timezones : %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 1, streetIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return lookupSeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}

	contactInfoIDs := make([]uuid.UUID, 0, len(contactInfos))
	for _, ci := range contactInfos {
		contactInfoIDs = append(contactInfoIDs, ci.ID)
	}

	customers, err := customersbus.TestSeedCustomers(ctx, 2, streetIDs, contactInfoIDs, uuid.UUIDs{adminUser.ID}, busDomain.Customers)
	if err != nil {
		return lookupSeedData{}, fmt.Errorf("seeding customers : %w", err)
	}

	return lookupSeedData{
		SeedData: unitest.SeedData{
			Users:  []unitest.User{{User: adminUser}},
			Admins: []unitest.User{{User: adminUser}},
		},
		Customer: customers[0],
	}, nil
}

// =============================================================================

func lookupEntityActionTests(sd lookupSeedData) []unitest.Table {
	return []unitest.Table{
		lookupSingleRow(sd),
		lookupWithTemplate(sd),
		lookupNoResults(sd),
		validateInvalidTable(sd),
		validateEmptyFilter(sd),
		validateEmptyFields(sd),
	}
}

func lookupSingleRow(sd lookupSeedData) unitest.Table {
	return unitest.Table{
		Name: "lookup_single_row",
		ExpResp: map[string]any{
			"found": true,
		},
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(fmt.Sprintf(`{
				"entity": "sales.customers",
				"filter": {"id": "%s"},
				"fields": ["id", "name"],
				"output_key": "customer"
			}`, sd.Customer.ID))

			execContext := workflow.ActionExecutionContext{
				EntityID:    sd.Customer.ID,
				EntityName:  "sales.customers",
				EventType:   "on_update",
				UserID:      sd.Admins[0].ID,
				ExecutionID: uuid.New(),
				Timestamp:   time.Now().UTC(),
			}

			result, err := sd.LookupHandler.Execute(ctx, config, execContext)
			if err != nil {
				return err
			}

			return result
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, ok := got.(map[string]any)
			if !ok {
				return fmt.Sprintf("error occurred or wrong type returned: %v", got)
			}

			if gotResp["found"] != true {
				return fmt.Sprintf("expected found=true, got %v", gotResp["found"])
			}

			if gotResp["customer"] == nil {
				return "expected customer data, got nil"
			}

			return ""
		},
	}
}

func lookupWithTemplate(sd lookupSeedData) unitest.Table {
	return unitest.Table{
		Name: "lookup_with_template",
		ExpResp: map[string]any{
			"found": true,
		},
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"entity": "sales.customers",
				"filter": {"id": "{{entity_id}}"},
				"fields": ["id", "name"],
				"output_key": "customer"
			}`)

			execContext := workflow.ActionExecutionContext{
				EntityID:    sd.Customer.ID,
				EntityName:  "sales.customers",
				EventType:   "on_update",
				UserID:      sd.Admins[0].ID,
				ExecutionID: uuid.New(),
				Timestamp:   time.Now().UTC(),
			}

			result, err := sd.LookupHandler.Execute(ctx, config, execContext)
			if err != nil {
				return err
			}

			return result
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, ok := got.(map[string]any)
			if !ok {
				return fmt.Sprintf("error occurred or wrong type returned: %v", got)
			}

			if gotResp["found"] != true {
				return fmt.Sprintf("expected found=true, got %v", gotResp["found"])
			}

			return ""
		},
	}
}

func lookupNoResults(sd lookupSeedData) unitest.Table {
	return unitest.Table{
		Name: "lookup_no_results",
		ExpResp: map[string]any{
			"found": false,
		},
		ExcFunc: func(ctx context.Context) any {
			nonExistentID := uuid.New()
			config := json.RawMessage(fmt.Sprintf(`{
				"entity": "sales.customers",
				"filter": {"id": "%s"},
				"fields": ["id", "name"],
				"output_key": "customer"
			}`, nonExistentID))

			execContext := workflow.ActionExecutionContext{
				EntityID:    nonExistentID,
				EntityName:  "sales.customers",
				EventType:   "on_update",
				UserID:      sd.Admins[0].ID,
				ExecutionID: uuid.New(),
				Timestamp:   time.Now().UTC(),
			}

			result, err := sd.LookupHandler.Execute(ctx, config, execContext)
			if err != nil {
				return err
			}

			return result
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, ok := got.(map[string]any)
			if !ok {
				return fmt.Sprintf("error occurred or wrong type returned: %v", got)
			}

			if gotResp["found"] != false {
				return fmt.Sprintf("expected found=false, got %v", gotResp["found"])
			}

			return ""
		},
	}
}

func validateInvalidTable(sd lookupSeedData) unitest.Table {
	return unitest.Table{
		Name:    "validate_invalid_table",
		ExpResp: "error",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"entity": "invalid.table",
				"filter": {"id": "123"},
				"fields": ["id"],
				"output_key": "result"
			}`)

			err := sd.LookupHandler.Validate(config)
			if err == nil {
				return "expected validation error, got nil"
			}

			return "error"
		},
		CmpFunc: func(got any, exp any) string {
			if got != "error" {
				return fmt.Sprintf("expected validation error: %v", got)
			}
			return ""
		},
	}
}

func validateEmptyFilter(sd lookupSeedData) unitest.Table {
	return unitest.Table{
		Name:    "validate_empty_filter",
		ExpResp: "error",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"entity": "sales.customers",
				"filter": {},
				"fields": ["id"],
				"output_key": "result"
			}`)

			err := sd.LookupHandler.Validate(config)
			if err == nil {
				return "expected validation error, got nil"
			}

			return "error"
		},
		CmpFunc: func(got any, exp any) string {
			if got != "error" {
				return fmt.Sprintf("expected validation error: %v", got)
			}
			return ""
		},
	}
}

func validateEmptyFields(sd lookupSeedData) unitest.Table {
	return unitest.Table{
		Name:    "validate_empty_fields",
		ExpResp: "error",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"entity": "sales.customers",
				"filter": {"id": "123"},
				"fields": [],
				"output_key": "result"
			}`)

			err := sd.LookupHandler.Validate(config)
			if err == nil {
				return "expected validation error, got nil"
			}

			return "error"
		},
		CmpFunc: func(got any, exp any) string {
			if got != "error" {
				return fmt.Sprintf("expected validation error: %v", got)
			}
			return ""
		},
	}
}
