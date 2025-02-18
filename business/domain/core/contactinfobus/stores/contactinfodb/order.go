package contactinfodb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	contactinfobus.OrderByID:                   "contact_info_id",
	contactinfobus.OrderByAddress:              "address",
	contactinfobus.OrderByEmail:                "email_address",
	contactinfobus.OrderByPrimaryPhoneNumber:   "primary_phone_number",
	contactinfobus.OrderBySecondaryPhoneNumber: "secondary_phone_number",
	contactinfobus.OrderByTimezone:             "timezone",
	contactinfobus.OrderByPreferredContactType: "preferred_contact_type",
	contactinfobus.OrderByNotes:                "notes",
	contactinfobus.OrderByAvailableHoursStart:  "available_hours_start",
	contactinfobus.OrderByAvailableHoursEnd:    "available_hours_end",
	contactinfobus.OrderByFirstName:            "first_name",
	contactinfobus.OrderByLastName:             "last_name",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
