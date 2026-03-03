package dbtest

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/business/domain/hr/officebus"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// GeographyHRSeed holds the results of seeding geography and HR data.
type GeographyHRSeed struct {
	Cities       []citybus.City
	Streets      []streetbus.Street
	ContactInfos []contactinfosbus.ContactInfos
	Offices      []officebus.Office
}

func seedGeographyHR(ctx context.Context, busDomain BusDomain) (GeographyHRSeed, error) {
	count := 5

	// ADDRESSES
	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return GeographyHRSeed{}, fmt.Errorf("querying regions : %w", err)
	}
	ids := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		ids = append(ids, r.ID)
	}

	ctys, err := citybus.TestSeedCities(ctx, count, ids, busDomain.City)
	if err != nil {
		return GeographyHRSeed{}, fmt.Errorf("seeding cities : %w", err)
	}

	ctyIDs := make([]uuid.UUID, 0, len(ctys))
	for _, c := range ctys {
		ctyIDs = append(ctyIDs, c.ID)
	}

	strs, err := streetbus.TestSeedStreets(ctx, count, ctyIDs, busDomain.Street)
	if err != nil {
		return GeographyHRSeed{}, fmt.Errorf("seeding streets : %w", err)
	}
	strIDs := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		strIDs = append(strIDs, s.ID)
	}

	offices, err := officebus.TestSeedOffices(ctx, 10, strIDs, busDomain.Office)
	if err != nil {
		return GeographyHRSeed{}, fmt.Errorf("seeding offices : %w", err)
	}

	// Query timezones from seed data for contact_infos FK
	tzs, err := busDomain.Timezone.Query(ctx, timezonebus.QueryFilter{}, timezonebus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		return GeographyHRSeed{}, fmt.Errorf("querying timezones : %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, count, strIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return GeographyHRSeed{}, fmt.Errorf("seeding contact info : %w", err)
	}

	return GeographyHRSeed{
		Cities:       ctys,
		Streets:      strs,
		ContactInfos: contactInfos,
		Offices:      offices,
	}, nil
}
