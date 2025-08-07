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

	if qp.Address != "" {
		filter.Address = &qp.Address
	}

	if qp.AvailableHoursStart != "" {
		valid := timeonly.ValidateTimeOnlyFmt(qp.AvailableHoursStart)
		if !valid {
			return contactinfosbus.QueryFilter{}, errs.NewFieldsError("available_hours_start", errors.New("not valid"))
		}
		filter.AvailableHoursStart = &qp.AvailableHoursStart
	}

	if qp.AvailableHoursEnd != "" {
		valid := timeonly.ValidateTimeOnlyFmt(qp.AvailableHoursEnd)
		if !valid {
			return contactinfosbus.QueryFilter{}, errs.NewFieldsError("available_hours_end", errors.New("not valid"))
		}
		filter.AvailableHoursEnd = &qp.AvailableHoursEnd
	}

	if qp.Timezone != "" {
		filter.Timezone = &qp.Timezone
	}

	if qp.PreferredContactType != "" {
		filter.PreferredContactType = &qp.PreferredContactType
	}

	if qp.Notes != "" {
		filter.Notes = &qp.Notes
	}

	return filter, nil
}
