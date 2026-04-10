package userpreferencesbus

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/google/uuid"
)

// TestSeedUserPreferences seeds user preferences for the provided user IDs.
// Each user gets a "floor.font_scale" preference set to "medium".
func TestSeedUserPreferences(ctx context.Context, userIDs uuid.UUIDs, api *Business) ([]UserPreference, error) {
	prefs := make([]UserPreference, 0, len(userIDs))

	for _, userID := range userIDs {
		np := NewUserPreference{
			UserID: userID,
			Key:    "floor.font_scale",
			Value:  json.RawMessage(`"medium"`),
		}

		pref, err := api.Set(ctx, np)
		if err != nil {
			return nil, err
		}

		prefs = append(prefs, pref)
	}

	sort.Slice(prefs, func(i, j int) bool {
		if prefs[i].UserID.String() == prefs[j].UserID.String() {
			return prefs[i].Key < prefs[j].Key
		}
		return prefs[i].UserID.String() < prefs[j].UserID.String()
	})

	return prefs, nil
}
