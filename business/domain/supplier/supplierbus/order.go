package supplierbus

import (
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

const (
	OrderBySupplierID   = "supplier_id"
	OrderByContactID    = "contact_id"
	OrderByName         = "name"
	OrderByPaymentTerms = "payment_terms"
	OrderByLeadTimeDays = "lead_time_days"
	OrderByRating       = "rating"
	OrderByIsActive     = "is_active"
	OrderByCreatedDate  = "created_date"
	OrderByUpdatedDate  = "updated_date"
)
