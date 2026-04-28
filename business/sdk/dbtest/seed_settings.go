package dbtest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/config/settingsbus"
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus/levers"
)

// seedSettings inserts the 11 SMB-default lever keys into config.settings.
// Called from seedFrontend.go BEFORE seedScenarios so scenarios that
// reference these keys via lever_overrides have something to override.
//
// Idempotency: settingsbus.Create returns ErrUniqueEntry on duplicate key.
// We treat that as success because make reseed-frontend wipes the DB
// before re-running, so a duplicate at this point would mean two callers
// in the same chain — a developer error worth surfacing.
func seedSettings(ctx context.Context, busDomain BusDomain) error {
	for _, key := range levers.KnownKeys {
		raw, err := json.Marshal(levers.Defaults[key])
		if err != nil {
			return fmt.Errorf("seed setting marshal %s: %w", key, err)
		}
		ns := settingsbus.NewSetting{
			Key:         key,
			Value:       json.RawMessage(raw),
			Description: "Phase 0g.B5 lever — see design doc 2026-04-24 §3.3",
		}
		if _, err := busDomain.Settings.Create(ctx, ns); err != nil {
			return fmt.Errorf("seed setting %s: %w", key, err)
		}
	}
	return nil
}
