package data_test

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/dataapp"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

func queryByID200(sd apitest.SeedData) []apitest.Table {
	tmp := tablebuilder.StoredConfig{}

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/data/id/%s", sd.SimpleTableConfig.ID.String()),
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &tmp,
			ExpResp:    sd.SimpleTableConfig,
			CmpFunc: func(got any, exp any) string {

				gotResp := got.(*tablebuilder.StoredConfig)
				expResp := exp.(*tablebuilder.StoredConfig)

				// Normalize times for comparison
				expResp.CreatedDate = gotResp.CreatedDate.Round(0).Truncate(time.Microsecond)
				gotResp.CreatedDate = gotResp.CreatedDate.Round(0).Truncate(time.Microsecond)

				expResp.UpdatedDate = gotResp.UpdatedDate.Round(0).Truncate(time.Microsecond)
				gotResp.UpdatedDate = gotResp.UpdatedDate.Round(0).Truncate(time.Microsecond)

				dbtest.NormalizeJSONFields(gotResp, &expResp)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func queryByName200(sd apitest.SeedData) []apitest.Table {
	tmp := tablebuilder.StoredConfig{}

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/data/name/%s", sd.SimpleTableConfig.Name),
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &tmp,
			ExpResp:    sd.SimpleTableConfig,
			CmpFunc: func(got any, exp any) string {

				gotResp := got.(*tablebuilder.StoredConfig)
				expResp := exp.(*tablebuilder.StoredConfig)

				// Normalize times for comparison
				expResp.CreatedDate = gotResp.CreatedDate.Round(0).Truncate(time.Microsecond)
				gotResp.CreatedDate = gotResp.CreatedDate.Round(0).Truncate(time.Microsecond)

				expResp.UpdatedDate = gotResp.UpdatedDate.Round(0).Truncate(time.Microsecond)
				gotResp.UpdatedDate = gotResp.UpdatedDate.Round(0).Truncate(time.Microsecond)

				dbtest.NormalizeJSONFields(gotResp, &expResp)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func queryByUser200(sd apitest.SeedData) []apitest.Table {
	compare := dataapp.TableConfigList{}

	exp := dataapp.TableConfigList{}

	tmp := []tablebuilder.StoredConfig{}
	tmp = append(tmp, *sd.SimpleTableConfig)
	tmp = append(tmp, *sd.ComplexTableConfig)
	tmp = append(tmp, *sd.PageTableConfig)
	tmp = append(tmp, *sd.KPIChartConfig)
	tmp = append(tmp, *sd.BarChartConfig)
	tmp = append(tmp, *sd.PieChartConfig)

	exp.Items = dataapp.ToAppTableConfigs(tmp)

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/data/user/%s", sd.Admins[0].ID.String()),
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &compare,
			ExpResp:    &exp,
			CmpFunc: func(got any, exp any) string {
				gotList := got.(*dataapp.TableConfigList)
				expList := exp.(*dataapp.TableConfigList)

				// Sort both slices by ID first
				sort.Slice(gotList.Items, func(i, j int) bool {
					return gotList.Items[i].ID < gotList.Items[j].ID
				})
				sort.Slice(expList.Items, func(i, j int) bool {
					return expList.Items[i].ID < expList.Items[j].ID
				})

				// Now normalize JSON fields on the sorted slices
				dbtest.NormalizeJSONFields(gotList.Items, expList.Items)

				return cmp.Diff(gotList, expList)
			},
		},
	}
}

// TODO: Once permissions are added on a per-table basis to these configs, we
// need to add tests for violation of this.
