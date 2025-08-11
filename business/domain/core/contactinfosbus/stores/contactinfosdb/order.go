package contactinfosdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	contactinfosbus.OrderByID:                   "id",
	contactinfosbus.OrderByStreetID:             "street_id",
	contactinfosbus.OrderByDeliveryAddressID:    "delivery_address_id",
	contactinfosbus.OrderByEmail:                "email_address",
	contactinfosbus.OrderByPrimaryPhoneNumber:   "primary_phone_number",
	contactinfosbus.OrderBySecondaryPhoneNumber: "secondary_phone_number",
	contactinfosbus.OrderByTimezone:             "timezone",
	contactinfosbus.OrderByPreferredContactType: "preferred_contact_type",
	contactinfosbus.OrderByNotes:                "notes",
	contactinfosbus.OrderByAvailableHoursStart:  "available_hours_start",
	contactinfosbus.OrderByAvailableHoursEnd:    "available_hours_end",
	contactinfosbus.OrderByFirstName:            "first_name",
	contactinfosbus.OrderByLastName:             "last_name",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
