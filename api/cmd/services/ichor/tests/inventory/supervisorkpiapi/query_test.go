package supervisorkpiapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/supervisorkpiapp"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/supervisor/kpis",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &supervisorkpiapp.KPIs{},
			ExpResp: &supervisorkpiapp.KPIs{
				PendingApprovals:    0,
				PendingAdjustments:  len(sd.InventoryAdjustments),
				PendingTransfers:    0,
				PendingInspections:  0,
				PendingPutAwayTasks: len(sd.PutAwayTasks),
				ActiveAlerts:        0,
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func query401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "unauthenticated",
			URL:        "/v1/inventory/supervisor/kpis",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &supervisorkpiapp.KPIs{},
			ExpResp:    &supervisorkpiapp.KPIs{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
