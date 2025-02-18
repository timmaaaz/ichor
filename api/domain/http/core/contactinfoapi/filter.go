package contactinfoapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/core/contactinfoapp"
)

func parseQueryParams(r *http.Request) (contactinfoapp.QueryParams, error) {
	values := r.URL.Query()

	filter := contactinfoapp.QueryParams{
		Page:                 values.Get("page"),
		Rows:                 values.Get("rows"),
		OrderBy:              values.Get("orderBy"),
		FirstName:            values.Get("first_name"),
		LastName:             values.Get("last_name"),
		EmailAddress:         values.Get("email_address"),
		PrimaryPhone:         values.Get("primary_phone"),
		SecondaryPhone:       values.Get("secondary_phone"),
		Address:              values.Get("address"),
		ID:                   values.Get("contact_info_id"),
		AvailableHoursStart:  values.Get("available_hours_start"),
		AvailableHoursEnd:    values.Get("available_hours_end"),
		Timezone:             values.Get("timezone"),
		PreferredContactType: values.Get("preferred_contact_type"),
		Notes:                values.Get("notes"),
	}

	return filter, nil

}
