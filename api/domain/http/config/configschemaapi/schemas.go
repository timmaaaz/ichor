package configschemaapi

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed schemas/*.json
var schemaFS embed.FS

// configSchemas holds the loaded JSON schemas keyed by name.
var configSchemas map[string]json.RawMessage

// contentTypes holds the static content type definitions.
var contentTypes []ContentTypeInfo

// pageActionTypes holds the static page action type definitions.
var pageActionTypes []PageActionTypeInfo

func init() {
	// Load table_config and layout schemas.
	schemaNames := []string{"table_config", "layout"}
	configSchemas = make(map[string]json.RawMessage, len(schemaNames))

	for _, name := range schemaNames {
		schemaPath := fmt.Sprintf("schemas/%s.json", name)
		schemaBytes, err := schemaFS.ReadFile(schemaPath)
		if err != nil {
			panic(fmt.Sprintf("failed to load config schema %s: %v", name, err))
		}

		var dummy any
		if err := json.Unmarshal(schemaBytes, &dummy); err != nil {
			panic(fmt.Sprintf("invalid JSON in config schema %s: %v", name, err))
		}

		configSchemas[name] = schemaBytes
	}

	// Load content types from the embedded JSON file.
	ctBytes, err := schemaFS.ReadFile("schemas/content_types.json")
	if err != nil {
		panic(fmt.Sprintf("failed to load content_types.json: %v", err))
	}

	var ctSchema struct {
		Default []ContentTypeInfo `json:"default"`
	}
	if err := json.Unmarshal(ctBytes, &ctSchema); err != nil {
		panic(fmt.Sprintf("invalid content_types.json: %v", err))
	}

	contentTypes = ctSchema.Default

	// Load page action types from the embedded JSON file.
	patBytes, err := schemaFS.ReadFile("schemas/page_action_types.json")
	if err != nil {
		panic(fmt.Sprintf("failed to load page_action_types.json: %v", err))
	}

	var patSchema struct {
		Default []PageActionTypeInfo `json:"default"`
	}
	if err := json.Unmarshal(patBytes, &patSchema); err != nil {
		panic(fmt.Sprintf("invalid page_action_types.json: %v", err))
	}

	pageActionTypes = patSchema.Default
}
