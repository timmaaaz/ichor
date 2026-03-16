package inspectionapi_test

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/inspectionapp"
)

func fail200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "fail-with-quarantine",
			Token:      sd.Admins[0].Token,
			URL:        "/v1/inventory/quality-inspections/" + sd.Inspections[0].InspectionID + "/fail",
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &inspectionapp.FailInspection{
				Notes:         "Contamination detected",
				QuarantineLot: true,
			},
			GotResp: &inspectionapp.FailInspectionResult{},
			CmpFunc: func(got any, exp any) string {
				result := got.(*inspectionapp.FailInspectionResult)
				if result.Inspection.Status != "failed" {
					return "expected inspection status 'failed', got '" + result.Inspection.Status + "'"
				}
				if result.LotStatus != "quarantined" {
					return "expected lot_status 'quarantined', got '" + result.LotStatus + "'"
				}
				return ""
			},
		},
	}
}

func fail200NoQuarantine(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "fail-without-quarantine",
			Token:      sd.Admins[0].Token,
			URL:        "/v1/inventory/quality-inspections/" + sd.Inspections[1].InspectionID + "/fail",
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &inspectionapp.FailInspection{
				Notes:         "Minor defect noted",
				QuarantineLot: false,
			},
			GotResp: &inspectionapp.FailInspectionResult{},
			CmpFunc: func(got any, exp any) string {
				result := got.(*inspectionapp.FailInspectionResult)
				if result.Inspection.Status != "failed" {
					return "expected inspection status 'failed', got '" + result.Inspection.Status + "'"
				}
				if result.LotStatus != "" {
					return "expected empty lot_status, got '" + result.LotStatus + "'"
				}
				return ""
			},
		},
	}
}

func fail403(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "fail-forbidden",
			Token:      sd.Users[0].Token,
			URL:        "/v1/inventory/quality-inspections/" + sd.Inspections[2].InspectionID + "/fail",
			Method:     http.MethodPost,
			StatusCode: http.StatusForbidden,
			Input: &inspectionapp.FailInspection{
				Notes:         "Should not work",
				QuarantineLot: true,
			},
		},
	}
}

func fail404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "fail-not-found",
			Token:      sd.Admins[0].Token,
			URL:        "/v1/inventory/quality-inspections/00000000-0000-0000-0000-000000000000/fail",
			Method:     http.MethodPost,
			StatusCode: http.StatusNotFound,
			Input: &inspectionapp.FailInspection{
				Notes:         "Does not exist",
				QuarantineLot: false,
			},
		},
	}
}
