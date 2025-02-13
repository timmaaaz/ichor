package reportstoapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/users/reportstobus"
)

func parseFilter(qp QueryParams) (reportstobus.QueryFilter, error) {
	var filter reportstobus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return reportstobus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.BossID != "" {
		bossID, err := uuid.Parse(qp.BossID)
		if err != nil {
			return reportstobus.QueryFilter{}, errs.NewFieldsError("boss_id", err)
		}
		filter.BossID = &bossID
	}

	if qp.ReporterID != "" {
		reporterID, err := uuid.Parse(qp.ReporterID)
		if err != nil {
			return reportstobus.QueryFilter{}, errs.NewFieldsError("reporter_id", err)
		}
		filter.ReporterID = &reporterID
	}

	return filter, nil
}
