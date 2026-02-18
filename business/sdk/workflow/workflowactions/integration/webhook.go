package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// templateVarPattern matches {{variable_name}} patterns for template substitution.
var templateVarPattern = regexp.MustCompile(`\{\{([^}]+)\}\}`)

// WebhookConfig represents the configuration for a call_webhook action.
type WebhookConfig struct {
	URL                 string            `json:"url"`
	Method              string            `json:"method"`
	Headers             map[string]string `json:"headers,omitempty"`
	Body                string            `json:"body,omitempty"`
	TimeoutSeconds      int               `json:"timeout_seconds,omitempty"`
	ExpectedStatusCodes []int             `json:"expected_status_codes,omitempty"`
}

// CallWebhookHandler handles call_webhook actions by making outbound HTTP requests.
type CallWebhookHandler struct {
	log        *logger.Logger
	httpClient *http.Client
}

// NewCallWebhookHandler creates a new webhook handler.
func NewCallWebhookHandler(log *logger.Logger) *CallWebhookHandler {
	return &CallWebhookHandler{
		log: log,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewCallWebhookHandlerWithClient creates a webhook handler with a custom HTTP client (for testing).
func NewCallWebhookHandlerWithClient(log *logger.Logger, client *http.Client) *CallWebhookHandler {
	return &CallWebhookHandler{
		log:        log,
		httpClient: client,
	}
}

func (h *CallWebhookHandler) GetType() string              { return "call_webhook" }
func (h *CallWebhookHandler) SupportsManualExecution() bool { return true }
func (h *CallWebhookHandler) IsAsync() bool                 { return false }

func (h *CallWebhookHandler) GetDescription() string {
	return "Make an outbound HTTP request to a webhook URL"
}

func (h *CallWebhookHandler) GetOutputPorts() []workflow.OutputPort {
	return []workflow.OutputPort{
		{Name: "success", Description: "HTTP request succeeded with expected status code", IsDefault: true},
		{Name: "client_error", Description: "HTTP 4xx response"},
		{Name: "server_error", Description: "HTTP 5xx response"},
		{Name: "unexpected_response", Description: "Response code not in expected_status_codes (non-error)"},
		{Name: "timeout", Description: "Request timed out"},
		{Name: "failure", Description: "Request could not be sent"},
	}
}

func (h *CallWebhookHandler) Validate(config json.RawMessage) error {
	var cfg WebhookConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if cfg.URL == "" {
		return fmt.Errorf("url is required")
	}

	parsed, err := url.Parse(cfg.URL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Template variables in the URL may make scheme detection unreliable,
	// so only enforce HTTPS when the scheme is clearly not https and not templated.
	if !strings.Contains(cfg.URL, "{{") {
		if parsed.Scheme != "https" && !isInternalHost(parsed.Host) {
			return fmt.Errorf("only https URLs are allowed for external webhooks (got %s)", parsed.Scheme)
		}
	}

	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
	}
	method := strings.ToUpper(cfg.Method)
	if method == "" {
		method = "POST"
	}
	if !validMethods[method] {
		return fmt.Errorf("invalid method %q, must be GET/POST/PUT/PATCH/DELETE", cfg.Method)
	}

	if cfg.TimeoutSeconds > 120 {
		return fmt.Errorf("timeout_seconds must be <= 120")
	}

	return nil
}

func (h *CallWebhookHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
	var cfg WebhookConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return buildWebhookResult("failure", 0, ""), fmt.Errorf("parse webhook config: %w", err)
	}

	// Resolve template variables in URL, body, and headers.
	resolvedURL := resolveTemplateVars(cfg.URL, execCtx.RawData)
	resolvedBody := resolveTemplateVars(cfg.Body, execCtx.RawData)

	// Apply per-request timeout.
	timeout := 30 * time.Second
	if cfg.TimeoutSeconds > 0 {
		timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	}
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build HTTP request.
	method := strings.ToUpper(cfg.Method)
	if method == "" {
		method = "POST"
	}

	var bodyReader io.Reader
	if resolvedBody != "" {
		bodyReader = bytes.NewBufferString(resolvedBody)
	}

	req, err := http.NewRequestWithContext(reqCtx, method, resolvedURL, bodyReader)
	if err != nil {
		return buildWebhookResult("failure", 0, ""), fmt.Errorf("build request: %w", err)
	}

	// Set headers with template resolution.
	for key, value := range cfg.Headers {
		req.Header.Set(key, resolveTemplateVars(value, execCtx.RawData))
	}
	if resolvedBody != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request.
	resp, err := h.httpClient.Do(req)
	if err != nil {
		if reqCtx.Err() != nil {
			return buildWebhookResult("timeout", 0, ""), nil
		}
		return buildWebhookResult("failure", 0, ""), fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body (truncate to 10KB).
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 10*1024))

	// Determine output port based on status code.
	expectedCodes := cfg.ExpectedStatusCodes
	if len(expectedCodes) == 0 {
		expectedCodes = []int{200, 201, 202, 204}
	}
	output := determineWebhookOutput(resp.StatusCode, expectedCodes)

	if h.log != nil {
		h.log.Info(ctx, "call_webhook executed",
			"url", resolvedURL, "method", method,
			"status", resp.StatusCode, "output", output)
	}

	result := buildWebhookResult(output, resp.StatusCode, string(respBody))
	result["content_type"] = resp.Header.Get("Content-Type")

	return result, nil
}

// determineWebhookOutput maps an HTTP status code to a named output port.
func determineWebhookOutput(statusCode int, expectedCodes []int) string {
	if slices.Contains(expectedCodes, statusCode) {
		return "success"
	}
	if statusCode >= 500 {
		return "server_error"
	}
	if statusCode >= 400 {
		return "client_error"
	}
	return "unexpected_response"
}

// buildWebhookResult constructs the standard result map returned by Execute.
func buildWebhookResult(output string, statusCode int, body string) map[string]any {
	return map[string]any{
		"output":      output,
		"status_code": statusCode,
		"body":        body,
	}
}

// privateRanges holds CIDR blocks for all private/loopback/link-local addresses.
var privateRanges []*net.IPNet

func init() {
	cidrs := []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC 1918
		"172.16.0.0/12",  // RFC 1918 (covers 172.16–172.31)
		"192.168.0.0/16", // RFC 1918
		"169.254.0.0/16", // link-local
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 ULA
		"fe80::/10",      // IPv6 link-local
	}
	for _, cidr := range cidrs {
		_, block, _ := net.ParseCIDR(cidr)
		privateRanges = append(privateRanges, block)
	}
}

// isInternalHost returns true for localhost, private, and link-local addresses.
func isInternalHost(host string) bool {
	h := host

	// Handle bracketed IPv6 like [::1]:8080.
	if strings.HasPrefix(h, "[") {
		if close := strings.Index(h, "]"); close != -1 {
			h = h[1:close]
		}
	} else if idx := strings.LastIndex(h, ":"); idx != -1 && !strings.Contains(h, "::") {
		// Strip port for IPv4 or hostname:port (but not bare IPv6 like ::1).
		h = h[:idx]
	}

	if h == "localhost" {
		return true
	}

	ip := net.ParseIP(h)
	if ip == nil {
		return false
	}

	for _, block := range privateRanges {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

// resolveTemplateVars replaces {{variable_name}} patterns with values from the data map.
func resolveTemplateVars(template string, data map[string]any) string {
	if data == nil {
		return template
	}

	return templateVarPattern.ReplaceAllStringFunc(template, func(match string) string {
		varName := match[2 : len(match)-2]
		if value, ok := data[varName]; ok {
			return fmt.Sprintf("%v", value)
		}
		return match
	})
}
