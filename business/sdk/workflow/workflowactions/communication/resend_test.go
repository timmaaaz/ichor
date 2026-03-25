package communication_test

import (
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
)

func TestNewResendEmailClient_EmptyAPIKey(t *testing.T) {
	client := communication.NewResendEmailClient(communication.ResendConfig{
		APIKey: "",
		From:   "test@example.com",
	})
	if client != nil {
		t.Fatal("expected nil for empty API key")
	}
}

func TestNewResendEmailClient_EmptyFrom(t *testing.T) {
	client := communication.NewResendEmailClient(communication.ResendConfig{
		APIKey: "re_test_key",
		From:   "",
	})
	if client != nil {
		t.Fatal("expected nil for empty From")
	}
}

func TestNewResendEmailClient_BothEmpty(t *testing.T) {
	client := communication.NewResendEmailClient(communication.ResendConfig{})
	if client != nil {
		t.Fatal("expected nil for empty config")
	}
}

func TestNewResendEmailClient_ValidConfig(t *testing.T) {
	client := communication.NewResendEmailClient(communication.ResendConfig{
		APIKey: "re_test_key_12345",
		From:   "noreply@example.com",
	})
	if client == nil {
		t.Fatal("expected non-nil client for valid config")
	}
}
