package supplierdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	supplierbus.OrderBySupplierID:     "id`",
	supplierbus.OrderByContactInfosID: "contact_infos_id",
	supplierbus.OrderByName:           "name",
	supplierbus.OrderByPaymentTerms:   "payment_terms",
	supplierbus.OrderByLeadTimeDays:   "lead_time_days",
	supplierbus.OrderByRating:         "rating",
	supplierbus.OrderByIsActive:       "is_active",
	supplierbus.OrderByCreatedDate:    "created_date",
	supplierbus.OrderByUpdatedDate:    "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
