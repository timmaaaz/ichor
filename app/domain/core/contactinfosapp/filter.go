package contactinfosapp

import (
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/foundation/timeutil/timeonly"
)

func parseFilter(qp QueryParams) (contactinfosbus.QueryFilter, error) {
	var filter contactinfosbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return contactinfosbus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.FirstName != "" {
		filter.FirstName = &qp.FirstName
	}

	if qp.LastName != "" {
		filter.LastName = &qp.LastName
	}

	if qp.EmailAddress != "" {
		filter.EmailAddress = &qp.EmailAddress
	}

	if qp.PrimaryPhone != "" {
		filter.PrimaryPhone = &qp.PrimaryPhone
	}

	if qp.SecondaryPhone != "" {
		filter.SecondaryPhone = &qp.SecondaryPhone
	}

	if qp.StreetID != "" {
		id, err := uuid.Parse(qp.StreetID)
		if err != nil {
			return contactinfosbus.QueryFilter{}, errs.NewFieldsError("street_id", err)
		}
		filter.StreetID = &id
	}

	if qp.DeliveryAddressID != "" {
		id, err := uuid.Parse(qp.DeliveryAddressID)
		if err != nil {
			return contactinfosbus.QueryFilter{}, errs.NewFieldsError("delivery_address_id", err)
		}
		filter.DeliveryAddressID = &id
	}

	if qp.AvailableHoursStart != "" {
		normalized, err := timeonly.NormalizeTimeOnly(qp.AvailableHoursStart)
		if err != nil {
			return contactinfosbus.QueryFilter{}, errs.NewFieldsError("available_hours_start", errors.New("not valid"))
		}
		filter.AvailableHoursStart = &normalized
	}

	if qp.AvailableHoursEnd != "" {
		normalized, err := timeonly.NormalizeTimeOnly(qp.AvailableHoursEnd)
		if err != nil {
			return contactinfosbus.QueryFilter{}, errs.NewFieldsError("available_hours_end", errors.New("not valid"))
		}
		filter.AvailableHoursEnd = &normalized
	}

	if qp.TimezoneID != "" {
		id, err := uuid.Parse(qp.TimezoneID)
		if err != nil {
			return contactinfosbus.QueryFilter{}, errs.NewFieldsError("timezone_id", err)
		}
		filter.TimezoneID = &id
	}

	if qp.PreferredContactType != "" {
		filter.PreferredContactType = &qp.PreferredContactType
	}

	if qp.Notes != "" {
		filter.Notes = &qp.Notes
	}

	return filter, nil
}
