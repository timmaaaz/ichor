package contactinfobus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID                   = "id"
	OrderByFirstName            = "first_name"
	OrderByLastName             = "last_name"
	OrderByEmail                = "email_address"
	OrderByPrimaryPhoneNumber   = "primary_phone_number"
	OrderBySecondaryPhoneNumber = "secondary_phone_number"
	OrderByAddress              = "address"
	OrderByAvailableHoursStart  = "available_hours_start"
	OrderByAvailableHoursEnd    = "available_hours_end"
	OrderByTimezone             = "timezone"
	OrderByPreferredContactType = "preferred_contact_type"
	OrderByNotes                = "notes"
)
