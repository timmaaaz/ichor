package contactinfosapp

import (
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("first_name", order.ASC)

var orderByFields = map[string]string{
	"id":                     contactinfosbus.OrderByID,
	"first_name":             contactinfosbus.OrderByFirstName,
	"last_name":              contactinfosbus.OrderByLastName,
	"email_address":          contactinfosbus.OrderByEmail,
	"primary_phone_number":   contactinfosbus.OrderByPrimaryPhoneNumber,
	"secondary_phone_number": contactinfosbus.OrderBySecondaryPhoneNumber,
	"street_id":              contactinfosbus.OrderByStreetID,
	"available_hours_start":  contactinfosbus.OrderByAvailableHoursStart,
	"available_hours_end":    contactinfosbus.OrderByAvailableHoursEnd,
	"timezone":               contactinfosbus.OrderByTimezone,
	"preferred_contact_type": contactinfosbus.OrderByPreferredContactType,
	"notes":                  contactinfosbus.OrderByNotes,
}
