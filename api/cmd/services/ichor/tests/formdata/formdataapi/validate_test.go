package formdataapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/formdata/formdataapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/formdataregistry"
)

// =============================================================================
// Success Tests (200)
// =============================================================================

// validate200_ValidForm tests that a complete form with all required fields validates successfully.
func validate200_ValidForm(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "validate-complete-asset-form",
			URL:        fmt.Sprintf("/v1/formdata/%s/validate", sd.Forms[assetForm].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &formdataapp.FormValidationRequest{
				Operations: map[string]formdataregistry.EntityOperation{
					"assets": formdataregistry.OperationCreate,
				},
			},
			GotResp: &formdataapp.FormValidationResult{},
			ExpResp: &formdataapp.FormValidationResult{
				Valid:  true,
				Errors: nil,
			},
			CmpFunc: func(got, exp any) string {
				gotVal, ok := got.(*formdataapp.FormValidationResult)
				if !ok {
					return "got is not *formdataapp.FormValidationResult"
				}

				expVal := exp.(*formdataapp.FormValidationResult)

				return cmp.Diff(expVal, gotVal)
			},
		},
	}
}

// validate200_MultiEntityForm tests validation of a form with multiple entities.
func validate200_MultiEntityForm(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "validate-multi-entity-form",
			URL:        fmt.Sprintf("/v1/formdata/%s/validate", sd.Forms[multiEntityForm].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &formdataapp.FormValidationRequest{
				Operations: map[string]formdataregistry.EntityOperation{
					"users":  formdataregistry.OperationCreate,
					"assets": formdataregistry.OperationCreate,
				},
			},
			GotResp: &formdataapp.FormValidationResult{},
			ExpResp: &formdataapp.FormValidationResult{
				Valid:  true,
				Errors: nil,
			},
			CmpFunc: func(got, exp any) string {
				gotVal, ok := got.(*formdataapp.FormValidationResult)
				if !ok {
					return "got is not *formdataapp.FormValidationResult"
				}

				expVal := exp.(*formdataapp.FormValidationResult)

				return cmp.Diff(expVal, gotVal)
			},
		},
	}
}

// validate200_UnregisteredEntity tests validation with an entity that's not registered in the system.
func validate200_UnregisteredEntity(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "validate-unregistered-entity",
			URL:        fmt.Sprintf("/v1/formdata/%s/validate", sd.Forms[userForm].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &formdataapp.FormValidationRequest{
				Operations: map[string]formdataregistry.EntityOperation{
					"nonexistent": formdataregistry.OperationCreate,
				},
			},
			GotResp: &formdataapp.FormValidationResult{},
			ExpResp: &formdataapp.FormValidationResult{
				Valid: false,
			},
			CmpFunc: func(got, exp any) string {
				gotVal, ok := got.(*formdataapp.FormValidationResult)
				if !ok {
					return "got is not *formdataapp.FormValidationResult"
				}

				// Check that validation failed
				if gotVal.Valid {
					return "expected validation to fail for unregistered entity"
				}

				// Check that there's an error for the unregistered entity
				if len(gotVal.Errors) == 0 {
					return "expected validation error for unregistered entity"
				}

				if gotVal.Errors[0].EntityName != "nonexistent" {
					return "expected error for nonexistent entity"
				}

				return ""
			},
		},
	}
}

// =============================================================================
// Error Tests
// =============================================================================

// validate400 tests validation with an invalid operation type.
func validate400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "validate-invalid-operation",
			URL:        fmt.Sprintf("/v1/formdata/%s/validate", sd.Forms[userForm].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: map[string]interface{}{
				"operations": map[string]string{
					"users": "invalid_operation",
				},
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "entity users has invalid operation: invalid_operation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

// validate401 tests validation without authentication.
func validate401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "validate-401",
			URL:        fmt.Sprintf("/v1/formdata/%s/validate", sd.Forms[userForm].ID),
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &formdataapp.FormValidationRequest{
				Operations: map[string]formdataregistry.EntityOperation{
					"users": formdataregistry.OperationCreate,
				},
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

// validate404 tests validation with a non-existent form.
func validate404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "validate-404-form-not-found",
			URL:        "/v1/formdata/00000000-0000-0000-0000-000000000000/validate",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusNotFound,
			Input: &formdataapp.FormValidationRequest{
				Operations: map[string]formdataregistry.EntityOperation{
					"users": formdataregistry.OperationCreate,
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
