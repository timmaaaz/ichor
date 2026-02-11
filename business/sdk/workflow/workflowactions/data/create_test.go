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

func Test_CreateEntityAction(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_CreateEntityAction")

	sd, err := insertCreateSeedData(t, db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	sd.CreateHandler = data.NewCreateEntityHandler(log, db.DB)

	unitest.Run(t, createEntityActionTests(sd), "createEntityAction")
}

// =============================================================================

type createSeedData struct {
	unitest.SeedData
	Customer       customersbus.Customers
	CustomersBus   *customersbus.Business
	ContactInfoIDs []uuid.UUID
	StreetIDs      []uuid.UUID
	CreateHandler  *data.CreateEntityHandler
}

func insertCreateSeedData(t *testing.T, busDomain dbtest.BusDomain) (createSeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return createSeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	adminUser := admins[0]

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return createSeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, 1, regionIDs, busDomain.City)
	if err != nil {
		return createSeedData{}, fmt.Errorf("seeding cities : %w", err)
	}

	cityIDs := make([]uuid.UUID, 0, len(cities))
	for _, c := range cities {
		cityIDs = append(cityIDs, c.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 1, cityIDs, busDomain.Street)
	if err != nil {
		return createSeedData{}, fmt.Errorf("seeding streets : %w", err)
	}

	streetIDs := make([]uuid.UUID, 0, len(streets))
	for _, s := range streets {
		streetIDs = append(streetIDs, s.ID)
	}

	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return createSeedData{}, fmt.Errorf("querying timezones : %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 1, streetIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return createSeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}

	contactInfoIDs := make([]uuid.UUID, 0, len(contactInfos))
	for _, ci := range contactInfos {
		contactInfoIDs = append(contactInfoIDs, ci.ID)
	}

	customers, err := customersbus.TestSeedCustomers(ctx, 1, streetIDs, contactInfoIDs, uuid.UUIDs{adminUser.ID}, busDomain.Customers)
	if err != nil {
		return createSeedData{}, fmt.Errorf("seeding customers : %w", err)
	}

	return createSeedData{
		SeedData: unitest.SeedData{
			Users:  []unitest.User{{User: adminUser}},
			Admins: []unitest.User{{User: adminUser}},
		},
		Customer:       customers[0],
		CustomersBus:   busDomain.Customers,
		ContactInfoIDs: contactInfoIDs,
		StreetIDs:      streetIDs,
	}, nil
}

// =============================================================================

func createEntityActionTests(sd createSeedData) []unitest.Table {
	return []unitest.Table{
		createBasicEntity(sd),
		createWithTemplates(sd),
		createAutoGeneratesID(sd),
		createValidateInvalidTable(sd),
		createValidateEmptyFields(sd),
		createEntityModifier(sd),
	}
}

func createBasicEntity(sd createSeedData) unitest.Table {
	return unitest.Table{
		Name: "create_basic_entity",
		ExpResp: map[string]any{
			"status": "success",
		},
		ExcFunc: func(ctx context.Context) any {
			newID := uuid.New()
			now := time.Now().UTC()
			config := json.RawMessage(fmt.Sprintf(`{
				"target_entity": "sales.customers",
				"fields": {
					"id": "%s",
					"name": "Test Customer Created",
					"contact_id": "%s",
					"delivery_address_id": "%s",
					"notes": "",
					"created_by": "%s",
					"updated_by": "%s",
					"created_date": "%s",
					"updated_date": "%s"
				}
			}`, newID, sd.ContactInfoIDs[0], sd.StreetIDs[0], sd.Admins[0].ID, sd.Admins[0].ID, now.Format(time.RFC3339), now.Format(time.RFC3339)))

			execContext := workflow.ActionExecutionContext{
				EntityID:    sd.Customer.ID,
				EntityName:  "sales.customers",
				EventType:   "on_create",
				UserID:      sd.Admins[0].ID,
				ExecutionID: uuid.New(),
				Timestamp:   now,
			}

			result, err := sd.CreateHandler.Execute(ctx, config, execContext)
			if err != nil {
				return err
			}

			// Verify it exists
			_, err = sd.CustomersBus.QueryByID(ctx, newID)
			if err != nil {
				return fmt.Errorf("created customer not found: %w", err)
			}

			return result
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, ok := got.(map[string]any)
			if !ok {
				return fmt.Sprintf("error occurred or wrong type: %v", got)
			}

			if gotResp["status"] != "success" {
				return fmt.Sprintf("status mismatch: got %v, want success", gotResp["status"])
			}

			return ""
		},
	}
}

func createWithTemplates(sd createSeedData) unitest.Table {
	return unitest.Table{
		Name: "create_with_templates",
		ExpResp: map[string]any{
			"status": "success",
		},
		ExcFunc: func(ctx context.Context) any {
			newID := uuid.New()
			now := time.Now().UTC()
			config := json.RawMessage(fmt.Sprintf(`{
				"target_entity": "sales.customers",
				"fields": {
					"id": "%s",
					"name": "Customer by {{user_id}}",
					"contact_id": "%s",
					"delivery_address_id": "%s",
					"created_by": "{{user_id}}",
					"updated_by": "{{user_id}}",
					"created_date": "%s",
					"updated_date": "%s"
				}
			}`, newID, sd.ContactInfoIDs[0], sd.StreetIDs[0], now.Format(time.RFC3339), now.Format(time.RFC3339)))

			execContext := workflow.ActionExecutionContext{
				EntityID:    sd.Customer.ID,
				EntityName:  "sales.customers",
				EventType:   "on_create",
				UserID:      sd.Admins[0].ID,
				ExecutionID: uuid.New(),
				Timestamp:   now,
			}

			result, err := sd.CreateHandler.Execute(ctx, config, execContext)
			if err != nil {
				return err
			}

			return result
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, ok := got.(map[string]any)
			if !ok {
				return fmt.Sprintf("error occurred or wrong type: %v", got)
			}

			if gotResp["status"] != "success" {
				return fmt.Sprintf("status mismatch: got %v, want success", gotResp["status"])
			}

			return ""
		},
	}
}

func createAutoGeneratesID(sd createSeedData) unitest.Table {
	return unitest.Table{
		Name: "create_auto_generates_id",
		ExpResp: map[string]any{
			"status": "success",
		},
		ExcFunc: func(ctx context.Context) any {
			now := time.Now().UTC()
			// No "id" field — handler should auto-generate
			config := json.RawMessage(fmt.Sprintf(`{
				"target_entity": "sales.customers",
				"fields": {
					"name": "Auto ID Customer",
					"contact_id": "%s",
					"delivery_address_id": "%s",
					"created_by": "%s",
					"updated_by": "%s",
					"created_date": "%s",
					"updated_date": "%s"
				}
			}`, sd.ContactInfoIDs[0], sd.StreetIDs[0], sd.Admins[0].ID, sd.Admins[0].ID, now.Format(time.RFC3339), now.Format(time.RFC3339)))

			execContext := workflow.ActionExecutionContext{
				EntityID:    sd.Customer.ID,
				EntityName:  "sales.customers",
				EventType:   "on_create",
				UserID:      sd.Admins[0].ID,
				ExecutionID: uuid.New(),
				Timestamp:   now,
			}

			result, err := sd.CreateHandler.Execute(ctx, config, execContext)
			if err != nil {
				return err
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("unexpected result type: %T", result)
			}

			// Verify created_id is a valid UUID
			createdID, ok := resultMap["created_id"].(uuid.UUID)
			if !ok {
				return fmt.Errorf("created_id is not uuid.UUID: %T", resultMap["created_id"])
			}

			if createdID == uuid.Nil {
				return fmt.Errorf("created_id is nil UUID")
			}

			return result
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, ok := got.(map[string]any)
			if !ok {
				return fmt.Sprintf("error occurred or wrong type: %v", got)
			}

			if gotResp["status"] != "success" {
				return fmt.Sprintf("status mismatch: got %v, want success", gotResp["status"])
			}

			return ""
		},
	}
}

func createValidateInvalidTable(sd createSeedData) unitest.Table {
	return unitest.Table{
		Name:    "validate_invalid_table",
		ExpResp: "error",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"target_entity": "invalid.table",
				"fields": {"name": "test"}
			}`)

			err := sd.CreateHandler.Validate(config)
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

func createValidateEmptyFields(sd createSeedData) unitest.Table {
	return unitest.Table{
		Name:    "validate_empty_fields",
		ExpResp: "error",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"target_entity": "sales.customers",
				"fields": {}
			}`)

			err := sd.CreateHandler.Validate(config)
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

func createEntityModifier(sd createSeedData) unitest.Table {
	return unitest.Table{
		Name: "entity_modifier_returns_on_create",
		ExpResp: map[string]any{
			"event_type": "on_create",
		},
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"target_entity": "sales.customers",
				"fields": {"name": "test"}
			}`)

			mods := sd.CreateHandler.GetEntityModifications(config)
			if len(mods) == 0 {
				return fmt.Errorf("expected modifications, got none")
			}

			return map[string]any{
				"event_type":  mods[0].EventType,
				"entity_name": mods[0].EntityName,
			}
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, ok := got.(map[string]any)
			if !ok {
				return fmt.Sprintf("error occurred or wrong type: %v", got)
			}

			if gotResp["event_type"] != "on_create" {
				return fmt.Sprintf("event_type mismatch: got %v, want on_create", gotResp["event_type"])
			}

			return ""
		},
	}
}
