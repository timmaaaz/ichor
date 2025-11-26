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
// Helper Functions
// =============================================================================

func mustMarshal(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal: %v", err))
	}
	return data
}
