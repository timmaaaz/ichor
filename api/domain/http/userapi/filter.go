package userapi

import (
	"net/http"

	"bitbucket.org/superiortechnologies/ichor/app/domain/userapp"
)

func parseQueryParams(r *http.Request) (userapp.QueryParams, error) {
	values := r.URL.Query()

	filter := userapp.QueryParams{
		Page:               values.Get("page"),
		Rows:               values.Get("row"),
		OrderBy:            values.Get("orderBy"),
		ID:                 values.Get("user_id"),
		RequestedBy:        values.Get("requested_by"),
		ApprovedBy:         values.Get("approved_by"),
		TitleID:            values.Get("title_id"),
		OfficeID:           values.Get("office_id"),
		Username:           values.Get("username"),
		FirstName:          values.Get("first_name"),
		LastName:           values.Get("last_name"),
		Email:              values.Get("email"),
		Enabled:            values.Get("enabled"),
		StartBirthday:      values.Get("start_birthday"),
		EndBirthday:        values.Get("end_birthday"),
		StartDateHired:     values.Get("start_date_hired"),
		EndDateHired:       values.Get("end_date_hired"),
		StartDateRequested: values.Get("start_date_requested"),
		EndDateRequested:   values.Get("end_date_requested"),
		StartDateApproved:  values.Get("start_date_approved"),
		EndDateApproved:    values.Get("end_date_approved"),
		StartCreatedDate:   values.Get("start_created_date"),
		EndCreatedDate:     values.Get("end_created_date"),
	}

	return filter, nil
}
