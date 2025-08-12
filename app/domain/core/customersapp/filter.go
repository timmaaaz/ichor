package customersapp

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/customersbus"
)

func parseFilter(qp QueryParams) (customersbus.QueryFilter, error) {
	var filter customersbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return customersbus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.ContactID != "" {
		id, err := uuid.Parse(qp.ContactID)
		if err != nil {
			return customersbus.QueryFilter{}, errs.NewFieldsError("contact_id", err)
		}
		filter.ContactID = &id
	}

	if qp.DeliveryAddressID != "" {
		id, err := uuid.Parse(qp.DeliveryAddressID)
		if err != nil {
			return customersbus.QueryFilter{}, errs.NewFieldsError("delivery_address_id", err)
		}
		filter.DeliveryAddressID = &id
	}

	if qp.Notes != "" {
		filter.Notes = &qp.Notes
	}

	if qp.CreatedBy != "" {
		id, err := uuid.Parse(qp.CreatedBy)
		if err != nil {
			return customersbus.QueryFilter{}, errs.NewFieldsError("created_by", err)
		}
		filter.CreatedBy = &id
	}

	if qp.UpdatedBy != "" {
		id, err := uuid.Parse(qp.UpdatedBy)
		if err != nil {
			return customersbus.QueryFilter{}, errs.NewFieldsError("updated_by", err)
		}
		filter.UpdatedBy = &id
	}

	if qp.StartCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.StartCreatedDate)
		if err != nil {
			return customersbus.QueryFilter{}, errs.NewFieldsError("start_created_date", err)
		}
		filter.StartCreatedDate = &t
	}

	if qp.EndCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.EndCreatedDate)
		if err != nil {
			return customersbus.QueryFilter{}, errs.NewFieldsError("end_created_date", err)
		}
		filter.EndCreatedDate = &t
	}

	if qp.EndCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.EndCreatedDate)
		if err != nil {
			return customersbus.QueryFilter{}, errs.NewFieldsError("end_created_date", err)
		}
		filter.EndCreatedDate = &t
	}

	if qp.StartUpdatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.StartUpdatedDate)
		if err != nil {
			return customersbus.QueryFilter{}, errs.NewFieldsError("start_updated_date", err)
		}
		filter.StartUpdatedDate = &t
	}

	if qp.EndUpdatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.EndUpdatedDate)
		if err != nil {
			return customersbus.QueryFilter{}, errs.NewFieldsError("end_updated_date", err)
		}
		filter.EndUpdatedDate = &t
	}

	return filter, nil
}
