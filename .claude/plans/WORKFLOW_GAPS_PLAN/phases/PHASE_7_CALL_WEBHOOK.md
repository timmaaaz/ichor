# Phase 8: Add call_webhook Action

**Category**: Backend
**Status**: Pending
**Dependencies**: None
**Effort**: Medium

---

## Overview

Currently there is no way for workflows to send data to external systems. This phase adds a `call_webhook` action that makes an outbound HTTP request with a configurable URL, method, headers, and body — all supporting `{{template_vars}}` substitution.

Common ERP use cases:
- Notify a shipping carrier when an order is dispatched
- Push an invoice to an accounting system (QuickBooks/Xero)
- Trigger a third-party logistics system when inventory reaches a threshold
- Send order data to a customer's system

---

## Goals

1. New `call_webhook` action handler in a new `integration/` subpackage
2. URL, headers, and body support template variable substitution
3. Fine-grained output ports based on HTTP response status codes
4. HTTPS-only enforcement (security)
5. Configurable timeout

---

## Task Breakdown

### Task 1: Create integration/ subpackage and Handler

**New file**: `business/sdk/workflow/workflowactions/integration/webhook.go`

```go
package integration

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "github.com/timmaaaz/ichor/foundation/logger"
)

type CallWebhookHandler struct {
    log        *logger.Logger
    httpClient *http.Client // injectable for testing
}

func NewCallWebhookHandler(log *logger.Logger) *CallWebhookHandler {
    return &CallWebhookHandler{
        log: log,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}
```

**Config struct**:
```go
type WebhookConfig struct {
    URL                  string            `json:"url"`
    Method               string            `json:"method"`                  // GET, POST, PUT, PATCH
    Headers              map[string]string `json:"headers,omitempty"`
    Body                 string            `json:"body,omitempty"`          // JSON string, supports {{vars}}
    TimeoutSeconds       int               `json:"timeout_seconds,omitempty"` // default 30, max 120
    ExpectedStatusCodes  []int             `json:"expected_status_codes,omitempty"` // default [200, 201, 202, 204]
}
```

**Validate()**:
```go
func (h *CallWebhookHandler) Validate(config json.RawMessage) error {
    var cfg WebhookConfig
    if err := json.Unmarshal(config, &cfg); err != nil {
        return fmt.Errorf("invalid config: %w", err)
    }
    if cfg.URL == "" {
        return fmt.Errorf("url is required")
    }

    // Security: enforce HTTPS (allow http for localhost/internal only)
    parsed, err := url.Parse(cfg.URL)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }
    if parsed.Scheme != "https" && !isInternalHost(parsed.Host) {
        return fmt.Errorf("only https URLs are allowed for external webhooks (got %s)", parsed.Scheme)
    }

    validMethods := map[string]bool{"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true}
    if !validMethods[strings.ToUpper(cfg.Method)] {
        return fmt.Errorf("invalid method %q, must be GET/POST/PUT/PATCH/DELETE", cfg.Method)
    }
    if cfg.TimeoutSeconds > 120 {
        return fmt.Errorf("timeout_seconds must be <= 120")
    }
    return nil
}

func isInternalHost(host string) bool {
    return strings.HasPrefix(host, "localhost") ||
        strings.HasPrefix(host, "127.") ||
        strings.HasPrefix(host, "10.") ||
        strings.HasPrefix(host, "192.168.")
}
```

**Execute()**:
```go
func (h *CallWebhookHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (interface{}, error) {
    var cfg WebhookConfig
    if err := json.Unmarshal(config, &cfg); err != nil {
        return nil, fmt.Errorf("parse webhook config: %w", err)
    }

    // Resolve template variables
    resolvedURL := resolveTemplateVars(cfg.URL, execCtx.RawData)
    resolvedBody := resolveTemplateVars(cfg.Body, execCtx.RawData)

    // Apply timeout
    timeout := time.Duration(30) * time.Second
    if cfg.TimeoutSeconds > 0 {
        timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
    }
    reqCtx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    // Build request
    var bodyReader io.Reader
    if resolvedBody != "" {
        bodyReader = bytes.NewBufferString(resolvedBody)
    }
    req, err := http.NewRequestWithContext(reqCtx, strings.ToUpper(cfg.Method), resolvedURL, bodyReader)
    if err != nil {
        return buildWebhookResult("failure", 0, "", nil), fmt.Errorf("build request: %w", err)
    }

    // Set headers (with template resolution)
    for key, value := range cfg.Headers {
        req.Header.Set(key, resolveTemplateVars(value, execCtx.RawData))
    }
    if resolvedBody != "" && req.Header.Get("Content-Type") == "" {
        req.Header.Set("Content-Type", "application/json")
    }

    // Execute request
    resp, err := h.httpClient.Do(req)
    if err != nil {
        if ctx.Err() != nil {
            return buildWebhookResult("timeout", 0, "", nil), nil // route to timeout port, not error
        }
        return buildWebhookResult("failure", 0, "", nil), fmt.Errorf("http request: %w", err)
    }
    defer resp.Body.Close()

    // Read response body (truncate to 10KB)
    respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 10*1024))

    // Determine output port
    expectedCodes := cfg.ExpectedStatusCodes
    if len(expectedCodes) == 0 {
        expectedCodes = []int{200, 201, 202, 204}
    }
    output := determineWebhookOutput(resp.StatusCode, expectedCodes)

    h.log.Info(ctx, "call_webhook executed",
        "url", resolvedURL, "method", cfg.Method,
        "status", resp.StatusCode, "output", output)

    return buildWebhookResult(output, resp.StatusCode, string(respBody), resp.Header), nil
}

func determineWebhookOutput(statusCode int, expectedCodes []int) string {
    for _, code := range expectedCodes {
        if statusCode == code {
            return "success"
        }
    }
    if statusCode >= 400 && statusCode < 500 {
        return "client_error"
    }
    if statusCode >= 500 {
        return "server_error"
    }
    return "success" // 2xx/3xx not in expected list but still "success"
}

func buildWebhookResult(output string, statusCode int, body string, headers http.Header) map[string]interface{} {
    result := map[string]interface{}{
        "output":      output,
        "status_code": statusCode,
        "body":        body,
    }
    if headers != nil {
        result["content_type"] = headers.Get("Content-Type")
    }
    return result
}
```

**Output ports**:
```go
func (h *CallWebhookHandler) GetOutputPorts() []workflow.OutputPort {
    return []workflow.OutputPort{
        {Name: "success", Description: "HTTP request succeeded with expected status code", IsDefault: true},
        {Name: "client_error", Description: "HTTP 4xx response"},
        {Name: "server_error", Description: "HTTP 5xx response"},
        {Name: "timeout", Description: "Request timed out"},
        {Name: "failure", Description: "Request could not be sent"},
    }
}
```

**Other methods**:
```go
func (h *CallWebhookHandler) GetType() string              { return "call_webhook" }
func (h *CallWebhookHandler) SupportsManualExecution() bool { return true }
func (h *CallWebhookHandler) IsAsync() bool                 { return false }
func (h *CallWebhookHandler) GetDescription() string {
    return "Make an outbound HTTP request to a webhook URL"
}
```

Note: `resolveTemplateVars` needs to be either copied into this package or the communication package's function needs to be moved to a shared location. Consider moving it to `business/sdk/workflow/template.go` as a package-level function.

### Task 2: Register Handler

**File**: `business/sdk/workflow/workflowactions/register.go`

```go
// Integration actions
registry.Register(integration.NewCallWebhookHandler(config.Log))
```

Add to `RegisterAll` only (not `RegisterCoreActions` — network calls not appropriate for test environments).

---

## Validation

```bash
go build ./...

# Test with a local echo server or httpbin
curl -X POST localhost:8080/workflow/execute \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"action_type": "call_webhook", "config": {"url": "https://httpbin.org/post", "method": "POST", "body": "{\"entity_id\": \"{{entity_id}}\"}"}}'
```

---

## Gotchas

- **`resolveTemplateVars` accessibility**: The function is currently in the `communication` package. Either copy it into `integration/` (duplication but no coupling) or move it to `business/sdk/workflow/` as a shared utility. The template processor in `template.go` does something similar but is a full struct. A standalone function is simpler for this use case.
- **SSRF (Server-Side Request Forgery)**: The webhook handler could be used to probe internal services. The `isInternalHost` check helps but is not a complete SSRF defense. Consider adding a blocklist for cloud metadata endpoints (`169.254.169.254`, etc.) if this is a security concern.
- **Response body size**: Limit to 10KB (`io.LimitReader`) to prevent memory exhaustion from large responses.
- **Secrets in headers**: If users put API keys in `headers`, these will appear in workflow config (stored in DB). Document that secrets should be stored separately and reference via `{{variable}}` from context. For now, this is acceptable — it's no worse than other config fields.
- **Temporal retry behavior**: Temporal retries failed activities by default (3 attempts). For webhooks, this could cause duplicate deliveries. The handler should either be idempotent by design (recommended: document this) or set `MaximumAttempts: 1` in `activityOptions()` for this handler type.
