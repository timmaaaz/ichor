package dbtest

import (
	"context"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus"
)

func seedUserPreferences(ctx context.Context, busDomain BusDomain, userIDs uuid.UUIDs) error {
	_, err := userpreferencesbus.TestSeedUserPreferences(ctx, userIDs, busDomain.UserPreferences)
	if err != nil {
		return err
	}

	return nil
}
