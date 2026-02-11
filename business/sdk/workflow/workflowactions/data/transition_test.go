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

func Test_TransitionStatusAction(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_TransitionStatusAction")

	sd, err := insertTransitionSeedData(t, db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	sd.TransitionHandler = data.NewTransitionStatusHandler(log, db.DB)

	unitest.Run(t, transitionStatusTests(sd), "transitionStatusAction")
}

// =============================================================================

type transitionSeedData struct {
	unitest.SeedData
	Customer          customersbus.Customers
	TransitionHandler *data.TransitionStatusHandler
}

func insertTransitionSeedData(t *testing.T, busDomain dbtest.BusDomain) (transitionSeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return transitionSeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	adminUser := admins[0]

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return transitionSeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, 1, regionIDs, busDomain.City)
	if err != nil {
		return transitionSeedData{}, fmt.Errorf("seeding cities : %w", err)
	}

	cityIDs := make([]uuid.UUID, 0, len(cities))
	for _, c := range cities {
		cityIDs = append(cityIDs, c.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 1, cityIDs, busDomain.Street)
	if err != nil {
		return transitionSeedData{}, fmt.Errorf("seeding streets : %w", err)
	}

	streetIDs := make([]uuid.UUID, 0, len(streets))
	for _, s := range streets {
		streetIDs = append(streetIDs, s.ID)
	}

	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return transitionSeedData{}, fmt.Errorf("querying timezones : %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 1, streetIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return transitionSeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}

	contactInfoIDs := make([]uuid.UUID, 0, len(contactInfos))
	for _, ci := range contactInfos {
		contactInfoIDs = append(contactInfoIDs, ci.ID)
	}

	customers, err := customersbus.TestSeedCustomers(ctx, 1, streetIDs, contactInfoIDs, uuid.UUIDs{adminUser.ID}, busDomain.Customers)
	if err != nil {
		return transitionSeedData{}, fmt.Errorf("seeding customers : %w", err)
	}

	return transitionSeedData{
		SeedData: unitest.SeedData{
			Users:     []unitest.User{{User: adminUser}},
			Admins:    []unitest.User{{User: adminUser}},
			Customers: customers,
		},
		Customer: customers[0],
	}, nil
}

// =============================================================================

func transitionStatusTests(sd transitionSeedData) []unitest.Table {
	return []unitest.Table{
		transitionValid(sd),
		transitionInvalidRejected(sd),
		transitionWithTemplate(sd),
		transitionEntityNotFound(sd),
		transitionValidateEmptyValidFrom(sd),
		transitionValidateInvalidTable(sd),
		transitionEntityModifier(sd),
	}
}

func transitionValid(sd transitionSeedData) unitest.Table {
	return unitest.Table{
		Name: "transition_valid",
		ExpResp: map[string]any{
			"transitioned": true,
		},
		ExcFunc: func(ctx context.Context) any {
			// Customer name is set by TestSeedCustomers - we'll transition the name field
			currentName := sd.Customer.Name

			config := json.RawMessage(fmt.Sprintf(`{
				"target_entity": "sales.customers",
				"target_id": "%s",
				"status_field": "name",
				"to_status": "Premium Customer",
				"valid_from_statuses": ["%s"]
			}`, sd.Customer.ID, currentName))

			execContext := workflow.ActionExecutionContext{
				EntityID:    sd.Customer.ID,
				EntityName:  "sales.customers",
				EventType:   "on_update",
				UserID:      sd.Admins[0].ID,
				ExecutionID: uuid.New(),
				Timestamp:   time.Now().UTC(),
			}

			result, err := sd.TransitionHandler.Execute(ctx, config, execContext)
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

			if gotResp["transitioned"] != true {
				return fmt.Sprintf("expected transitioned=true, got %v", gotResp["transitioned"])
			}

			return ""
		},
	}
}

func transitionInvalidRejected(sd transitionSeedData) unitest.Table {
	return unitest.Table{
		Name:    "transition_invalid_rejected",
		ExpResp: "invalid_transition",
		ExcFunc: func(ctx context.Context) any {
			// Customer name is now "Premium Customer" from previous test,
			// but we only allow transition from "NonExistentStatus".
			// Handler returns output="invalid_transition" (no error) for routing.
			config := json.RawMessage(fmt.Sprintf(`{
				"target_entity": "sales.customers",
				"target_id": "%s",
				"status_field": "name",
				"to_status": "VIP",
				"valid_from_statuses": ["NonExistentStatus"]
			}`, sd.Customer.ID))

			execContext := workflow.ActionExecutionContext{
				EntityID:    sd.Customer.ID,
				EntityName:  "sales.customers",
				EventType:   "on_update",
				UserID:      sd.Admins[0].ID,
				ExecutionID: uuid.New(),
				Timestamp:   time.Now().UTC(),
			}

			result, err := sd.TransitionHandler.Execute(ctx, config, execContext)
			if err != nil {
				return fmt.Errorf("unexpected error: %w", err)
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map[string]any, got %T", result)
			}

			return resultMap["output"].(string)
		},
		CmpFunc: func(got any, exp any) string {
			if got != exp {
				return fmt.Sprintf("expected transition error: %v", got)
			}
			return ""
		},
	}
}

func transitionWithTemplate(sd transitionSeedData) unitest.Table {
	return unitest.Table{
		Name: "transition_with_template",
		ExpResp: map[string]any{
			"transitioned": true,
		},
		ExcFunc: func(ctx context.Context) any {
			// Use {{entity_id}} template for target_id
			config := json.RawMessage(`{
				"target_entity": "sales.customers",
				"target_id": "{{entity_id}}",
				"status_field": "name",
				"to_status": "Template Transitioned",
				"valid_from_statuses": ["Premium Customer"]
			}`)

			execContext := workflow.ActionExecutionContext{
				EntityID:    sd.Customer.ID,
				EntityName:  "sales.customers",
				EventType:   "on_update",
				UserID:      sd.Admins[0].ID,
				ExecutionID: uuid.New(),
				Timestamp:   time.Now().UTC(),
			}

			result, err := sd.TransitionHandler.Execute(ctx, config, execContext)
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

			if gotResp["transitioned"] != true {
				return fmt.Sprintf("expected transitioned=true, got %v", gotResp["transitioned"])
			}

			return ""
		},
	}
}

func transitionEntityNotFound(sd transitionSeedData) unitest.Table {
	return unitest.Table{
		Name:    "transition_entity_not_found",
		ExpResp: "error",
		ExcFunc: func(ctx context.Context) any {
			nonExistent := uuid.New()
			config := json.RawMessage(fmt.Sprintf(`{
				"target_entity": "sales.customers",
				"target_id": "%s",
				"status_field": "name",
				"to_status": "VIP",
				"valid_from_statuses": ["Standard"]
			}`, nonExistent))

			execContext := workflow.ActionExecutionContext{
				EntityID:    nonExistent,
				EntityName:  "sales.customers",
				EventType:   "on_update",
				UserID:      sd.Admins[0].ID,
				ExecutionID: uuid.New(),
				Timestamp:   time.Now().UTC(),
			}

			_, err := sd.TransitionHandler.Execute(ctx, config, execContext)
			if err == nil {
				return "expected error for non-existent entity, got nil"
			}

			return "error"
		},
		CmpFunc: func(got any, exp any) string {
			if got != "error" {
				return fmt.Sprintf("expected error: %v", got)
			}
			return ""
		},
	}
}

func transitionValidateEmptyValidFrom(sd transitionSeedData) unitest.Table {
	return unitest.Table{
		Name:    "validate_empty_valid_from",
		ExpResp: "error",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"target_entity": "sales.customers",
				"target_id": "123",
				"status_field": "name",
				"to_status": "VIP",
				"valid_from_statuses": []
			}`)

			err := sd.TransitionHandler.Validate(config)
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

func transitionValidateInvalidTable(sd transitionSeedData) unitest.Table {
	return unitest.Table{
		Name:    "validate_invalid_table",
		ExpResp: "error",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"target_entity": "invalid.table",
				"target_id": "123",
				"status_field": "name",
				"to_status": "VIP",
				"valid_from_statuses": ["Standard"]
			}`)

			err := sd.TransitionHandler.Validate(config)
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

func transitionEntityModifier(sd transitionSeedData) unitest.Table {
	return unitest.Table{
		Name: "entity_modifier_returns_on_update",
		ExpResp: map[string]any{
			"event_type": "on_update",
		},
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"target_entity": "sales.customers",
				"target_id": "123",
				"status_field": "name",
				"to_status": "VIP",
				"valid_from_statuses": ["Standard"]
			}`)

			mods := sd.TransitionHandler.GetEntityModifications(config)
			if len(mods) == 0 {
				return fmt.Errorf("expected modifications, got none")
			}

			return map[string]any{
				"event_type": mods[0].EventType,
				"fields":     mods[0].Fields,
			}
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, ok := got.(map[string]any)
			if !ok {
				return fmt.Sprintf("error occurred or wrong type: %v", got)
			}

			if gotResp["event_type"] != "on_update" {
				return fmt.Sprintf("event_type mismatch: got %v, want on_update", gotResp["event_type"])
			}

			fields, ok := gotResp["fields"].([]string)
			if !ok || len(fields) == 0 {
				return "expected fields to contain status_field"
			}

			if fields[0] != "name" {
				return fmt.Sprintf("field mismatch: got %v, want name", fields[0])
			}

			return ""
		},
	}
}
