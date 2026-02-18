package communication

import (
	"fmt"

	"github.com/resend/resend-go/v2"
)

// EmailClient defines the interface for sending emails.
// Using an interface allows test doubles without a live email service.
type EmailClient interface {
	Send(from string, to []string, subject, body string) (string, error)
}

// ResendConfig holds Resend API configuration.
type ResendConfig struct {
	APIKey string
	From   string
}

// ResendEmailClient implements EmailClient using the Resend API.
type ResendEmailClient struct {
	client *resend.Client
	from   string
}

// NewResendEmailClient creates a new Resend email client.
// Returns nil if config is incomplete (graceful degradation — handler will log and skip).
func NewResendEmailClient(config ResendConfig) *ResendEmailClient {
	if config.APIKey == "" || config.From == "" {
		return nil
	}
	return &ResendEmailClient{
		client: resend.NewClient(config.APIKey),
		from:   config.From,
	}
}

// Send delivers an email via the Resend API.
// Returns the Resend email ID on success.
func (c *ResendEmailClient) Send(from string, to []string, subject, body string) (string, error) {
	if from == "" {
		from = c.from
	}

	params := &resend.SendEmailRequest{
		From:    from,
		To:      to,
		Subject: subject,
		Text:    body,
	}

	resp, err := c.client.Emails.Send(params)
	if err != nil {
		return "", fmt.Errorf("resend send: %w", err)
	}

	return resp.Id, nil
}

// MockEmailClient is a test double for EmailClient.
// It records calls without making real API requests.
// Exported so it can be used by tests in other packages (e.g. integration tests).
type MockEmailClient struct {
	// SendCalls captures all calls to Send, in order.
	SendCalls []MockEmailSend

	// SendErr, if non-nil, is returned from every Send call.
	SendErr error

	// IDPrefix is prepended to the call index to generate mock email IDs.
	// Defaults to "mock-email-".
	IDPrefix string

	callCount int
}

// MockEmailSend records a single Send invocation.
type MockEmailSend struct {
	From    string
	To      []string
	Subject string
	Body    string
}

// Send records the call and returns either an error or a synthetic email ID.
func (m *MockEmailClient) Send(from string, to []string, subject, body string) (string, error) {
	m.SendCalls = append(m.SendCalls, MockEmailSend{
		From:    from,
		To:      to,
		Subject: subject,
		Body:    body,
	})

	if m.SendErr != nil {
		return "", m.SendErr
	}

	prefix := m.IDPrefix
	if prefix == "" {
		prefix = "mock-email-"
	}
	m.callCount++
	return fmt.Sprintf("%s%d", prefix, m.callCount), nil
}

// Reset clears all recorded calls and resets the call count.
func (m *MockEmailClient) Reset() {
	m.SendCalls = nil
	m.callCount = 0
}

// CallCount returns the number of Send calls made.
func (m *MockEmailClient) CallCount() int {
	return len(m.SendCalls)
}
