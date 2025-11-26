package metricsapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/products/metricsapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/products/quality-metrics/%s", sd.Metrics[0].MetricID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &metricsapp.UpdateMetric{
				ProductID:         &sd.Products[3].ProductID,
				ReturnRate:        dbtest.StringPointer("3.33"),
				DefectRate:        dbtest.StringPointer("4.32"),
				MeasurementPeriod: dbtest.StringPointer("6 months"),
			},
			GotResp: &metricsapp.Metric{},
			ExpResp: &metricsapp.Metric{
				ProductID:         sd.Products[3].ProductID,
				ReturnRate:        "3.33",
				DefectRate:        "4.32",
				MeasurementPeriod: "6 months",
				CreatedDate:       sd.Metrics[0].CreatedDate,
				MetricID:          sd.Metrics[0].MetricID,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*metricsapp.Metric)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*metricsapp.Metric)
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "malformed-product-id",
			URL:        fmt.Sprintf("/v1/products/quality-metrics/%s", sd.Metrics[0].MetricID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &metricsapp.UpdateMetric{
				ProductID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"product_id","error":"product_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-product-cost-uuid",
			URL:        fmt.Sprintf("/v1/products/quality-metrics/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &metricsapp.UpdateMetric{
				ProductID: &sd.Metrics[0].ProductID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "invalid UUID length: 10"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-product-defect-rate",
			URL:        fmt.Sprintf("/v1/products/quality-metrics/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &metricsapp.UpdateMetric{
				ProductID:  &sd.Metrics[0].ProductID,
				DefectRate: dbtest.StringPointer("not a float"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "invalid UUID length: 10"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-product-return-rate",
			URL:        fmt.Sprintf("/v1/products/quality-metrics/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &metricsapp.UpdateMetric{
				ProductID:  &sd.Metrics[0].ProductID,
				ReturnRate: dbtest.StringPointer("not a float"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "invalid UUID length: 10"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-interval",
			URL:        fmt.Sprintf("/v1/products/quality-metrics/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &metricsapp.UpdateMetric{
				ProductID:         &sd.Metrics[0].ProductID,
				MeasurementPeriod: dbtest.StringPointer("not an interval"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "invalid UUID length: 10"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/products/quality-metrics/%s", sd.Metrics[0].MetricID),
			Token:      "&nbsp",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badsig",
			URL:        fmt.Sprintf("/v1/products/quality-metrics/%s", sd.Metrics[0].MetricID),
			Token:      sd.Users[0].Token + "A",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "roleadminonly",
			URL:        fmt.Sprintf("/v1/products/quality-metrics/%s", sd.Metrics[0].MetricID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: products.quality_metrics"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func update404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "supplier-dne",
			URL:        fmt.Sprintf("/v1/products/quality-metrics/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &metricsapp.UpdateMetric{
				ProductID: &sd.Products[0].ProductID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "query by id: metric not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "contact-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/products/quality-metrics/%s", sd.Metrics[0].MetricID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &metricsapp.UpdateMetric{
				ProductID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
