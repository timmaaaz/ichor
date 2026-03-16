package supervisorkpiapi_test

import (
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
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &supervisorkpiapp.KPIs{},
			ExpResp: &supervisorkpiapp.KPIs{
				PendingApprovals:    0,
				PendingAdjustments:  len(sd.InventoryAdjustments),
				PendingTransfers:    0,
				OpenInspections:     0,
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
