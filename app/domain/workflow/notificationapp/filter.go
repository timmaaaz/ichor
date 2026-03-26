package notificationapp

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
)

func parseFilter(qp QueryParams) (notificationbus.QueryFilter, error) {
	var filter notificationbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return notificationbus.QueryFilter{}, err
		}
		filter.ID = &id
	}

	if qp.IsRead != "" {
		b, err := strconv.ParseBool(qp.IsRead)
		if err != nil {
			return notificationbus.QueryFilter{}, err
		}
		filter.IsRead = &b
	}

	if qp.Priority != "" {
		filter.Priority = &qp.Priority
	}

	if qp.SourceEntityName != "" {
		filter.SourceEntityName = &qp.SourceEntityName
	}

	if qp.SourceEntityID != "" {
		id, err := uuid.Parse(qp.SourceEntityID)
		if err != nil {
			return notificationbus.QueryFilter{}, err
		}
		filter.SourceEntityID = &id
	}

	return filter, nil
}

func parseCountFilter(isRead string, userID uuid.UUID) (notificationbus.QueryFilter, error) {
	filter := notificationbus.QueryFilter{
		UserID: &userID,
	}

	if isRead != "" {
		b, err := strconv.ParseBool(isRead)
		if err != nil {
			return notificationbus.QueryFilter{}, err
		}
		filter.IsRead = &b
	}

	return filter, nil
}
