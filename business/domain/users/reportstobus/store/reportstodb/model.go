package reportstodb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/users/reportstobus"
)

type reportsTo struct {
	ID         uuid.UUID `db:"reports_to_id"`
	BossID     uuid.UUID `db:"boss_id"`
	ReporterID uuid.UUID `db:"reporter_id"`
}

func toDBReportsTo(bus reportstobus.ReportsTo) reportsTo {
	return reportsTo{
		ID:         bus.ID,
		BossID:     bus.BossID,
		ReporterID: bus.ReporterID,
	}
}

func toBusReportsTo(at reportsTo) reportstobus.ReportsTo {
	return reportstobus.ReportsTo{
		ID:         at.ID,
		BossID:     at.BossID,
		ReporterID: at.ReporterID,
	}
}

func toBusReportsTos(ats []reportsTo) []reportstobus.ReportsTo {
	busTags := make([]reportstobus.ReportsTo, len(ats))
	for i, at := range ats {
		busTags[i] = toBusReportsTo(at)
	}
	return busTags
}
