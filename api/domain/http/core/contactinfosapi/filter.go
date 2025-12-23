package contactinfosapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/core/contactinfosapp"
)

func parseQueryParams(r *http.Request) (contactinfosapp.QueryParams, error) {
	values := r.URL.Query()

	filter := contactinfosapp.QueryParams{
		Page:                 values.Get("page"),
		Rows:                 values.Get("rows"),
		OrderBy:              values.Get("orderBy"),
		FirstName:            values.Get("first_name"),
		LastName:             values.Get("last_name"),
		EmailAddress:         values.Get("email_address"),
		PrimaryPhone:         values.Get("primary_phone"),
		SecondaryPhone:       values.Get("secondary_phone"),
		StreetID:             values.Get("street_id"),
		DeliveryAddressID:    values.Get("delivery_address_id"),
		ID:                   values.Get("id"),
		AvailableHoursStart:  values.Get("available_hours_start"),
		AvailableHoursEnd:    values.Get("available_hours_end"),
		TimezoneID:           values.Get("timezone_id"),
		PreferredContactType: values.Get("preferred_contact_type"),
		Notes:                values.Get("notes"),
	}

	return filter, nil

}
