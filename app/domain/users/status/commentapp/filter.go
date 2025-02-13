package commentapp

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/users/status/commentbus"
)

func parseFilter(qp QueryParams) (commentbus.QueryFilter, error) {
	var filter commentbus.QueryFilter

	if qp.UserID != "" {
		id, err := uuid.Parse(qp.UserID)
		if err != nil {
			return commentbus.QueryFilter{}, errs.NewFieldsError("user_id", err)
		}
		filter.UserID = &id
	}

	if qp.CommenterID != "" {
		id, err := uuid.Parse(qp.CommenterID)
		if err != nil {
			return commentbus.QueryFilter{}, errs.NewFieldsError("commenter_id", err)
		}
		filter.CommenterID = &id
	}

	if qp.Comment != "" {
		filter.Comment = &qp.Comment
	}

	if qp.CreatedDate != "" {
		date, err := time.Parse(time.RFC3339Nano, qp.CreatedDate)
		if err != nil {
			return commentbus.QueryFilter{}, errs.NewFieldsError("created_date", err)
		}
		filter.CreatedDate = &date
	}

	return filter, nil
}
