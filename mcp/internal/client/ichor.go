// Package client provides an HTTP client for the Ichor REST API.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client wraps the Ichor REST API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// New creates a new Ichor API client.
func New(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request with auth headers and returns the response body.
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (json.RawMessage, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return json.RawMessage(respBody), nil
}

// get performs a GET request.
func (c *Client) get(ctx context.Context, path string) (json.RawMessage, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil)
}

// post performs a POST request with a JSON body.
func (c *Client) post(ctx context.Context, path string, payload json.RawMessage) (json.RawMessage, error) {
	return c.doRequest(ctx, http.MethodPost, path, bytes.NewReader(payload))
}

// put performs a PUT request with a JSON body.
func (c *Client) put(ctx context.Context, path string, payload json.RawMessage) (json.RawMessage, error) {
	return c.doRequest(ctx, http.MethodPut, path, bytes.NewReader(payload))
}

// GetCatalog calls GET /v1/agent/catalog.
func (c *Client) GetCatalog(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/v1/agent/catalog")
}

// GetActionTypes calls GET /v1/workflow/action-types.
func (c *Client) GetActionTypes(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/v1/workflow/action-types")
}

// GetActionTypeSchema calls GET /v1/workflow/action-types/{type}/schema.
func (c *Client) GetActionTypeSchema(ctx context.Context, actionType string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/workflow/action-types/"+actionType+"/schema")
}

// GetFieldTypes calls GET /v1/config/form-field-types.
func (c *Client) GetFieldTypes(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/v1/config/form-field-types")
}

// GetFieldTypeSchema calls GET /v1/config/form-field-types/{type}/schema.
func (c *Client) GetFieldTypeSchema(ctx context.Context, fieldType string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/config/form-field-types/"+fieldType+"/schema")
}

// GetTableConfigSchema calls GET /v1/config/schemas/table-config.
func (c *Client) GetTableConfigSchema(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/v1/config/schemas/table-config")
}

// GetLayoutSchema calls GET /v1/config/schemas/layout.
func (c *Client) GetLayoutSchema(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/v1/config/schemas/layout")
}

// GetContentTypes calls GET /v1/config/schemas/content-types.
func (c *Client) GetContentTypes(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/v1/config/schemas/content-types")
}

// GetTriggerTypes calls GET /v1/workflow/trigger-types.
func (c *Client) GetTriggerTypes(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/v1/workflow/trigger-types")
}

// GetEntityTypes calls GET /v1/workflow/entity-types.
func (c *Client) GetEntityTypes(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/v1/workflow/entity-types")
}

// GetEntities calls GET /v1/workflow/entities.
func (c *Client) GetEntities(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/v1/workflow/entities")
}

// GetWorkflowRules calls GET /v1/workflow/rules.
func (c *Client) GetWorkflowRules(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/v1/workflow/rules")
}

// GetWorkflowRule calls GET /v1/workflow/rules/{id}.
func (c *Client) GetWorkflowRule(ctx context.Context, id string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/workflow/rules/"+id)
}

// GetWorkflowRuleActions calls GET /v1/workflow/rules/{id}/actions.
func (c *Client) GetWorkflowRuleActions(ctx context.Context, ruleID string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/workflow/rules/"+ruleID+"/actions")
}

// GetWorkflowRuleEdges calls GET /v1/workflow/rules/{ruleID}/edges.
func (c *Client) GetWorkflowRuleEdges(ctx context.Context, ruleID string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/workflow/rules/"+ruleID+"/edges")
}

// GetPageConfigs calls GET /v1/config/page-configs/all.
func (c *Client) GetPageConfigs(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/v1/config/page-configs/all")
}

// GetPageConfig calls GET /v1/config/page-configs/id/{config_id}.
func (c *Client) GetPageConfig(ctx context.Context, id string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/config/page-configs/id/"+id)
}

// GetPageConfigByName calls GET /v1/config/page-configs/name/{name}.
func (c *Client) GetPageConfigByName(ctx context.Context, name string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/config/page-configs/name/"+name)
}

// GetPageContent calls GET /v1/config/page-configs/content/{page_config_id}.
func (c *Client) GetPageContent(ctx context.Context, pageConfigID string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/config/page-configs/content/"+pageConfigID)
}

// GetForms calls GET /v1/config/forms.
func (c *Client) GetForms(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/v1/config/forms")
}

// GetFormFull calls GET /v1/config/forms/{form_id}/full.
func (c *Client) GetFormFull(ctx context.Context, formID string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/config/forms/"+formID+"/full")
}

// GetFormByNameFull calls GET /v1/config/forms/name/{form_name}/full.
func (c *Client) GetFormByNameFull(ctx context.Context, formName string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/config/forms/name/"+formName+"/full")
}

// GetTableConfigs calls GET /v1/data/configs/all.
func (c *Client) GetTableConfigs(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/v1/data/configs/all")
}

// GetTableConfig calls GET /v1/data/id/{table_config_id}.
func (c *Client) GetTableConfig(ctx context.Context, id string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/data/id/"+id)
}

// GetTableConfigByName calls GET /v1/data/name/{name}.
func (c *Client) GetTableConfigByName(ctx context.Context, name string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/data/name/"+name)
}

// GetSchemas calls GET /v1/introspection/schemas.
func (c *Client) GetSchemas(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/v1/introspection/schemas")
}

// GetTables calls GET /v1/introspection/schemas/{schema}/tables.
func (c *Client) GetTables(ctx context.Context, schema string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/introspection/schemas/"+schema+"/tables")
}

// GetColumns calls GET /v1/introspection/tables/{schema}/{table}/columns.
func (c *Client) GetColumns(ctx context.Context, schema, table string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/introspection/tables/"+schema+"/"+table+"/columns")
}

// GetRelationships calls GET /v1/introspection/tables/{schema}/{table}/relationships.
func (c *Client) GetRelationships(ctx context.Context, schema, table string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/introspection/tables/"+schema+"/"+table+"/relationships")
}

// GetEnumTypes calls GET /v1/introspection/enums/{schema}.
func (c *Client) GetEnumTypes(ctx context.Context, schema string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/introspection/enums/"+schema)
}

// GetEnumValues calls GET /v1/introspection/enums/{schema}/{name}.
func (c *Client) GetEnumValues(ctx context.Context, schema, name string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/introspection/enums/"+schema+"/"+name)
}

// GetEnumOptions calls GET /v1/config/enums/{schema}/{name}/options.
func (c *Client) GetEnumOptions(ctx context.Context, schema, name string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/config/enums/"+schema+"/"+name+"/options")
}

// GetPageActionTypes calls GET /v1/config/schemas/page-action-types.
func (c *Client) GetPageActionTypes(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/v1/config/schemas/page-action-types")
}

// GetPageActions calls GET /v1/config/page-configs/actions/{page_config_id}.
func (c *Client) GetPageActions(ctx context.Context, pageConfigID string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/config/page-configs/actions/"+pageConfigID)
}

// GetPageAction calls GET /v1/config/page-actions/{action_id}.
func (c *Client) GetPageAction(ctx context.Context, actionID string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/config/page-actions/"+actionID)
}

// GetTemplates calls GET /v1/workflow/templates.
func (c *Client) GetTemplates(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/v1/workflow/templates")
}

// GetActiveTemplates calls GET /v1/workflow/templates/active.
func (c *Client) GetActiveTemplates(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/v1/workflow/templates/active")
}

// =========================================================================
// Write Methods
// =========================================================================

// CreateWorkflow calls POST /v1/workflow/rules/full.
func (c *Client) CreateWorkflow(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	return c.post(ctx, "/v1/workflow/rules/full", payload)
}

// UpdateWorkflow calls PUT /v1/workflow/rules/{id}/full.
func (c *Client) UpdateWorkflow(ctx context.Context, id string, payload json.RawMessage) (json.RawMessage, error) {
	return c.put(ctx, "/v1/workflow/rules/"+id+"/full", payload)
}

// ValidateWorkflow calls POST /v1/workflow/rules/full?dry_run=true.
func (c *Client) ValidateWorkflow(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	return c.post(ctx, "/v1/workflow/rules/full?dry_run=true", payload)
}

// CreatePageConfig calls POST /v1/config/page-configs.
func (c *Client) CreatePageConfig(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	return c.post(ctx, "/v1/config/page-configs", payload)
}

// UpdatePageConfig calls PUT /v1/config/page-configs/id/{config_id}.
func (c *Client) UpdatePageConfig(ctx context.Context, id string, payload json.RawMessage) (json.RawMessage, error) {
	return c.put(ctx, "/v1/config/page-configs/id/"+id, payload)
}

// CreatePageContent calls POST /v1/config/page-content.
func (c *Client) CreatePageContent(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	return c.post(ctx, "/v1/config/page-content", payload)
}

// UpdatePageContent calls PUT /v1/config/page-content/{content_id}.
func (c *Client) UpdatePageContent(ctx context.Context, id string, payload json.RawMessage) (json.RawMessage, error) {
	return c.put(ctx, "/v1/config/page-content/"+id, payload)
}

// CreateForm calls POST /v1/config/forms.
func (c *Client) CreateForm(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	return c.post(ctx, "/v1/config/forms", payload)
}

// UpdateForm calls PUT /v1/config/forms/{form_id}.
func (c *Client) UpdateForm(ctx context.Context, id string, payload json.RawMessage) (json.RawMessage, error) {
	return c.put(ctx, "/v1/config/forms/"+id, payload)
}

// CreateFormField calls POST /v1/config/form-fields.
func (c *Client) CreateFormField(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	return c.post(ctx, "/v1/config/form-fields", payload)
}

// UpdateFormField calls PUT /v1/config/form-fields/{field_id}.
func (c *Client) UpdateFormField(ctx context.Context, id string, payload json.RawMessage) (json.RawMessage, error) {
	return c.put(ctx, "/v1/config/form-fields/"+id, payload)
}

// CreateTableConfig calls POST /v1/data.
func (c *Client) CreateTableConfig(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	return c.post(ctx, "/v1/data", payload)
}

// UpdateTableConfig calls PUT /v1/data/{table_config_id}.
func (c *Client) UpdateTableConfig(ctx context.Context, id string, payload json.RawMessage) (json.RawMessage, error) {
	return c.put(ctx, "/v1/data/"+id, payload)
}

// ValidateTableConfig calls POST /v1/data/validate.
func (c *Client) ValidateTableConfig(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	return c.post(ctx, "/v1/data/validate", payload)
}
