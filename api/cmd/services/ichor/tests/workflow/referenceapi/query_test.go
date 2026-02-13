package reference_test

import (
	"fmt"
	"net/http"

	"github.com/timmaaaz/ichor/api/domain/http/workflow/referenceapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

// expectedActionTypes are the action types that should always be available.
// Defined as a package constant to avoid typos and improve maintainability.
var expectedActionTypes = []string{
	"allocate_inventory",
	"check_inventory",
	"check_reorder_point",
	"commit_allocation",
	"create_alert",
	"create_entity",
	"delay",
	"evaluate_condition",
	"log_audit_entry",
	"lookup_entity",
	"release_reservation",
	"reserve_inventory",
	"seek_approval",
	"send_email",
	"send_notification",
	"transition_status",
	"update_field",
}

// =============================================================================
// Trigger Types Tests

func queryTriggerTypes200(sd ReferenceSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/workflow/trigger-types",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &[]referenceapi.TriggerType{},
			ExpResp:    &[]referenceapi.TriggerType{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*[]referenceapi.TriggerType)
				if !exists {
					return "error getting trigger types response"
				}

				// Should have at least the seeded trigger types
				if len(*gotResp) < len(sd.TriggerTypes) {
					return "expected at least seeded trigger types"
				}

				return ""
			},
		},
	}

	return table
}

func queryTriggerTypes401(sd ReferenceSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/trigger-types",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}

	return table
}

// =============================================================================
// Entity Types Tests

func queryEntityTypes200(sd ReferenceSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/workflow/entity-types",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &[]referenceapi.EntityType{},
			ExpResp:    &[]referenceapi.EntityType{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*[]referenceapi.EntityType)
				if !exists {
					return "error getting entity types response"
				}

				// Should have entity types from migrations
				if len(*gotResp) == 0 {
					return "expected at least some entity types"
				}

				return ""
			},
		},
	}

	return table
}

func queryEntityTypes401(sd ReferenceSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/entity-types",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}

	return table
}

// =============================================================================
// Entities Tests

func queryEntities200(sd ReferenceSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/workflow/entities",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &[]referenceapi.Entity{},
			ExpResp:    &[]referenceapi.Entity{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*[]referenceapi.Entity)
				if !exists {
					return "error getting entities response"
				}

				// Should have entities from migrations
				if len(*gotResp) == 0 {
					return "expected at least some entities"
				}

				return ""
			},
		},
	}

	return table
}

func queryEntities401(sd ReferenceSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/entities",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}

	return table
}

func queryEntitiesWithFilter200(sd ReferenceSeedData) []apitest.Table {
	// Only run this if we have entity types
	if len(sd.EntityTypes) == 0 {
		return nil
	}

	entityTypeID := sd.EntityTypes[0].ID

	table := []apitest.Table{
		{
			Name:       "filter-by-entity-type",
			URL:        "/v1/workflow/entities?entity_type_id=" + entityTypeID.String(),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &[]referenceapi.Entity{},
			ExpResp:    &[]referenceapi.Entity{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*[]referenceapi.Entity)
				if !exists {
					return "error getting entities response"
				}

				// All returned entities should have the filtered entity type
				for _, e := range *gotResp {
					if e.EntityTypeID != entityTypeID {
						return "entity has wrong entity type id"
					}
				}

				return ""
			},
		},
	}

	return table
}

// =============================================================================
// Action Types Tests

func queryActionTypes200(sd ReferenceSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/workflow/action-types",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &[]referenceapi.ActionTypeInfo{},
			ExpResp:    &[]referenceapi.ActionTypeInfo{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*[]referenceapi.ActionTypeInfo)
				if !exists {
					return "error getting action types response"
				}

				// Should have all 17 predefined action types
				expectedTypes := make(map[string]bool)
				for _, t := range expectedActionTypes {
					expectedTypes[t] = true
				}

				for _, at := range *gotResp {
					delete(expectedTypes, at.Type)

					// Every type must have output ports
					if len(at.OutputPorts) == 0 {
						return fmt.Sprintf("action type %q has no output ports", at.Type)
					}

					// At least one port must be the default
					hasDefault := false
					for _, port := range at.OutputPorts {
						if port.IsDefault {
							hasDefault = true
							break
						}
					}
					if !hasDefault {
						return fmt.Sprintf("action type %q has no default output port", at.Type)
					}

					// Must have a config schema
					if len(at.ConfigSchema) == 0 {
						return fmt.Sprintf("action type %q has no config schema", at.Type)
					}

					// Category must be valid
					validCategories := map[string]bool{
						"communication": true, "inventory": true, "control": true,
						"data": true, "approval": true,
					}
					if !validCategories[at.Category] {
						return fmt.Sprintf("action type %q has invalid category %q", at.Type, at.Category)
					}
				}

				if len(expectedTypes) > 0 {
					missing := make([]string, 0, len(expectedTypes))
					for t := range expectedTypes {
						missing = append(missing, t)
					}
					return fmt.Sprintf("missing expected action types: %v", missing)
				}

				return ""
			},
		},
	}

	return table
}

func queryActionTypes401(sd ReferenceSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/action-types",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}

	return table
}

// =============================================================================
// Action Type Schema Tests

func queryActionTypeSchema200(sd ReferenceSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "create_alert-schema",
			URL:        "/v1/workflow/action-types/create_alert/schema",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &referenceapi.ActionTypeInfo{},
			ExpResp:    &referenceapi.ActionTypeInfo{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*referenceapi.ActionTypeInfo)
				if !exists {
					return "error getting action type schema response"
				}

				if gotResp.Type != "create_alert" {
					return "wrong action type returned"
				}

				if gotResp.Name == "" {
					return "action type name is empty"
				}

				if len(gotResp.ConfigSchema) == 0 {
					return "config schema is empty"
				}

				return ""
			},
		},
		{
			Name:       "allocate_inventory-schema",
			URL:        "/v1/workflow/action-types/allocate_inventory/schema",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &referenceapi.ActionTypeInfo{},
			ExpResp:    &referenceapi.ActionTypeInfo{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*referenceapi.ActionTypeInfo)
				if !exists {
					return "error getting action type schema response"
				}

				if gotResp.Type != "allocate_inventory" {
					return "wrong action type returned"
				}

				if !gotResp.SupportsManual {
					return "allocate_inventory should support manual execution"
				}

				return ""
			},
		},
	}

	return table
}

func queryActionTypeSchema404(sd ReferenceSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "nonexistent",
			URL:        "/v1/workflow/action-types/nonexistent_action/schema",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodGet,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}

	return table
}

func queryActionTypeSchema401(sd ReferenceSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/action-types/create_alert/schema",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}

	return table
}
