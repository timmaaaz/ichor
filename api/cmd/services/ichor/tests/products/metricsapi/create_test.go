package metricsapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/products/metricsapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/quality/metrics",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &metricsapp.NewMetric{
				ProductID:         sd.Products[0].ProductID,
				ReturnRate:        "2.36",
				DefectRate:        "0.35",
				MeasurementPeriod: "7 days",
			},
			GotResp: &metricsapp.Metric{},
			ExpResp: &metricsapp.Metric{
				ProductID:         sd.Products[0].ProductID,
				ReturnRate:        "2.36",
				DefectRate:        "0.35",
				MeasurementPeriod: "7 days",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*metricsapp.Metric)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*metricsapp.Metric)
				expResp.MetricID = gotResp.MetricID
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing-product-id",
			URL:        "/v1/quality/metrics",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &metricsapp.NewMetric{
				ReturnRate:        "2.36",
				DefectRate:        "0.35",
				MeasurementPeriod: "7 days",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-return-rate",
			URL:        "/v1/quality/metrics",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &metricsapp.NewMetric{
				ProductID:         sd.Products[0].ProductID,
				DefectRate:        "0.35",
				MeasurementPeriod: "7 days",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"return_rate\",\"error\":\"return_rate is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-defect-rate",
			URL:        "/v1/quality/metrics",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &metricsapp.NewMetric{
				ProductID:         sd.Products[0].ProductID,
				ReturnRate:        "2.36",
				MeasurementPeriod: "7 days",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"defect_rate\",\"error\":\"defect_rate is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-measurement-period",
			URL:        "/v1/quality/metrics",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &metricsapp.NewMetric{
				ProductID:  sd.Products[0].ProductID,
				ReturnRate: "2.36",
				DefectRate: "0.35",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"measurement_period\",\"error\":\"measurement_period is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-return-rate",
			URL:        "/v1/quality/metrics",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &metricsapp.NewMetric{
				ProductID:         sd.Products[0].ProductID,
				ReturnRate:        "not-a-float",
				DefectRate:        "0.35",
				MeasurementPeriod: "7 days",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `[{"field":"returnRate","error":"parsing error: strconv.ParseFloat: parsing \"not-a-float\": invalid syntax"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-defect-rate",
			URL:        "/v1/quality/metrics",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &metricsapp.NewMetric{
				ProductID:         sd.Products[0].ProductID,
				DefectRate:        "not-a-float",
				ReturnRate:        "0.35",
				MeasurementPeriod: "7 days",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `[{"field":"defectRate","error":"parsing error: strconv.ParseFloat: parsing \"not-a-float\": invalid syntax"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-measurement-period",
			URL:        "/v1/quality/metrics",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &metricsapp.NewMetric{
				ProductID:         sd.Products[0].ProductID,
				ReturnRate:        "3",
				DefectRate:        "0.35",
				MeasurementPeriod: "7 not real units",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `[{"field":"measurementPeriod","error":"invalid interval format: \"invalid interval format for 7 not real units, must be A year(s) B month(s) C day(s)\""}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-product-id",
			URL:        "/v1/quality/metrics",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &metricsapp.NewMetric{
				ProductID:         "not-a-uuid",
				ReturnRate:        "2.36",
				DefectRate:        "0.35",
				MeasurementPeriod: "7 days",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "product-id-not-valid-fk",
			URL:        "/v1/quality/metrics",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &metricsapp.NewMetric{
				ProductID:         uuid.New().String(),
				ReturnRate:        "2.36",
				DefectRate:        "0.35",
				MeasurementPeriod: "7 days",
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "empty token",
			URL:        "/v1/quality/metrics",
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad token",
			URL:        "/v1/quality/metrics",
			Token:      sd.Admins[0].Token[:10],
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad sig",
			URL:        "/v1/quality/metrics",
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "roleadminonly",
			URL:        "/v1/quality/metrics",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: quality_metrics"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
