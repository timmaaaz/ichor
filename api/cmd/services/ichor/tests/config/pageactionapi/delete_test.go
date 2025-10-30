package pageaction_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func delete200(sd apitest.SeedData) []apitest.Table {
	// Get last action for deletion
	if len(sd.PageActions) == 0 {
		return []apitest.Table{}
	}

	actionToDelete := sd.PageActions[len(sd.PageActions)-1]

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-actions/" + actionToDelete.ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusNoContent,
			GotResp:    nil,
			ExpResp:    nil,
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}
}

func delete401(sd apitest.SeedData) []apitest.Table {
	if len(sd.PageActions) == 0 {
		return []apitest.Table{}
	}

	actionToDelete := sd.PageActions[0]

	return []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/config/page-actions/" + actionToDelete.ID,
			Token:      sd.Users[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}