package data_test

import (
	"net/http"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/dataapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

func queryAll200(sd apitest.SeedData) []apitest.Table {
	expected := dataapp.ToAppTableConfigList([]tablebuilder.StoredConfig{
		*sd.SimpleTableConfig,
		*sd.PageTableConfig,
		*sd.ComplexTableConfig,
	})

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/data/configs/all",
			Method:     http.MethodGet,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			GotResp:    &dataapp.TableConfigList{},
			ExpResp:    &expected,
			CmpFunc: func(got, exp any) string {
				gotList := got.(*dataapp.TableConfigList)
				expList := exp.(*dataapp.TableConfigList)

				// Sort both slices by ID for consistent comparison
				sort.Slice(gotList.Items, func(i, j int) bool {
					return gotList.Items[i].ID < gotList.Items[j].ID
				})
				sort.Slice(expList.Items, func(i, j int) bool {
					return expList.Items[i].ID < expList.Items[j].ID
				})

				// Normalize JSON fields
				dbtest.NormalizeJSONFields(gotList.Items, expList.Items)

				return cmp.Diff(gotList, expList)
			},
		},
	}

	return table
}

func queryAll401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/data/configs/all",
			Method:     http.MethodGet,
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "expected authorization header format: Bearer <token>"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
