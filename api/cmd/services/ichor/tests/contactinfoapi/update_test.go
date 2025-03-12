package contactinfoapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/contactinfoapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/core/contactinfo/%s", sd.ContactInfo[1].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &contactinfoapp.UpdateContactInfo{
				FirstName: dbtest.StringPointer("Conner"),
			},
			GotResp: &contactinfoapp.ContactInfo{},
			ExpResp: &contactinfoapp.ContactInfo{
				ID:                   sd.ContactInfo[1].ID,
				FirstName:            "Conner",
				LastName:             sd.ContactInfo[1].LastName,
				EmailAddress:         sd.ContactInfo[1].EmailAddress,
				PrimaryPhone:         sd.ContactInfo[1].PrimaryPhone,
				Address:              sd.ContactInfo[1].Address,
				AvailableHoursStart:  sd.ContactInfo[1].AvailableHoursStart,
				AvailableHoursEnd:    sd.ContactInfo[1].AvailableHoursEnd,
				Timezone:             sd.ContactInfo[1].Timezone,
				PreferredContactType: sd.ContactInfo[1].PreferredContactType,
				Notes:                sd.ContactInfo[1].Notes,
				SecondaryPhone:       sd.ContactInfo[1].SecondaryPhone,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*contactinfoapp.ContactInfo)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*contactinfoapp.ContactInfo)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "bad-id",
			URL:        "/v1/core/contactinfo/abc",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input:      &contactinfoapp.UpdateContactInfo{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "invalid UUID length: 3"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-available-hours-start",
			URL:        "/v1/core/contactinfo/" + sd.ContactInfo[1].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &contactinfoapp.UpdateContactInfo{
				AvailableHoursStart: dbtest.StringPointer("abcdefghij"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `invalid time format for starting hours: "abcdefghij"`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-available-hours-end",
			URL:        "/v1/core/contactinfo/" + sd.ContactInfo[1].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &contactinfoapp.UpdateContactInfo{
				AvailableHoursEnd: dbtest.StringPointer("abcdefghij"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `invalid time format for ending hours: "abcdefghij"`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func update401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/core/contactinfo/%s", sd.ContactInfo[0].ID),
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
			URL:        fmt.Sprintf("/v1/core/contactinfo/%s", sd.ContactInfo[0].ID),
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
			URL:        fmt.Sprintf("/v1/core/contactinfo/%s", sd.ContactInfo[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: contact_info"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
