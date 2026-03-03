package settingsbus

// QueryFilter defines the filters available for querying settings.
type QueryFilter struct {
	Key    *string
	Prefix *string // Filter by key prefix (e.g., "inventory.")
}
