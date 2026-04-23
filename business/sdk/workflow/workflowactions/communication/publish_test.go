package communication

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
)

// TestBuildAlertPayload_AllFields asserts every alertbus.Alert field that has
// a corresponding alertapi HTTP-visible JSON tag appears in the WS payload,
// so frontend subscribers receive the same shape from either channel.
func TestBuildAlertPayload_AllFields(t *testing.T) {
	created := time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 4, 23, 10, 5, 0, 0, time.UTC)
	expires := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)

	alert := alertbus.Alert{
		ID:               uuid.New(),
		AlertType:        "test_alert",
		Severity:         alertbus.SeverityHigh,
		Title:            "Test",
		Message:          "This is a test",
		Context:          json.RawMessage(`{"foo":"bar"}`),
		SourceEntityName: "orders",
		SourceEntityID:   uuid.New(),
		SourceRuleID:     uuid.New(),
		SourceRuleName:   "Rule 1",
		ActionURL:        "/orders/123",
		Status:           alertbus.StatusActive,
		ExpiresDate:      &expires,
		CreatedDate:      created,
		UpdatedDate:      updated,
	}

	got := BuildAlertPayload(alert)

	// Every HTTP-visible field must appear in the WS payload.
	wantKeys := []string{
		"id", "alertType", "severity", "title", "message",
		"context", "sourceEntityName", "sourceEntityId",
		"sourceRuleId", "sourceRuleName", "actionUrl",
		"status", "expiresDate", "createdDate", "updatedDate",
	}
	for _, key := range wantKeys {
		if _, ok := got[key]; !ok {
			t.Errorf("payload missing key %q — frontend subscriber will see undefined", key)
		}
	}

	if got["id"] != alert.ID.String() {
		t.Errorf("id = %v, want %s", got["id"], alert.ID.String())
	}
	if got["sourceRuleId"] != alert.SourceRuleID.String() {
		t.Errorf("sourceRuleId = %v, want %s", got["sourceRuleId"], alert.SourceRuleID.String())
	}
	if got["actionUrl"] != alert.ActionURL {
		t.Errorf("actionUrl = %v, want %s", got["actionUrl"], alert.ActionURL)
	}
	if got["expiresDate"] != expires.Format(time.RFC3339) {
		t.Errorf("expiresDate = %v, want %s", got["expiresDate"], expires.Format(time.RFC3339))
	}
}

// TestBuildRecipientAlertMessage_User asserts user-targeted messages carry
// msg.UserID set and do NOT stuff the user UUID into the payload.
func TestBuildRecipientAlertMessage_User(t *testing.T) {
	alert := alertbus.Alert{ID: uuid.New()}
	alertData := map[string]interface{}{"id": alert.ID.String()}
	userID := uuid.New()
	recipient := alertbus.AlertRecipient{
		RecipientType: "user",
		RecipientID:   userID,
	}

	got := buildRecipientAlertMessage(alert, alertData, recipient)

	if got.Type != "alert" {
		t.Errorf("Type = %q, want %q", got.Type, "alert")
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.EntityID != alert.ID {
		t.Errorf("EntityID = %v, want %v", got.EntityID, alert.ID)
	}
	if _, ok := got.Payload["role_id"]; ok {
		t.Error("user-targeted message should not carry role_id in payload")
	}
	if _, ok := got.Payload["alert"]; !ok {
		t.Error("payload missing `alert` key")
	}
}

// TestBuildRecipientAlertMessage_Role asserts role-targeted messages leave
// msg.UserID zero and set role_id in the payload (the consumer reads it from there).
func TestBuildRecipientAlertMessage_Role(t *testing.T) {
	alert := alertbus.Alert{ID: uuid.New()}
	alertData := map[string]interface{}{"id": alert.ID.String()}
	roleID := uuid.New()
	recipient := alertbus.AlertRecipient{
		RecipientType: "role",
		RecipientID:   roleID,
	}

	got := buildRecipientAlertMessage(alert, alertData, recipient)

	if got.UserID != uuid.Nil {
		t.Errorf("UserID should be zero for role-targeted messages, got %v", got.UserID)
	}
	gotRoleID, ok := got.Payload["role_id"].(string)
	if !ok {
		t.Fatalf("payload.role_id missing or wrong type: %v", got.Payload["role_id"])
	}
	if gotRoleID != roleID.String() {
		t.Errorf("role_id = %q, want %q", gotRoleID, roleID.String())
	}
}

// TestBuildAlertPayload_OmitsZeroOptional asserts optional fields with zero
// values are absent (matches the `omitempty` JSON tag behavior of alertapi.Alert).
func TestBuildAlertPayload_OmitsZeroOptional(t *testing.T) {
	alert := alertbus.Alert{
		ID:          uuid.New(),
		AlertType:   "test_alert",
		Severity:    alertbus.SeverityLow,
		Title:       "Minimal",
		Message:     "Minimal alert",
		Status:      alertbus.StatusActive,
		CreatedDate: time.Now(),
		UpdatedDate: time.Now(),
		// Intentionally unset: Context, SourceEntityID, SourceRuleID, SourceEntityName,
		// SourceRuleName, ActionURL, ExpiresDate.
	}

	got := BuildAlertPayload(alert)

	mustAbsent := []string{"sourceEntityId", "sourceRuleId", "sourceEntityName", "sourceRuleName", "actionUrl", "expiresDate"}
	for _, key := range mustAbsent {
		if _, ok := got[key]; ok {
			t.Errorf("payload should omit zero-value optional key %q, got %v", key, got[key])
		}
	}

	mustPresent := []string{"id", "alertType", "severity", "title", "message", "status", "createdDate", "updatedDate"}
	for _, key := range mustPresent {
		if _, ok := got[key]; !ok {
			t.Errorf("payload missing required key %q", key)
		}
	}
}
