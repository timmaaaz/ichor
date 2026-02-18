# Phase 4: Implement send_email SMTP

**Category**: Backend
**Status**: Pending
**Dependencies**: None
**Effort**: Medium

---

## Overview

`send_email` is a stub that logs intent and returns a fake `email_id`. This phase implements real SMTP delivery using Go's stdlib `net/smtp` package (no new dependencies). SMTP credentials are configured via `ICHOR_SMTP_*` environment variables following the project's `ICHOR_*` env var convention.

The handler is already marked `IsAsync: true` — this is correct, as email delivery should not block the Temporal workflow. However, the current `AsyncActivityHandler.StartAsync` method is not implemented. For this phase, we'll implement sync delivery in `Execute()` (simpler, sufficient for most use cases). The `IsAsync` flag can stay as-is since Temporal activities have their own retry and timeout semantics.

---

## Goals

1. Add SMTP config to the application config struct
2. Create an `SMTPClient` interface + `net/smtp` implementation
3. Update `SendEmailHandler` to use the client
4. Support `{{template_vars}}` in subject and body

---

## Task Breakdown

### Task 1: Add SMTP Config

**File**: `api/cmd/services/ichor/main.go`

Find the `cfg` struct and add an SMTP section following the existing pattern:

```go
SMTP struct {
    Host     string `conf:"default:"`
    Port     int    `conf:"default:587"`
    Username string `conf:"default:"`
    Password string `conf:"default:,mask"` // mask prevents logging
    From     string `conf:"default:"`
    TLS      bool   `conf:"default:true"`
} `conf:""`
```

The prefix `ICHOR_` is added automatically by the conf library, so env vars become `ICHOR_SMTP_HOST`, `ICHOR_SMTP_PORT`, etc.

**File**: `zarf/k8s/dev/ichor/ichor.yaml` — add the SMTP env vars as comments or with empty defaults (no real credentials in git).

### Task 2: Create SMTPClient Interface and Implementation

**New file**: `business/sdk/workflow/workflowactions/communication/smtp.go`

```go
package communication

import (
    "crypto/tls"
    "fmt"
    "net"
    "net/smtp"
    "strings"
)

// SMTPClient defines the interface for sending emails.
// Using an interface allows test doubles without an SMTP server.
type SMTPClient interface {
    Send(from string, to []string, subject, body string) error
}

// SMTPConfig holds SMTP server configuration.
type SMTPConfig struct {
    Host     string
    Port     int
    Username string
    Password string
    From     string
    TLS      bool
}

// NetSMTPClient implements SMTPClient using Go's net/smtp stdlib.
type NetSMTPClient struct {
    config SMTPConfig
}

// NewNetSMTPClient creates a new SMTP client.
// Returns nil if config is incomplete (graceful degradation).
func NewNetSMTPClient(config SMTPConfig) *NetSMTPClient {
    if config.Host == "" || config.From == "" {
        return nil
    }
    return &NetSMTPClient{config: config}
}

func (c *NetSMTPClient) Send(from string, to []string, subject, body string) error {
    addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
    auth := smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)

    headers := fmt.Sprintf(
        "From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n",
        from, strings.Join(to, ", "), subject,
    )
    message := []byte(headers + body)

    if c.config.TLS {
        tlsConfig := &tls.Config{ServerName: c.config.Host}
        conn, err := tls.Dial("tcp", addr, tlsConfig)
        if err != nil {
            return fmt.Errorf("smtp tls dial: %w", err)
        }
        client, err := smtp.NewClient(conn, c.config.Host)
        if err != nil {
            return fmt.Errorf("smtp new client: %w", err)
        }
        defer client.Close()
        if err := client.Auth(auth); err != nil {
            return fmt.Errorf("smtp auth: %w", err)
        }
        if err := client.Mail(from); err != nil {
            return fmt.Errorf("smtp mail from: %w", err)
        }
        for _, recipient := range to {
            if err := client.Rcpt(recipient); err != nil {
                return fmt.Errorf("smtp rcpt to %s: %w", recipient, err)
            }
        }
        w, err := client.Data()
        if err != nil {
            return fmt.Errorf("smtp data: %w", err)
        }
        if _, err := w.Write(message); err != nil {
            return fmt.Errorf("smtp write: %w", err)
        }
        return w.Close()
    }

    // STARTTLS / plain
    return smtp.SendMail(addr, auth, from, to, message)
}
```

### Task 3: Update SendEmailHandler

**File**: `business/sdk/workflow/workflowactions/communication/email.go`

```go
type SendEmailHandler struct {
    log        *logger.Logger
    db         *sqlx.DB
    smtpClient SMTPClient // NEW
    smtpFrom   string     // NEW: sender address
}

func NewSendEmailHandler(log *logger.Logger, db *sqlx.DB, smtpClient SMTPClient, smtpFrom string) *SendEmailHandler {
    return &SendEmailHandler{
        log:        log,
        db:         db,
        smtpClient: smtpClient,
        smtpFrom:   smtpFrom,
    }
}
```

Update `Execute()`:

```go
func (h *SendEmailHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (interface{}, error) {
    var cfg struct {
        Recipients      []string `json:"recipients"`
        Subject         string   `json:"subject"`
        Body            string   `json:"body"`
        SimulateFailure bool     `json:"simulate_failure,omitempty"`
        FailureMessage  string   `json:"failure_message,omitempty"`
    }
    if err := json.Unmarshal(config, &cfg); err != nil {
        return nil, fmt.Errorf("failed to parse email config: %w", err)
    }

    if cfg.SimulateFailure {
        // ... keep existing simulate_failure logic ...
    }

    // Resolve template variables
    subject := resolveTemplateVars(cfg.Subject, execCtx.RawData)
    body := resolveTemplateVars(cfg.Body, execCtx.RawData)

    emailID := uuid.New()
    now := time.Now()

    if h.smtpClient != nil {
        if err := h.smtpClient.Send(h.smtpFrom, cfg.Recipients, subject, body); err != nil {
            return nil, fmt.Errorf("send email: %w", err)
        }
    } else {
        h.log.Warn(ctx, "send_email: no SMTP client configured, skipping delivery",
            "recipients", cfg.Recipients, "subject", subject)
    }

    h.log.Info(ctx, "send_email executed",
        "email_id", emailID, "recipients", cfg.Recipients, "subject", subject)

    return map[string]interface{}{
        "email_id":   emailID.String(),
        "status":     "sent",
        "sent_at":    now.Format(time.RFC3339),
        "recipients": cfg.Recipients,
        "subject":    subject,
    }, nil
}
```

### Task 4: Update register.go

**File**: `business/sdk/workflow/workflowactions/register.go`

Add `SMTPClient` and `SMTPFrom` to `ActionConfig`:

```go
type ActionConfig struct {
    Log         *logger.Logger
    DB          *sqlx.DB
    QueueClient *rabbitmq.WorkflowQueue
    SMTPClient  communication.SMTPClient // ADD
    SMTPFrom    string                   // ADD

    Buses BusDependencies
}
```

In `RegisterAll` and `RegisterCoreActions`:
```go
registry.Register(communication.NewSendEmailHandler(config.Log, config.DB, config.SMTPClient, config.SMTPFrom))
```

### Task 5: Wire in all.go

**File**: `api/cmd/services/ichor/build/all/all.go`

Create the SMTP client from config and pass to ActionConfig:

```go
smtpClient := communication.NewNetSMTPClient(communication.SMTPConfig{
    Host:     cfg.SMTP.Host,
    Port:     cfg.SMTP.Port,
    Username: cfg.SMTP.Username,
    Password: cfg.SMTP.Password,
    From:     cfg.SMTP.From,
    TLS:      cfg.SMTP.TLS,
})

actionConfig := workflowactions.ActionConfig{
    Log:         log,
    DB:          db,
    QueueClient: workflowQueue,
    SMTPClient:  smtpClient,  // nil if Host is empty
    SMTPFrom:    cfg.SMTP.From,
    Buses: workflowactions.BusDependencies{...},
}
```

---

## Testing

Test with [mailhog](https://github.com/mailhog/MailHog) or [mailpit](https://github.com/axllent/mailpit) running locally:
```bash
ICHOR_SMTP_HOST=localhost ICHOR_SMTP_PORT=1025 ICHOR_SMTP_TLS=false make run
```

---

## Validation

```bash
go build ./...

# Check communication package imports
grep -r "net/smtp" business/sdk/workflow/workflowactions/communication/

# Check config struct
grep -A 8 "SMTP struct" api/cmd/services/ichor/main.go
```

---

## Gotchas

- **`communication` package import in register.go** — the `SMTPClient` interface is in the `communication` package. Use a type alias or define the interface in a shared location to avoid circular imports. Alternatively, keep `SMTPClient` defined in `register.go` itself as a local interface with just `Send(...)`.
- **Nil SMTP client is graceful** — if `ICHOR_SMTP_HOST` is empty, `NewNetSMTPClient` returns nil. The handler logs a warning and skips delivery. This allows the service to start without SMTP configured.
- **`db` field may become unused** — same as `send_notification`. Keep for future features (delivery history).
- **HTML email** — not in scope for this phase. Plain text only. Add HTML support as a future enhancement via a `content_type` config field.
