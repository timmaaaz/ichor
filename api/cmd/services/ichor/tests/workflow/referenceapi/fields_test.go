package reference_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/workflow/fieldschema"
)

// fieldSchema200 tests successful responses from GET /v1/workflow/entities/{entity}/fields.
func fieldSchema200(sd ReferenceSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "known-enum-entity",
			URL:        "/v1/workflow/entities/inventory.put_away_tasks/fields",
			Token:      sd.Users[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &fieldschema.EntitySchema{},
			ExpResp: &fieldschema.EntitySchema{
				Entity: "inventory.put_away_tasks",
				Fields: []fieldschema.FieldSchema{
					{Name: "status", Type: "enum", Values: []string{"pending", "in_progress", "completed", "cancelled"}},
				},
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp, cmpopts.IgnoreFields(fieldschema.FieldSchema{}, "Description"))
			},
		},
		{
			Name:       "entity-with-no-registered-enums",
			URL:        "/v1/workflow/entities/inventory.inventory_items/fields",
			Token:      sd.Users[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &fieldschema.EntitySchema{},
			ExpResp:    &fieldschema.EntitySchema{Entity: "inventory.inventory_items", Fields: []fieldschema.FieldSchema{}},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

// fieldSchema404 tests 404 responses for unknown entities.
func fieldSchema404(sd ReferenceSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "nonexistent-entity",
			URL:        "/v1/workflow/entities/nonexistent.table/fields",
			Token:      sd.Users[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusNotFound,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "entity not found"),
			CmpFunc: func(got, exp any) string {
				gotErr, ok := got.(*errs.Error)
				if !ok {
					return fmt.Sprintf("expected *errs.Error, got %T", got)
				}
				expErr, ok := exp.(*errs.Error)
				if !ok {
					return fmt.Sprintf("expected *errs.Error exp, got %T", exp)
				}
				if gotErr.Code != expErr.Code {
					return fmt.Sprintf("code: got %v, want %v", gotErr.Code, expErr.Code)
				}
				return ""
			},
		},
	}
}

// fieldSchema401 tests unauthorized responses.
func fieldSchema401(_ ReferenceSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/entities/inventory.put_away_tasks/fields",
			Token:      "",
			Method:     http.MethodGet,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc: func(got, exp any) string {
				return ""
			},
		},
	}
}
