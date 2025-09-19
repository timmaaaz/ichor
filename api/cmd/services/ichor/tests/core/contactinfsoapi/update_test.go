package contactinfosapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/contactinfosapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/core/contact-infos/%s", sd.ContactInfos[1].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &contactinfosapp.UpdateContactInfos{
				FirstName: dbtest.StringPointer("Conner"),
			},
			GotResp: &contactinfosapp.ContactInfos{},
			ExpResp: &contactinfosapp.ContactInfos{
				ID:                   sd.ContactInfos[1].ID,
				FirstName:            "Conner",
				LastName:             sd.ContactInfos[1].LastName,
				EmailAddress:         sd.ContactInfos[1].EmailAddress,
				PrimaryPhone:         sd.ContactInfos[1].PrimaryPhone,
				StreetID:             sd.ContactInfos[1].StreetID,
				DeliveryAddressID:    sd.ContactInfos[1].DeliveryAddressID,
				AvailableHoursStart:  sd.ContactInfos[1].AvailableHoursStart,
				AvailableHoursEnd:    sd.ContactInfos[1].AvailableHoursEnd,
				Timezone:             sd.ContactInfos[1].Timezone,
				PreferredContactType: sd.ContactInfos[1].PreferredContactType,
				Notes:                sd.ContactInfos[1].Notes,
				SecondaryPhone:       sd.ContactInfos[1].SecondaryPhone,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*contactinfosapp.ContactInfos)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*contactinfosapp.ContactInfos)

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
			URL:        "/v1/core/contact-infos/abc",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input:      &contactinfosapp.UpdateContactInfos{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "invalid UUID length: 3"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-available-hours-start",
			URL:        "/v1/core/contact-infos/" + sd.ContactInfos[1].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &contactinfosapp.UpdateContactInfos{
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
			URL:        "/v1/core/contact-infos/" + sd.ContactInfos[1].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &contactinfosapp.UpdateContactInfos{
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
			URL:        fmt.Sprintf("/v1/core/contact-infos/%s", sd.ContactInfos[0].ID),
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
			URL:        fmt.Sprintf("/v1/core/contact-infos/%s", sd.ContactInfos[0].ID),
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
			URL:        fmt.Sprintf("/v1/core/contact-infos/%s", sd.ContactInfos[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: contact_infos"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
