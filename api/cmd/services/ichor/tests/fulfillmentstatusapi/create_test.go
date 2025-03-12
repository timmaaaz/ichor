package fulfillmentstatus_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assets/fulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {

	newUUID, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/assets/fulfillmentstatus",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &fulfillmentstatusapp.NewFulfillmentStatus{
				IconId: newUUID.String(),
				Name:   "TestFulfillmentStatus",
			},
			GotResp: &fulfillmentstatusapp.FulfillmentStatus{},
			ExpResp: &fulfillmentstatusapp.FulfillmentStatus{
				IconID: newUUID.String(),
				Name:   "TestFulfillmentStatus",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*fulfillmentstatusapp.FulfillmentStatus)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*fulfillmentstatusapp.FulfillmentStatus)
				expResp.ID = gotResp.ID

				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func create400(sd apitest.SeedData) []apitest.Table {

	newUUID, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}

	table := []apitest.Table{
		{
			Name:       "missing icon id",
			URL:        "/v1/assets/fulfillmentstatus",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &fulfillmentstatusapp.NewFulfillmentStatus{
				Name: "missing icon id",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"icon_id\",\"error\":\"icon_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing name",
			URL:        "/v1/assets/fulfillmentstatus",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &fulfillmentstatusapp.NewFulfillmentStatus{
				IconId: newUUID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"name\",\"error\":\"name is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
