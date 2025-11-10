package pagecontentapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
)

func parseFilter(qp QueryParams) (pagecontentbus.QueryFilter, error) {
	var filter pagecontentbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return pagecontentbus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.PageConfigID != "" {
		pageConfigID, err := uuid.Parse(qp.PageConfigID)
		if err != nil {
			return pagecontentbus.QueryFilter{}, errs.NewFieldsError("page_config_id", err)
		}
		filter.PageConfigID = &pageConfigID
	}

	if qp.ContentType != "" {
		filter.ContentType = &qp.ContentType
	}

	if qp.ParentID != "" {
		parentID, err := uuid.Parse(qp.ParentID)
		if err != nil {
			return pagecontentbus.QueryFilter{}, errs.NewFieldsError("parent_id", err)
		}
		filter.ParentID = &parentID
	}

	if qp.IsVisible != "" {
		isVisible := qp.IsVisible == "true"
		filter.IsVisible = &isVisible
	}

	return filter, nil
}
