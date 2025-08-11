package contactinfosdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
)

func applyFilter(filter contactinfosbus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.FirstName != nil {
		data["first_name"] = *filter.FirstName
		wc = append(wc, "first_name = :first_name")
	}

	if filter.LastName != nil {
		data["last_name"] = *filter.LastName
		wc = append(wc, "last_name = :last_name")
	}

	if filter.EmailAddress != nil {
		data["email_address"] = *filter.EmailAddress
		wc = append(wc, "email_address = :email_address")
	}

	if filter.PrimaryPhone != nil {
		data["primary_phone_number"] = *filter.PrimaryPhone
		wc = append(wc, "primary_phone_number = :primary_phone_number")
	}

	if filter.SecondaryPhone != nil {
		data["secondary_phone_number"] = *filter.SecondaryPhone
		wc = append(wc, "secondary_phone_number = :secondary_phone_number")
	}

	if filter.StreetID != nil {
		data["street_id"] = *filter.StreetID
		wc = append(wc, "street_id = :street_id")
	}

	if filter.DeliveryAddressID != nil {
		data["delivery_address_id"] = *filter.DeliveryAddressID
		wc = append(wc, "delivery_address_id = :delivery_address_id")
	}

	// TODO figure out how to filter available hours properly through a query

	if filter.PreferredContactType != nil {
		data["preferred_contact_type"] = *filter.PreferredContactType
		wc = append(wc, "preferred_contact_type = :preferred_contact_type")
	}

	if filter.Notes != nil {
		data["notes"] = *filter.Notes
		wc = append(wc, "notes = :notes")
	}

	if filter.Timezone != nil {
		data["timezone"] = *filter.Timezone
		wc = append(wc, "timezone = :timezone")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
