package supplierapp

import (
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"id":               supplierbus.OrderBySupplierID,
	"contact_infos_id": supplierbus.OrderByContactInfosID,
	"name":             supplierbus.OrderByName,
	"payment_terms":    supplierbus.OrderByPaymentTerms,
	"lead_time_days":   supplierbus.OrderByLeadTimeDays,
	"rating":           supplierbus.OrderByRating,
	"is_active":        supplierbus.OrderByIsActive,
	"created_date":     supplierbus.OrderByCreatedDate,
	"updated_date":     supplierbus.OrderByUpdatedDate,
}
