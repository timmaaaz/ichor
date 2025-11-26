package data_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/dataapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// =============================================================================
// executeChartQuery (POST /v1/data/chart/{table_config_id})
// =============================================================================

func executeChartByID200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "kpi-chart",
			URL:        fmt.Sprintf("/v1/data/chart/%s", sd.KPIChartConfig.ID.String()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: dataapp.TableQuery{
				Filters: []dataapp.FilterParam{},
				Sort:    []dataapp.SortParam{},
				Dynamic: map[string]any{},
			},
			GotResp: &dataapp.ChartResponse{},
			ExpResp: &dataapp.ChartResponse{
				Type: "kpi",
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*dataapp.ChartResponse)
				if !exists {
					return "could not convert got to *dataapp.ChartResponse"
				}

				// Verify chart type
				if gotResp.Type != "kpi" {
					return fmt.Sprintf("expected type 'kpi', got '%s'", gotResp.Type)
				}

				// KPI charts must have KPI data
				if gotResp.KPI == nil {
					return "expected KPI data to be present, got nil"
				}

				// Verify KPI has a value (inventory items seeded have quantity > 0)
				if gotResp.KPI.Value == 0 && gotResp.KPI.Label == "" {
					return "KPI data appears empty"
				}

				// Verify meta is populated
				if gotResp.Meta.RowsProcessed == 0 {
					return "expected Meta.RowsProcessed > 0"
				}

				return ""
			},
		},
		{
			Name:       "bar-chart",
			URL:        fmt.Sprintf("/v1/data/chart/%s", sd.BarChartConfig.ID.String()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: dataapp.TableQuery{
				Filters: []dataapp.FilterParam{},
				Sort:    []dataapp.SortParam{},
				Dynamic: map[string]any{},
			},
			GotResp: &dataapp.ChartResponse{},
			ExpResp: &dataapp.ChartResponse{
				Type: "bar",
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*dataapp.ChartResponse)
				if !exists {
					return "could not convert got to *dataapp.ChartResponse"
				}

				if gotResp.Type != "bar" {
					return fmt.Sprintf("expected type 'bar', got '%s'", gotResp.Type)
				}

				// Bar charts must have categories and series
				if len(gotResp.Categories) == 0 {
					return "expected Categories to be populated for bar chart"
				}

				if len(gotResp.Series) == 0 {
					return "expected Series to be populated for bar chart"
				}

				return ""
			},
		},
		{
			Name:       "pie-chart",
			URL:        fmt.Sprintf("/v1/data/chart/%s", sd.PieChartConfig.ID.String()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: dataapp.TableQuery{
				Filters: []dataapp.FilterParam{},
				Sort:    []dataapp.SortParam{},
				Dynamic: map[string]any{},
			},
			GotResp: &dataapp.ChartResponse{},
			ExpResp: &dataapp.ChartResponse{
				Type: "pie",
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*dataapp.ChartResponse)
				if !exists {
					return "could not convert got to *dataapp.ChartResponse"
				}

				if gotResp.Type != "pie" {
					return fmt.Sprintf("expected type 'pie', got '%s'", gotResp.Type)
				}

				// Pie charts have series (each slice is a series item)
				if len(gotResp.Series) == 0 {
					return "expected Series to be populated for pie chart"
				}

				return ""
			},
		},
	}
}

func executeChartByID400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "invalid-uuid",
			URL:        "/v1/data/chart/not-a-valid-uuid",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: dataapp.TableQuery{
				Filters: []dataapp.FilterParam{},
				Sort:    []dataapp.SortParam{},
				Dynamic: map[string]any{},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				// Just verify we got an error response
				_, exists := got.(*errs.Error)
				if !exists {
					return "expected *errs.Error response"
				}
				return ""
			},
		},
	}
}

func executeChartByID401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        fmt.Sprintf("/v1/data/chart/%s", sd.KPIChartConfig.ID.String()),
			Token:      "",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "expected authorization header format: Bearer <token>"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func executeChartByID404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "config-not-found",
			URL:        fmt.Sprintf("/v1/data/chart/%s", uuid.New().String()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusNotFound,
			Input: dataapp.TableQuery{
				Filters: []dataapp.FilterParam{},
				Sort:    []dataapp.SortParam{},
				Dynamic: map[string]any{},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				_, exists := got.(*errs.Error)
				if !exists {
					return "expected *errs.Error response"
				}
				return ""
			},
		},
	}
}

// =============================================================================
// executeChartQueryByName (POST /v1/data/chart/name/{name})
// =============================================================================

func executeChartByName200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "kpi-by-name",
			URL:        "/v1/data/chart/name/kpi_inventory_chart",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: dataapp.TableQuery{
				Filters: []dataapp.FilterParam{},
				Sort:    []dataapp.SortParam{},
				Dynamic: map[string]any{},
			},
			GotResp: &dataapp.ChartResponse{},
			ExpResp: &dataapp.ChartResponse{
				Type: "kpi",
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*dataapp.ChartResponse)
				if !exists {
					return "could not convert got to *dataapp.ChartResponse"
				}

				if gotResp.Type != "kpi" {
					return fmt.Sprintf("expected type 'kpi', got '%s'", gotResp.Type)
				}

				if gotResp.KPI == nil {
					return "expected KPI data to be present"
				}

				return ""
			},
		},
		{
			Name:       "bar-by-name",
			URL:        "/v1/data/chart/name/bar_inventory_chart",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: dataapp.TableQuery{
				Filters: []dataapp.FilterParam{},
				Sort:    []dataapp.SortParam{},
				Dynamic: map[string]any{},
			},
			GotResp: &dataapp.ChartResponse{},
			ExpResp: &dataapp.ChartResponse{
				Type: "bar",
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*dataapp.ChartResponse)
				if !exists {
					return "could not convert got to *dataapp.ChartResponse"
				}

				if gotResp.Type != "bar" {
					return fmt.Sprintf("expected type 'bar', got '%s'", gotResp.Type)
				}

				return ""
			},
		},
	}
}

func executeChartByName404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "name-not-found",
			URL:        "/v1/data/chart/name/nonexistent_chart_config",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusNotFound,
			Input: dataapp.TableQuery{
				Filters: []dataapp.FilterParam{},
				Sort:    []dataapp.SortParam{},
				Dynamic: map[string]any{},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				_, exists := got.(*errs.Error)
				if !exists {
					return "expected *errs.Error response"
				}
				return ""
			},
		},
	}
}

// =============================================================================
// previewChartData (POST /v1/data/chart/preview)
// =============================================================================

func previewChartData200(sd apitest.SeedData) []apitest.Table {
	// Marshal the KPI config for preview
	configJSON, _ := json.Marshal(KPIChartConfig)

	return []apitest.Table{
		{
			Name:       "preview-kpi-chart",
			URL:        "/v1/data/chart/preview",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &dataapp.PreviewChartDataRequest{
				Config: configJSON,
				Query: dataapp.TableQuery{
					Filters: []dataapp.FilterParam{},
					Sort:    []dataapp.SortParam{},
					Dynamic: map[string]any{},
				},
			},
			GotResp: &dataapp.ChartResponse{},
			ExpResp: &dataapp.ChartResponse{
				Type: "kpi",
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*dataapp.ChartResponse)
				if !exists {
					return "could not convert got to *dataapp.ChartResponse"
				}

				if gotResp.Type != "kpi" {
					return fmt.Sprintf("expected type 'kpi', got '%s'", gotResp.Type)
				}

				if gotResp.KPI == nil {
					return "expected KPI data to be present"
				}

				return ""
			},
		},
	}
}

func previewChartData400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-config",
			URL:        "/v1/data/chart/preview",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &dataapp.PreviewChartDataRequest{
				Config: nil,
				Query:  dataapp.TableQuery{},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				_, exists := got.(*errs.Error)
				if !exists {
					return "expected *errs.Error response"
				}
				return ""
			},
		},
		{
			Name:       "invalid-json-config",
			URL:        "/v1/data/chart/preview",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &dataapp.PreviewChartDataRequest{
				Config: json.RawMessage(`{"invalid": json missing bracket`),
				Query:  dataapp.TableQuery{},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				_, exists := got.(*errs.Error)
				if !exists {
					return "expected *errs.Error response"
				}
				return ""
			},
		},
		{
			Name:       "missing-data-source",
			URL:        "/v1/data/chart/preview",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &dataapp.PreviewChartDataRequest{
				Config: json.RawMessage(`{"title":"Test","widget_type":"chart","visualization":"kpi","data_source":[]}`),
				Query:  dataapp.TableQuery{},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				_, exists := got.(*errs.Error)
				if !exists {
					return "expected *errs.Error response"
				}
				return ""
			},
		},
	}
}

func previewChartData401(sd apitest.SeedData) []apitest.Table {
	configJSON, _ := json.Marshal(KPIChartConfig)

	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        "/v1/data/chart/preview",
			Token:      "",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &dataapp.PreviewChartDataRequest{
				Config: configJSON,
				Query:  dataapp.TableQuery{},
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "expected authorization header format: Bearer <token>"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
