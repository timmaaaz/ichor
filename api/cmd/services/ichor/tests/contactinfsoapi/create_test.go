package contactinfosapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/contactinfosapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/contactinfos",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &contactinfosapp.NewContactInfos{
				FirstName:            "John",
				LastName:             "Doe",
				EmailAddress:         "johndoe@example.com",
				PrimaryPhone:         "+1234567890",
				StreetID:             sd.Streets[0].ID,
				AvailableHoursStart:  "8:00:00",
				AvailableHoursEnd:    "17:00:00",
				Timezone:             "America/New_York",
				PreferredContactType: "email",
			},
			GotResp: &contactinfosapp.ContactInfos{},
			ExpResp: &contactinfosapp.ContactInfos{
				FirstName:            "John",
				LastName:             "Doe",
				EmailAddress:         "johndoe@example.com",
				PrimaryPhone:         "+1234567890",
				StreetID:             sd.Streets[0].ID,
				AvailableHoursStart:  "8:00:00",
				AvailableHoursEnd:    "17:00:00",
				Timezone:             "America/New_York",
				PreferredContactType: "email",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*contactinfosapp.ContactInfos)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*contactinfosapp.ContactInfos)
				expResp.ID = gotResp.ID

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func create400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "missing first name",
			URL:        "/v1/core/contactinfos",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &contactinfosapp.ContactInfos{
				LastName:             "Doe",
				EmailAddress:         "johndoe@example.com",
				PrimaryPhone:         "+1234567890",
				StreetID:             sd.Streets[0].ID,
				AvailableHoursStart:  "8:00:00",
				AvailableHoursEnd:    "17:00:00",
				Timezone:             "America/New_York",
				PreferredContactType: "email",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"first_name\",\"error\":\"first_name is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing last name",
			URL:        "/v1/core/contactinfos",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &contactinfosapp.ContactInfos{
				FirstName:            "John",
				EmailAddress:         "johndoe@example.com",
				PrimaryPhone:         "+1234567890",
				StreetID:             sd.Streets[0].ID,
				AvailableHoursStart:  "8:00:00",
				AvailableHoursEnd:    "17:00:00",
				Timezone:             "America/New_York",
				PreferredContactType: "email",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"last_name\",\"error\":\"last_name is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing email",
			URL:        "/v1/core/contactinfos",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &contactinfosapp.ContactInfos{
				FirstName:            "John",
				LastName:             "Doe",
				PrimaryPhone:         "+1234567890",
				StreetID:             sd.Streets[0].ID,
				AvailableHoursStart:  "8:00:00",
				AvailableHoursEnd:    "17:00:00",
				Timezone:             "America/New_York",
				PreferredContactType: "email",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"email_address\",\"error\":\"email_address is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing primary phone",
			URL:        "/v1/core/contactinfos",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &contactinfosapp.ContactInfos{
				FirstName:            "John",
				LastName:             "Doe",
				EmailAddress:         "johndoe@example.com",
				StreetID:             sd.Streets[0].ID,
				AvailableHoursStart:  "8:00:00",
				AvailableHoursEnd:    "17:00:00",
				Timezone:             "America/New_York",
				PreferredContactType: "email",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"primary_phone\",\"error\":\"primary_phone is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing address",
			URL:        "/v1/core/contactinfos",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &contactinfosapp.ContactInfos{
				FirstName:            "John",
				LastName:             "Doe",
				EmailAddress:         "johndoe@example.com",
				PrimaryPhone:         "+1234567890",
				AvailableHoursStart:  "8:00:00",
				AvailableHoursEnd:    "17:00:00",
				Timezone:             "America/New_York",
				PreferredContactType: "email",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"street_id\",\"error\":\"street_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing available hours start",
			URL:        "/v1/core/contactinfos",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &contactinfosapp.ContactInfos{
				FirstName:            "John",
				LastName:             "Doe",
				EmailAddress:         "johndoe@example.com",
				PrimaryPhone:         "+1234567890",
				StreetID:             sd.Streets[0].ID,
				AvailableHoursEnd:    "17:00:00",
				Timezone:             "America/New_York",
				PreferredContactType: "email",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"available_hours_start\",\"error\":\"available_hours_start is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing available hours end",
			URL:        "/v1/core/contactinfos",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &contactinfosapp.ContactInfos{
				FirstName:            "John",
				LastName:             "Doe",
				EmailAddress:         "johndoe@example.com",
				PrimaryPhone:         "+1234567890",
				StreetID:             sd.Streets[0].ID,
				AvailableHoursStart:  "17:00:00",
				Timezone:             "America/New_York",
				PreferredContactType: "email",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"available_hours_end\",\"error\":\"available_hours_end is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing timezone",
			URL:        "/v1/core/contactinfos",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &contactinfosapp.ContactInfos{
				FirstName:            "John",
				LastName:             "Doe",
				EmailAddress:         "johndoe@example.com",
				PrimaryPhone:         "+1234567890",
				StreetID:             sd.Streets[0].ID,
				AvailableHoursStart:  "8:00:00",
				AvailableHoursEnd:    "17:00:00",
				PreferredContactType: "email",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"timezone\",\"error\":\"timezone is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing preferred contact types",
			URL:        "/v1/core/contactinfos",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &contactinfosapp.ContactInfos{
				FirstName:           "John",
				LastName:            "Doe",
				EmailAddress:        "johndoe@example.com",
				PrimaryPhone:        "+1234567890",
				StreetID:            sd.Streets[0].ID,
				AvailableHoursStart: "8:00:00",
				AvailableHoursEnd:   "17:00:00",
				Timezone:            "America/New_York",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"preferred_contact_type\",\"error\":\"preferred_contact_type is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func create401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "empty token",
			URL:        "/v1/core/contactinfos",
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
			URL:        "/v1/core/contactinfos",
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
			URL:        "/v1/core/contactinfos",
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
			URL:        "/v1/core/contactinfos",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: contact_infos"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
