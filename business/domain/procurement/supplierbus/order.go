package supplierbus

import (
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

const (
	OrderBySupplierID     = "id"
	OrderByContactInfosID = "contact_infos_id"
	OrderByName           = "name"
	OrderByPaymentTermID  = "payment_term_id"
	OrderByLeadTimeDays   = "lead_time_days"
	OrderByRating         = "rating"
	OrderByIsActive       = "is_active"
	OrderByCreatedDate    = "created_date"
	OrderByUpdatedDate    = "updated_date"
)
