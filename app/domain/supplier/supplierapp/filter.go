package supplierapp

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierbus/types"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (supplierbus.QueryFilter, error) {
	var filter supplierbus.QueryFilter

	if qp.SupplierID != "" {
		id, err := uuid.Parse(qp.SupplierID)
		if err != nil {
			return supplierbus.QueryFilter{}, errs.NewFieldsError("supplier_id", err)
		}
		filter.SupplierID = &id
	}

	if qp.ContactID != "" {
		id, err := uuid.Parse(qp.ContactID)
		if err != nil {
			return supplierbus.QueryFilter{}, errs.NewFieldsError("contact_id", err)
		}
		filter.ContactID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.CreatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return supplierbus.QueryFilter{}, errs.NewFieldsError("created_date", err)
		}
		filter.CreatedDate = &t
	}

	if qp.UpdatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.UpdatedDate)
		if err != nil {
			return supplierbus.QueryFilter{}, errs.NewFieldsError("updated_date", err)
		}
		filter.UpdatedDate = &t
	}

	if qp.IsActive != "" {
		isActive, err := strconv.ParseBool(qp.IsActive)
		if err != nil {
			return supplierbus.QueryFilter{}, errs.NewFieldsError("is_active", err)
		}
		filter.IsActive = &isActive
	}

	if qp.LeadTimeDays != "" {
		leadTimeDays, err := strconv.Atoi(qp.LeadTimeDays)
		if err != nil {
			return supplierbus.QueryFilter{}, errs.NewFieldsError("lead_time_days", err)
		}
		filter.LeadTimeDays = &leadTimeDays
	}

	if qp.Rating != "" {
		rating, err := types.ParseRoundedFloat(qp.Rating)
		if err != nil {
			return supplierbus.QueryFilter{}, errs.NewFieldsError("rating", err)
		}
		filter.Rating = &rating
	}

	if qp.PaymentTerms != "" {
		filter.PaymentTerms = &qp.PaymentTerms
	}

	return filter, nil
}
