package formdataapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/formdata/formdataapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

const (
	userForm = iota
	assetForm
	multiEntityForm
)

// =============================================================================
// Success Tests (200)
// =============================================================================

func upsertSingleEntityCreate200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "create-user",
			URL:        fmt.Sprintf("/v1/formdata/%s/upsert", sd.Forms[userForm].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &formdataapp.FormDataRequest{
				Operations: map[string]formdataapp.OperationMeta{
					"users": {
						Operation: "create",
						Order:     1,
					},
				},
				Data: map[string]json.RawMessage{
					"users": mustMarshal(map[string]any{
						"username":         "testuser123",
						"first_name":       "Test",
						"last_name":        "User",
						"email":            "testuser@example.com",
						"password":         "SecurePass123!",
						"password_confirm": "SecurePass123!",
						"birthday":         "1990-01-01",
						"roles":            []string{"USER"},
						"system_roles":     []string{"USER"},
						"enabled":          true,
						"requested_by":     sd.Admins[0].ID,
					}),
				},
			},
			GotResp: &formdataapp.FormDataResponse{},
			ExpResp: &formdataapp.FormDataResponse{
				Success: true,
				Results: map[string]interface{}{
					"users": map[string]interface{}{},
				},
			},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*formdataapp.FormDataResponse)
				if !ok {
					return "error occurred"
				}

				if !gotResp.Success {
					return "expected success to be true"
				}

				if _, exists := gotResp.Results["users"]; !exists {
					return "expected users result to exist"
				}

				// Check that a user was created (has an ID)
				userResult, ok := gotResp.Results["users"].(map[string]interface{})
				if !ok {
					return "expected users result to be a map"
				}

				if _, hasID := userResult["id"]; !hasID {
					return "expected user result to have an id field"
				}

				return ""
			},
		},
	}
}

func upsertSingleEntityUpdate200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "update-user",
			URL:        fmt.Sprintf("/v1/formdata/%s/upsert", sd.Forms[userForm].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &formdataapp.FormDataRequest{
				Operations: map[string]formdataapp.OperationMeta{
					"users": {
						Operation: "update",
						Order:     1,
					},
				},
				Data: map[string]json.RawMessage{
					"users": mustMarshal(map[string]any{
						"id":         sd.Users[0].ID,
						"first_name": dbtest.StringPointer("UpdatedFirstName"),
						"last_name":  dbtest.StringPointer("UpdatedLastName"),
					}),
				},
			},
			GotResp: &formdataapp.FormDataResponse{},
			ExpResp: &formdataapp.FormDataResponse{
				Success: true,
				Results: map[string]interface{}{
					"users": map[string]interface{}{},
				},
			},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*formdataapp.FormDataResponse)
				if !ok {
					return "error occurred"
				}

				if !gotResp.Success {
					return "expected success to be true"
				}

				if _, exists := gotResp.Results["users"]; !exists {
					return "expected users result to exist"
				}

				return ""
			},
		},
	}
}

func upsertMultiEntityWithForeignKey200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "create-user-and-asset",
			URL:        fmt.Sprintf("/v1/formdata/%s/upsert", sd.Forms[multiEntityForm].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &formdataapp.FormDataRequest{
				Operations: map[string]formdataapp.OperationMeta{
					"users": {
						Operation: "create",
						Order:     1,
					},
					"assets": {
						Operation: "create",
						Order:     2,
					},
				},
				Data: map[string]json.RawMessage{
					"users": mustMarshal(map[string]any{
						"username":         "newuser456",
						"first_name":       "New",
						"last_name":        "User",
						"email":            "newuser@example.com",
						"password":         "SecurePass123!",
						"password_confirm": "SecurePass123!",
						"birthday":         "1990-01-01",
						"roles":            []string{"USER"},
						"system_roles":     []string{"USER"},
						"enabled":          true,
						"requested_by":     sd.Admins[0].ID,
					}),
					"assets": mustMarshal(map[string]any{
						"asset_condition_id": sd.AssetConditions[0].ID,
						"valid_asset_id":     sd.ValidAssets[0].ID,
						"serial_number":      "SN123456789",
					}),
				},
			},
			GotResp: &formdataapp.FormDataResponse{},
			ExpResp: &formdataapp.FormDataResponse{
				Success: true,
				Results: map[string]interface{}{
					"users":  map[string]interface{}{},
					"assets": map[string]interface{}{},
				},
			},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*formdataapp.FormDataResponse)
				if !ok {
					return "error occurred"
				}

				if !gotResp.Success {
					return "expected success to be true"
				}

				if _, exists := gotResp.Results["users"]; !exists {
					return "expected users result to exist"
				}

				if _, exists := gotResp.Results["assets"]; !exists {
					return "expected assets result to exist"
				}

				// Verify both entities were created with IDs
				assetResult, ok := gotResp.Results["assets"].(map[string]interface{})
				if !ok {
					return "expected assets result to be a map"
				}

				userResult, ok := gotResp.Results["users"].(map[string]interface{})
				if !ok {
					return "expected users result to be a map"
				}

				if _, hasUserID := userResult["id"]; !hasUserID {
					return "expected user result to have an id"
				}

				if _, hasAssetID := assetResult["id"]; !hasAssetID {
					return "expected asset result to have an id"
				}

				return ""
			},
		},
	}
}

// =============================================================================
// Validation Error Tests (400)
// =============================================================================

func upsert400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing-operations",
			URL:        fmt.Sprintf("/v1/formdata/%s/upsert", sd.Forms[userForm].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &formdataapp.FormDataRequest{
				Data: map[string]json.RawMessage{
					"users": mustMarshal(map[string]any{
						"username": "test",
					}),
				},
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"operations\",\"error\":\"operations is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-data",
			URL:        fmt.Sprintf("/v1/formdata/%s/upsert", sd.Forms[userForm].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &formdataapp.FormDataRequest{
				Operations: map[string]formdataapp.OperationMeta{
					"users": {
						Operation: "create",
						Order:     1,
					},
				},
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"data\",\"error\":\"data is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "operations-data-mismatch",
			URL:        fmt.Sprintf("/v1/formdata/%s/upsert", sd.Forms[userForm].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &formdataapp.FormDataRequest{
				Operations: map[string]formdataapp.OperationMeta{
					"users": {
						Operation: "create",
						Order:     1,
					},
				},
				Data: map[string]json.RawMessage{
					"assets": mustMarshal(map[string]any{
						"name": "test",
					}),
				},
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "entity users in operations but missing from data"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "invalid-operation-type",
			URL:        fmt.Sprintf("/v1/formdata/%s/upsert", sd.Forms[userForm].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &formdataapp.FormDataRequest{
				Operations: map[string]formdataapp.OperationMeta{
					"users": {
						Operation: "delete",
						Order:     1,
					},
				},
				Data: map[string]json.RawMessage{
					"users": mustMarshal(map[string]any{
						"id": uuid.NewString(),
					}),
				},
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "entity users has invalid operation: delete"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "entity-not-registered",
			URL:        fmt.Sprintf("/v1/formdata/%s/upsert", sd.Forms[userForm].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &formdataapp.FormDataRequest{
				Operations: map[string]formdataapp.OperationMeta{
					"nonexistent": {
						Operation: "create",
						Order:     1,
					},
				},
				Data: map[string]json.RawMessage{
					"nonexistent": mustMarshal(map[string]any{
						"name": "test",
					}),
				},
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "form validation failed: [{EntityName:nonexistent Operation:create MissingFields:[entity 'nonexistent' not registered in system] AvailableFields:[]}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

// =============================================================================
// Authentication/Authorization Error Tests (401/403)
// =============================================================================

func upsert401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty token",
			URL:        fmt.Sprintf("/v1/formdata/%s/upsert", sd.Forms[userForm].ID),
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad token",
			URL:        fmt.Sprintf("/v1/formdata/%s/upsert", sd.Forms[userForm].ID),
			Token:      sd.Admins[0].Token[:10],
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad sig",
			URL:        fmt.Sprintf("/v1/formdata/%s/upsert", sd.Forms[userForm].ID),
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

// =============================================================================
// Not Found Error Tests (404)
// =============================================================================

func upsert404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "form-not-found",
			URL:        fmt.Sprintf("/v1/formdata/%s/upsert", uuid.New()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusNotFound,
			Input: &formdataapp.FormDataRequest{
				Operations: map[string]formdataapp.OperationMeta{
					"users": {
						Operation: "create",
						Order:     1,
					},
				},
				Data: map[string]json.RawMessage{
					"users": mustMarshal(map[string]any{
						"username": "test",
					}),
				},
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, ""),
			CmpFunc: func(got, exp any) string {
				gotErr, ok := got.(*errs.Error)
				if !ok {
					return "expected error response"
				}

				if gotErr.Code != errs.NotFound {
					return fmt.Sprintf("expected NotFound code, got %v", gotErr.Code)
				}

				return ""
			},
		},
	}
}

// =============================================================================
// Phase 4 Array Operation Tests
// =============================================================================

const (
	salesOrderForm = 3 // Form index in seedData.Forms array
)

// upsertOrderWithSingleLineItem200 tests creating an order with a single line item.
func upsertOrderWithSingleLineItem200(sd apitest.SeedData) []apitest.Table {
	payload := buildOrderWithLineItemsPayload(sd, 1)

	return []apitest.Table{
		{
			Name:       "create-order-single-line-item",
			URL:        fmt.Sprintf("/v1/formdata/%s/upsert", sd.Forms[salesOrderForm].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input:      payload,
			GotResp:    &formdataapp.FormDataResponse{},
			ExpResp:    &formdataapp.FormDataResponse{},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*formdataapp.FormDataResponse)
				if !ok {
					return "error occurred"
				}

				if !gotResp.Success {
					return "expected success to be true"
				}

				// Verify both order and line items were created
				if _, exists := gotResp.Results["orders"]; !exists {
					return "expected orders result to exist"
				}

				if _, exists := gotResp.Results["order_line_items"]; !exists {
					return "expected order_line_items result to exist"
				}

				return ""
			},
		},
	}
}

// upsertOrderWithMultipleLineItems200 tests creating an order with multiple line items.
func upsertOrderWithMultipleLineItems200(sd apitest.SeedData) []apitest.Table {
	payload := buildOrderWithLineItemsPayload(sd, 3)

	return []apitest.Table{
		{
			Name:       "create-order-multiple-line-items",
			URL:        fmt.Sprintf("/v1/formdata/%s/upsert", sd.Forms[salesOrderForm].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input:      payload,
			GotResp:    &formdataapp.FormDataResponse{},
			ExpResp:    &formdataapp.FormDataResponse{},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*formdataapp.FormDataResponse)
				if !ok {
					return "error occurred"
				}

				if !gotResp.Success {
					return "expected success to be true"
				}

				// Verify both order and line items were created
				if _, exists := gotResp.Results["orders"]; !exists {
					return "expected orders result to exist"
				}

				if _, exists := gotResp.Results["order_line_items"]; !exists {
					return "expected order_line_items result to exist"
				}

				// Verify line items is an array with 3 items
				lineItems, ok := gotResp.Results["order_line_items"].([]interface{})
				if !ok {
					return "expected sales.order_line_items to be an array"
				}

				if len(lineItems) != 3 {
					return fmt.Sprintf("expected 3 line items, got %d", len(lineItems))
				}

				return ""
			},
		},
	}
}

// upsertOrderWithArrayValidationError400 tests array validation error handling.
func upsertOrderWithArrayValidationError400(sd apitest.SeedData) []apitest.Table {
	payload := buildOrderWithInvalidLineItem(sd)

	return []apitest.Table{
		{
			Name:       "create-order-invalid-line-item",
			URL:        fmt.Sprintf("/v1/formdata/%s/upsert", sd.Forms[salesOrderForm].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input:      payload,
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotErr, ok := got.(*errs.Error)
				if !ok {
					return "expected error response"
				}

				if gotErr.Code != errs.InvalidArgument {
					return fmt.Sprintf("expected InvalidArgument code, got %v", gotErr.Code)
				}

				// Error message should mention the item index
				if !contains(gotErr.Message, "product_id") {
					return fmt.Sprintf("expected error message to mention missing field, got: %s", gotErr.Message)
				}

				return ""
			},
		},
	}
}

// upsertOrderEmptyLineItems400 tests empty line items array validation.
func upsertOrderEmptyLineItems400(sd apitest.SeedData) []apitest.Table {
	payload := buildOrderWithEmptyLineItems(sd)

	return []apitest.Table{
		{
			Name:       "create-order-empty-line-items",
			URL:        fmt.Sprintf("/v1/formdata/%s/upsert", sd.Forms[salesOrderForm].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input:      payload,
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotErr, ok := got.(*errs.Error)
				if !ok {
					return "expected error response"
				}

				if gotErr.Code != errs.InvalidArgument {
					return fmt.Sprintf("expected InvalidArgument code, got %v", gotErr.Code)
				}

				return ""
			},
		},
	}
}

// upsertBackwardCompatibility200 verifies single-object operations still work.
func upsertBackwardCompatibility200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "backward-compat-single-user",
			URL:        fmt.Sprintf("/v1/formdata/%s/upsert", sd.Forms[userForm].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &formdataapp.FormDataRequest{
				Operations: map[string]formdataapp.OperationMeta{
					"users": {
						Operation: "create",
						Order:     1,
					},
				},
				Data: map[string]json.RawMessage{
					"users": mustMarshal(map[string]any{
						"username":         "backwardcompat",
						"first_name":       "Backward",
						"last_name":        "Compat",
						"email":            "backwardcompat@example.com",
						"password":         "SecurePass123!",
						"password_confirm": "SecurePass123!",
						"birthday":         "1990-01-01",
						"roles":            []string{"USER"},
						"system_roles":     []string{"USER"},
						"enabled":          true,
						"requested_by":     sd.Admins[0].ID,
					}),
				},
			},
			GotResp: &formdataapp.FormDataResponse{},
			ExpResp: &formdataapp.FormDataResponse{},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*formdataapp.FormDataResponse)
				if !ok {
					return "error occurred"
				}

				if !gotResp.Success {
					return "expected success to be true"
				}

				if _, exists := gotResp.Results["users"]; !exists {
					return "expected users result to exist"
				}

				return ""
			},
		},
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func mustMarshal(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal: %v", err))
	}
	return data
}
