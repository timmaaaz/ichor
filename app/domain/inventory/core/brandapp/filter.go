package brandapp

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (brandbus.QueryFilter, error) {
	var filter brandbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return brandbus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.ContactInfoID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return brandbus.QueryFilter{}, errs.NewFieldsError("contact_info_id", err)
		}
		filter.ContactInfoID = &id
	}

	if qp.CreatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return brandbus.QueryFilter{}, errs.NewFieldsError("created_date", err)
		}
		filter.CreatedDate = &t
	}

	if qp.UpdatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.UpdatedDate)
		if err != nil {
			return brandbus.QueryFilter{}, errs.NewFieldsError("updated_date", err)
		}
		filter.UpdatedDate = &t
	}

	return filter, nil
}
