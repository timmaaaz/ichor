package contactinfoapp

import (
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("first_name", order.ASC)

var orderByFields = map[string]string{
	"contact_info_id":        contactinfobus.OrderByID,
	"first_name":             contactinfobus.OrderByFirstName,
	"last_name":              contactinfobus.OrderByLastName,
	"email_address":          contactinfobus.OrderByEmail,
	"primary_phone_number":   contactinfobus.OrderByPrimaryPhoneNumber,
	"secondary_phone_number": contactinfobus.OrderBySecondaryPhoneNumber,
	"address":                contactinfobus.OrderByAddress,
	"available_hours_start":  contactinfobus.OrderByAvailableHoursStart,
	"available_hours_end":    contactinfobus.OrderByAvailableHoursEnd,
	"timezone":               contactinfobus.OrderByTimezone,
	"preferred_contact_type": contactinfobus.OrderByPreferredContactType,
	"notes":                  contactinfobus.OrderByNotes,
}
