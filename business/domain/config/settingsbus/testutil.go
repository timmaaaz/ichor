package settingsbus

import (
	"context"
	"encoding/json"
	"sort"
)

func TestNewSetting(keys []string, values []string) []NewSetting {
	newSettings := make([]NewSetting, len(keys))
	for i, key := range keys {
		newSettings[i] = NewSetting{
			Key:         key,
			Value:       json.RawMessage(values[i%len(values)]),
			Description: "Test setting " + key,
		}
	}
	return newSettings
}

func TestSeedSettings(ctx context.Context, keys []string, values []string, api *Business) ([]Setting, error) {
	newSettings := TestNewSetting(keys, values)

	settings := make([]Setting, len(newSettings))
	for i, ns := range newSettings {
		s, err := api.Create(ctx, ns)
		if err != nil {
			return []Setting{}, err
		}
		settings[i] = s
	}

	sort.Slice(settings, func(i, j int) bool {
		return settings[i].Key < settings[j].Key
	})

	return settings, nil
}
