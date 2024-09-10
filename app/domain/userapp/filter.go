package userapp

import (
	"net/mail"
	"strconv"
	"time"

	"bitbucket.org/superiortechnologies/ichor/app/sdk/errs"
	"bitbucket.org/superiortechnologies/ichor/business/domain/userbus"
	"github.com/google/uuid"
)

func parseFilter(qp QueryParams) (userbus.QueryFilter, error) {
	var filter userbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("user_id", err)
		}
		filter.ID = &id
	}

	if qp.RequestedBy != "" {
		id, err := uuid.Parse(qp.RequestedBy)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("requested_by", err)
		}
		filter.RequestedBy = &id
	}

	if qp.ApprovedBy != "" {
		id, err := uuid.Parse(qp.ApprovedBy)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("approved_by", err)
		}
		filter.ApprovedBy = &id
	}

	if qp.TitleID != "" {
		id, err := uuid.Parse(qp.TitleID)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("title_id", err)
		}
		filter.TitleID = &id
	}

	if qp.OfficeID != "" {
		id, err := uuid.Parse(qp.OfficeID)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("office_id", err)
		}
		filter.OfficeID = &id
	}

	if qp.Username != "" {
		username, err := userbus.ParseName(qp.Username)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("username", err)
		}
		filter.Username = &username
	}

	if qp.FirstName != "" {
		firstName, err := userbus.ParseName(qp.FirstName)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("first_name", err)
		}
		filter.FirstName = &firstName
	}

	if qp.LastName != "" {
		lastName, err := userbus.ParseName(qp.LastName)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("last_name", err)
		}
		filter.LastName = &lastName
	}

	if qp.Email != "" {
		addr, err := mail.ParseAddress(qp.Email)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("email", err)
		}
		filter.Email = addr
	}

	if qp.Enabled != "" {
		enabled, err := strconv.ParseBool(qp.Enabled)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("enabled", err)
		}
		filter.Enabled = &enabled
	}

	if qp.StartBirthday != "" {
		t, err := time.Parse(time.RFC3339, qp.StartBirthday)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("start_birthday", err)
		}
		filter.StartBirthday = &t
	}

	if qp.EndBirthday != "" {
		t, err := time.Parse(time.RFC3339, qp.EndBirthday)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("end_birthday", err)
		}
		filter.EndBirthday = &t
	}

	if qp.StartDateHired != "" {
		t, err := time.Parse(time.RFC3339, qp.StartDateHired)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("start_date_hired", err)
		}
		filter.StartDateHired = &t
	}

	if qp.EndDateHired != "" {
		t, err := time.Parse(time.RFC3339, qp.EndDateHired)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("end_date_hired", err)
		}
		filter.EndDateHired = &t
	}

	if qp.StartDateRequested != "" {
		t, err := time.Parse(time.RFC3339, qp.StartDateRequested)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("start_date_requested", err)
		}
		filter.StartDateRequested = &t
	}

	if qp.EndDateRequested != "" {
		t, err := time.Parse(time.RFC3339, qp.EndDateRequested)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("end_date_requested", err)
		}
		filter.EndDateRequested = &t
	}

	if qp.StartDateApproved != "" {
		t, err := time.Parse(time.RFC3339, qp.StartDateApproved)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("start_date_approved", err)
		}
		filter.StartDateApproved = &t
	}

	if qp.StartCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.StartCreatedDate)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("start_created_date", err)
		}
		filter.StartCreatedDate = &t
	}

	if qp.EndCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.EndCreatedDate)
		if err != nil {
			return userbus.QueryFilter{}, errs.NewFieldsError("end_created_date", err)
		}
		filter.EndCreatedDate = &t
	}

	return filter, nil
}
