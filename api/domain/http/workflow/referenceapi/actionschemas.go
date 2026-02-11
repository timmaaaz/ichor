package referenceapi

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

//go:embed schemas/*.json
var schemaFS embed.FS

// ActionTypeInfo describes an action type and its configuration schema.
type ActionTypeInfo struct {
	Type           string               `json:"type"`
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	Category       string               `json:"category"`
	SupportsManual bool                 `json:"supports_manual_execution"`
	IsAsync        bool                 `json:"is_async"`
	ConfigSchema   json.RawMessage      `json:"config_schema"`
	OutputPorts    []workflow.OutputPort `json:"output_ports"`
}

// Encode implements web.Encoder for single ActionTypeInfo.
func (a ActionTypeInfo) Encode() ([]byte, string, error) {
	data, err := json.Marshal(a)
	return data, "application/json", err
}

// actionTypeMeta holds static metadata for an action type.
type actionTypeMeta struct {
	Name           string
	Description    string
	Category       string
	SupportsManual bool
	IsAsync        bool
}

// actionTypeMetadata contains the non-schema metadata for each action type.
// Schemas are loaded from embedded JSON files in schemas/*.json.
var actionTypeMetadata = map[string]actionTypeMeta{
	"allocate_inventory": {
		Name:           "Allocate Inventory",
		Description:    "Reserves or allocates inventory for an order or entity",
		Category:       "inventory",
		SupportsManual: true,
		IsAsync:        true,
	},
	"log_audit_entry": {
		Name:           "Log Audit Entry",
		Description:    "Write an audit trail entry to the workflow audit log",
		Category:       "data",
		SupportsManual: true,
		IsAsync:        false,
	},
	"check_inventory": {
		Name:           "Check Inventory",
		Description:    "Check inventory availability against a threshold and branch accordingly",
		Category:       "inventory",
		SupportsManual: false,
		IsAsync:        false,
	},
	"check_reorder_point": {
		Name:           "Check Reorder Point",
		Description:    "Check if inventory is below its reorder point and branch accordingly",
		Category:       "inventory",
		SupportsManual: false,
		IsAsync:        false,
	},
	"commit_allocation": {
		Name:           "Commit Allocation",
		Description:    "Commit reserved inventory to allocated status",
		Category:       "inventory",
		SupportsManual: true,
		IsAsync:        false,
	},
	"create_alert": {
		Name:           "Create Alert",
		Description:    "Creates an alert in the system with optional notification to users/roles",
		Category:       "communication",
		SupportsManual: true,
		IsAsync:        false,
	},
	"create_entity": {
		Name:           "Create Entity",
		Description:    "Create a new entity record in the database",
		Category:       "data",
		SupportsManual: true,
		IsAsync:        false,
	},
	"delay": {
		Name:           "Delay",
		Description:    "Pause workflow execution for a specified duration using durable timers",
		Category:       "control",
		SupportsManual: false,
		IsAsync:        false,
	},
	"evaluate_condition": {
		Name:           "Evaluate Condition",
		Description:    "Evaluates conditions against entity data and determines branch direction for workflow execution",
		Category:       "control",
		SupportsManual: false,
		IsAsync:        false,
	},
	"lookup_entity": {
		Name:           "Lookup Entity",
		Description:    "Look up entity data by filter criteria for use in downstream actions",
		Category:       "data",
		SupportsManual: true,
		IsAsync:        false,
	},
	"release_reservation": {
		Name:           "Release Reservation",
		Description:    "Release reserved inventory quantity back to available stock",
		Category:       "inventory",
		SupportsManual: true,
		IsAsync:        false,
	},
	"reserve_inventory": {
		Name:           "Reserve Inventory",
		Description:    "Reserve inventory items for future allocation with idempotency guarantees",
		Category:       "inventory",
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
	"transition_status": {
		Name:           "Transition Status",
		Description:    "Transition an entity field from one status to another with validation",
		Category:       "data",
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

// actionTypeSchemas holds the loaded JSON schemas keyed by action type.
var actionTypeSchemas map[string]json.RawMessage

func init() {
	actionTypeSchemas = make(map[string]json.RawMessage, len(actionTypeMetadata))

	for typeName := range actionTypeMetadata {
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

		actionTypeSchemas[typeName] = schemaBytes
	}
}

// GetActionTypes returns all action types in alphabetical order, enriched
// with output ports from the ActionRegistry when available.
func GetActionTypes(registry *workflow.ActionRegistry) []ActionTypeInfo {
	typeNames := make([]string, 0, len(actionTypeMetadata))
	for name := range actionTypeMetadata {
		typeNames = append(typeNames, name)
	}
	sort.Strings(typeNames)

	types := make([]ActionTypeInfo, 0, len(typeNames))
	for _, name := range typeNames {
		meta := actionTypeMetadata[name]

		var ports []workflow.OutputPort
		if registry != nil {
			ports = registry.GetOutputPorts(name)
		}
		if ports == nil {
			ports = workflow.DefaultOutputPorts()
		}

		types = append(types, ActionTypeInfo{
			Type:           name,
			Name:           meta.Name,
			Description:    meta.Description,
			Category:       meta.Category,
			SupportsManual: meta.SupportsManual,
			IsAsync:        meta.IsAsync,
			ConfigSchema:   actionTypeSchemas[name],
			OutputPorts:    ports,
		})
	}
	return types
}

// getActionTypeSchema returns the full info for a specific action type.
func getActionTypeSchema(actionType string, registry *workflow.ActionRegistry) (ActionTypeInfo, bool) {
	meta, found := actionTypeMetadata[actionType]
	if !found {
		return ActionTypeInfo{}, false
	}

	schema, hasSchema := actionTypeSchemas[actionType]
	if !hasSchema {
		return ActionTypeInfo{}, false
	}

	var ports []workflow.OutputPort
	if registry != nil {
		ports = registry.GetOutputPorts(actionType)
	}
	if ports == nil {
		ports = workflow.DefaultOutputPorts()
	}

	return ActionTypeInfo{
		Type:           actionType,
		Name:           meta.Name,
		Description:    meta.Description,
		Category:       meta.Category,
		SupportsManual: meta.SupportsManual,
		IsAsync:        meta.IsAsync,
		ConfigSchema:   schema,
		OutputPorts:    ports,
	}, true
}
