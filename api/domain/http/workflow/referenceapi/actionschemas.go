package referenceapi

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed schemas/*.json
var schemaFS embed.FS

// ActionTypeInfo describes an action type and its configuration schema.
type ActionTypeInfo struct {
	Type           string          `json:"type"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Category       string          `json:"category"`
	SupportsManual bool            `json:"supports_manual_execution"`
	IsAsync        bool            `json:"is_async"`
	ConfigSchema   json.RawMessage `json:"config_schema"`
}

// Encode implements web.Encoder for single ActionTypeInfo.
func (a ActionTypeInfo) Encode() ([]byte, string, error) {
	data, err := json.Marshal(a)
	return data, "application/json", err
}

// actionTypeMetadata contains the non-schema metadata for each action type.
// IMPORTANT: These schemas MUST match the TypeScript types in src/types/workflow.ts
// Schemas are loaded from embedded JSON files in init().
var actionTypeMetadata = map[string]struct {
	Name           string
	Description    string
	Category       string
	SupportsManual bool
	IsAsync        bool
}{
	"allocate_inventory": {
		Name:           "Allocate Inventory",
		Description:    "Reserves or allocates inventory for an order or entity",
		Category:       "inventory",
		SupportsManual: true,
		IsAsync:        true,
	},
	"create_alert": {
		Name:           "Create Alert",
		Description:    "Creates an alert in the system with optional notification to users/roles",
		Category:       "communication",
		SupportsManual: true,
		IsAsync:        false,
	},
	"seek_approval": {
		Name:           "Seek Approval",
		Description:    "Creates an approval request for specified users",
		Category:       "approval",
		SupportsManual: false,
		IsAsync:        true,
	},
	"send_email": {
		Name:           "Send Email",
		Description:    "Sends an email to specified recipients",
		Category:       "communication",
		SupportsManual: true,
		IsAsync:        true,
	},
	"send_notification": {
		Name:           "Send Notification",
		Description:    "Sends in-app notifications through various channels",
		Category:       "communication",
		SupportsManual: true,
		IsAsync:        false,
	},
	"update_field": {
		Name:           "Update Field",
		Description:    "Updates a field on the triggered entity or related entity",
		Category:       "data",
		SupportsManual: true,
		IsAsync:        false,
	},
}

// actionTypes is populated in init() with schemas loaded from embedded files.
var actionTypes map[string]ActionTypeInfo

func init() {
	actionTypes = make(map[string]ActionTypeInfo, len(actionTypeMetadata))

	for typeName, meta := range actionTypeMetadata {
		schemaPath := fmt.Sprintf("schemas/%s.json", typeName)
		schemaBytes, err := schemaFS.ReadFile(schemaPath)
		if err != nil {
			panic(fmt.Sprintf("failed to load schema for %s: %v", typeName, err))
		}

		// Validate JSON is valid
		var dummy interface{}
		if err := json.Unmarshal(schemaBytes, &dummy); err != nil {
			panic(fmt.Sprintf("invalid JSON schema for %s: %v", typeName, err))
		}

		actionTypes[typeName] = ActionTypeInfo{
			Type:           typeName,
			Name:           meta.Name,
			Description:    meta.Description,
			Category:       meta.Category,
			SupportsManual: meta.SupportsManual,
			IsAsync:        meta.IsAsync,
			ConfigSchema:   schemaBytes,
		}
	}
}

// GetActionTypes returns all action types in alphabetical order.
// Exported for testing to verify schema alignment with TypeScript types.
func GetActionTypes() []ActionTypeInfo {
	// Return in alphabetical order for consistency
	orderedTypes := []string{
		"allocate_inventory",
		"create_alert",
		"seek_approval",
		"send_email",
		"send_notification",
		"update_field",
	}

	types := make([]ActionTypeInfo, 0, len(orderedTypes))
	for _, name := range orderedTypes {
		if t, ok := actionTypes[name]; ok {
			types = append(types, t)
		}
	}
	return types
}

// getActionTypeSchema returns the schema for a specific action type.
func getActionTypeSchema(actionType string) (ActionTypeInfo, bool) {
	info, found := actionTypes[actionType]
	if !found {
		return ActionTypeInfo{}, false
	}

	return info, true
}
